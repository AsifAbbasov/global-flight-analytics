package postgres

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const activeAircraftMetricTestDatabaseURL = "TEST_DATABASE_URL"

var activeAircraftMetricSchemaCounter uint64

type activeAircraftMetricFixture struct {
	pool       *pgxpool.Pool
	repository *MetricsRepository
}

func TestMetricsRepositoryCountsDistinctActiveAircraftWithinWindow(
	t *testing.T,
) {
	fixture := newActiveAircraftMetricFixture(
		t,
	)
	now := activeAircraftMetricFixedTime()

	insertActiveAircraftState(
		t,
		fixture.pool,
		"ABC123",
		40.40,
		49.80,
		now.Add(-1*time.Minute),
		"airplanes.live",
	)
	insertActiveAircraftState(
		t,
		fixture.pool,
		"ABC123",
		40.41,
		49.81,
		now.Add(-2*time.Minute),
		"airplanes.live",
	)
	insertActiveAircraftState(
		t,
		fixture.pool,
		"DEF456",
		41.00,
		48.90,
		now.Add(-5*time.Minute),
		"opensky",
	)
	insertActiveAircraftState(
		t,
		fixture.pool,
		"STALE1",
		40.00,
		49.00,
		now.Add(-30*time.Minute),
		"airplanes.live",
	)

	summary, err := fixture.repository.CountActiveAircraft(
		context.Background(),
		metrics.ActiveAircraftQuery{
			ObservedFrom: now.Add(-15 * time.Minute),
			ObservedTo:   now,
		},
	)
	if err != nil {
		t.Fatalf(
			"count active aircraft: %v",
			err,
		)
	}

	if summary.Count != 2 {
		t.Fatalf(
			"expected 2 active aircraft, got %d",
			summary.Count,
		)
	}

	if !summary.HasObservations {
		t.Fatal(
			"expected observation summary to have observations",
		)
	}

	expectedSources := []string{
		"airplanes.live",
		"opensky",
	}
	if !reflect.DeepEqual(
		summary.SourceNames,
		expectedSources,
	) {
		t.Fatalf(
			"expected sources %v, got %v",
			expectedSources,
			summary.SourceNames,
		)
	}

	expectedLatest := now.Add(-1 * time.Minute)
	if !summary.LatestObservedAt.Equal(
		expectedLatest,
	) {
		t.Fatalf(
			"expected latest observation %s, got %s",
			expectedLatest,
			summary.LatestObservedAt,
		)
	}
}

func TestMetricsRepositoryAppliesActiveAircraftBounds(
	t *testing.T,
) {
	fixture := newActiveAircraftMetricFixture(
		t,
	)
	now := activeAircraftMetricFixedTime()

	insertActiveAircraftState(
		t,
		fixture.pool,
		"INSIDE1",
		40.40,
		49.80,
		now.Add(-1*time.Minute),
		"airplanes.live",
	)
	insertActiveAircraftState(
		t,
		fixture.pool,
		"OUTSIDE1",
		55.00,
		49.80,
		now.Add(-1*time.Minute),
		"airplanes.live",
	)

	summary, err := fixture.repository.CountActiveAircraft(
		context.Background(),
		metrics.ActiveAircraftQuery{
			ObservedFrom: now.Add(-15 * time.Minute),
			ObservedTo:   now,
			UseBounds:    true,
			Bounds: metrics.Bounds{
				MinLatitude:  38,
				MaxLatitude:  44,
				MinLongitude: 38,
				MaxLongitude: 51,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"count bounded active aircraft: %v",
			err,
		)
	}

	if summary.Count != 1 {
		t.Fatalf(
			"expected 1 bounded active aircraft, got %d",
			summary.Count,
		)
	}
}

func TestMetricsRepositoryReturnsEmptyActiveAircraftSummaryWhenNoRecentData(
	t *testing.T,
) {
	fixture := newActiveAircraftMetricFixture(
		t,
	)
	now := activeAircraftMetricFixedTime()

	insertActiveAircraftState(
		t,
		fixture.pool,
		"STALE1",
		40.40,
		49.80,
		now.Add(-2*time.Hour),
		"airplanes.live",
	)

	summary, err := fixture.repository.CountActiveAircraft(
		context.Background(),
		metrics.ActiveAircraftQuery{
			ObservedFrom: now.Add(-15 * time.Minute),
			ObservedTo:   now,
		},
	)
	if err != nil {
		t.Fatalf(
			"count empty active aircraft: %v",
			err,
		)
	}

	if summary.Count != 0 {
		t.Fatalf(
			"expected 0 active aircraft, got %d",
			summary.Count,
		)
	}

	if summary.HasObservations {
		t.Fatal(
			"expected no observations",
		)
	}
}

func newActiveAircraftMetricFixture(
	t *testing.T,
) *activeAircraftMetricFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(
		os.Getenv(activeAircraftMetricTestDatabaseURL),
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			activeAircraftMetricTestDatabaseURL,
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	bootstrap, err := pgx.Connect(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"connect to PostgreSQL test database: %v",
			err,
		)
	}

	schemaName := fmt.Sprintf(
		"active_aircraft_metric_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(
			&activeAircraftMetricSchemaCounter,
			1,
		),
	)
	quotedSchema := pgx.Identifier{
		schemaName,
	}.Sanitize()

	if _, err := bootstrap.Exec(
		ctx,
		"CREATE SCHEMA "+quotedSchema,
	); err != nil {
		_ = bootstrap.Close(
			ctx,
		)
		t.Fatalf(
			"create PostgreSQL test schema: %v",
			err,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(
		databaseURL,
	)
	if err != nil {
		_ = bootstrap.Close(
			ctx,
		)
		t.Fatalf(
			"parse PostgreSQL test pool config: %v",
			err,
		)
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(
			map[string]string,
		)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName

	pool, err := pgxpool.NewWithConfig(
		ctx,
		poolConfig,
	)
	if err != nil {
		_ = bootstrap.Close(
			ctx,
		)
		t.Fatalf(
			"create PostgreSQL test pool: %v",
			err,
		)
	}
	if err := pool.Ping(
		ctx,
	); err != nil {
		pool.Close()
		_ = bootstrap.Close(
			ctx,
		)
		t.Fatalf(
			"ping PostgreSQL test pool: %v",
			err,
		)
	}

	t.Cleanup(func() {
		pool.Close()

		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
		defer cleanupCancel()

		if _, err := bootstrap.Exec(
			cleanupCtx,
			"DROP SCHEMA IF EXISTS "+quotedSchema+" CASCADE",
		); err != nil {
			t.Errorf(
				"drop PostgreSQL test schema: %v",
				err,
			)
		}
		if err := bootstrap.Close(
			cleanupCtx,
		); err != nil {
			t.Errorf(
				"close PostgreSQL bootstrap connection: %v",
				err,
			)
		}
	})

	createActiveAircraftMetricSchema(
		t,
		pool,
	)

	return &activeAircraftMetricFixture{
		pool:       pool,
		repository: NewMetricsRepository(pool),
	}
}

func createActiveAircraftMetricSchema(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	_, err := pool.Exec(
		context.Background(),
		`
			CREATE TABLE flight_states (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				icao24 varchar(10) NOT NULL,
				latitude numeric,
				longitude numeric,
				observed_at timestamptz NOT NULL,
				source_name text NOT NULL
			);
		`,
	)
	if err != nil {
		t.Fatalf(
			"create active aircraft metric schema: %v",
			err,
		)
	}
}

func insertActiveAircraftState(
	t *testing.T,
	pool *pgxpool.Pool,
	icao24 string,
	latitude float64,
	longitude float64,
	observedAt time.Time,
	sourceName string,
) {
	t.Helper()

	_, err := pool.Exec(
		context.Background(),
		`
			INSERT INTO flight_states (
				icao24,
				latitude,
				longitude,
				observed_at,
				source_name
			) VALUES (
				$1,
				$2,
				$3,
				$4,
				$5
			)
		`,
		icao24,
		latitude,
		longitude,
		observedAt,
		sourceName,
	)
	if err != nil {
		t.Fatalf(
			"insert active aircraft state: %v",
			err,
		)
	}
}

func activeAircraftMetricFixedTime() time.Time {
	return time.Date(
		2026,
		time.July,
		10,
		20,
		15,
		0,
		0,
		time.UTC,
	)
}
