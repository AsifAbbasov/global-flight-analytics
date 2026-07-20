package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const trafficAltitudeTestDatabaseURL = "TEST_DATABASE_URL"

var trafficAltitudeSchemaCounter uint64

func TestTrafficRepositoryPreservesAltitudeSemantics(
	t *testing.T,
) {
	fixture := newTrafficAltitudeFixture(t)
	ctx := context.Background()
	runID := "11111111-1111-1111-1111-111111111111"
	finishedAt := time.Date(
		2026,
		time.July,
		20,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	mustExecTrafficAltitudeSQL(
		t,
		fixture.pool,
		`
			INSERT INTO ingestion_runs (
				id,
				finished_at,
				status,
				created_at
			)
			VALUES ($1, $2, 'success', $2)
		`,
		runID,
		finishedAt,
	)

	insertTrafficAltitudeState(
		t,
		fixture.pool,
		runID,
		"AAA001",
		40.0,
		49.0,
		intPointer(0),
		"observed",
		intPointer(1200),
		"observed",
		false,
		finishedAt,
	)
	insertTrafficAltitudeState(
		t,
		fixture.pool,
		runID,
		"AAA002",
		41.0,
		50.0,
		nil,
		"unavailable",
		intPointer(2400),
		"observed",
		false,
		finishedAt.Add(time.Second),
	)
	insertTrafficAltitudeState(
		t,
		fixture.pool,
		runID,
		"AAA003",
		42.0,
		51.0,
		nil,
		"unknown",
		nil,
		"unavailable",
		false,
		finishedAt.Add(2*time.Second),
	)
	insertTrafficAltitudeState(
		t,
		fixture.pool,
		runID,
		"AAA004",
		43.0,
		52.0,
		nil,
		"unavailable",
		nil,
		"unavailable",
		true,
		finishedAt.Add(3*time.Second),
	)

	items, err := fixture.repository.GetCurrent(ctx)
	if err != nil {
		t.Fatalf("get current traffic: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf(
			"current traffic count = %d, want 4",
			len(items),
		)
	}

	assertTrafficAltitudeItem(
		t,
		items[0],
		"AAA001",
		true,
		0,
		flightstate.AltitudeStatusObserved,
		traffic.AltitudeSourceGeometric,
	)
	assertTrafficAltitudeItem(
		t,
		items[1],
		"AAA002",
		true,
		2400,
		flightstate.AltitudeStatusObserved,
		traffic.AltitudeSourceBarometric,
	)
	assertTrafficAltitudeItem(
		t,
		items[2],
		"AAA003",
		false,
		0,
		flightstate.AltitudeStatusUnknown,
		traffic.AltitudeSourceNone,
	)
	assertTrafficAltitudeItem(
		t,
		items[3],
		"AAA004",
		true,
		0,
		flightstate.AltitudeStatusGround,
		traffic.AltitudeSourceGround,
	)

	bounded, err := fixture.repository.GetCurrentByBounds(
		ctx,
		traffic.Bounds{
			MinLatitude:  39.5,
			MaxLatitude:  41.5,
			MinLongitude: 48.5,
			MaxLongitude: 50.5,
		},
	)
	if err != nil {
		t.Fatalf("get bounded current traffic: %v", err)
	}
	if len(bounded) != 2 {
		t.Fatalf(
			"bounded traffic count = %d, want 2",
			len(bounded),
		)
	}

	assertTrafficAltitudeItem(
		t,
		bounded[0],
		"AAA001",
		true,
		0,
		flightstate.AltitudeStatusObserved,
		traffic.AltitudeSourceGeometric,
	)
	assertTrafficAltitudeItem(
		t,
		bounded[1],
		"AAA002",
		true,
		2400,
		flightstate.AltitudeStatusObserved,
		traffic.AltitudeSourceBarometric,
	)
}

type trafficAltitudeFixture struct {
	pool       *pgxpool.Pool
	repository *TrafficRepository
}

func newTrafficAltitudeFixture(
	t *testing.T,
) *trafficAltitudeFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(
		os.Getenv(trafficAltitudeTestDatabaseURL),
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			trafficAltitudeTestDatabaseURL,
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf(
			"connect to PostgreSQL test database: %v",
			err,
		)
	}

	schemaName := fmt.Sprintf(
		"traffic_altitude_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(
			&trafficAltitudeSchemaCounter,
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
		_ = bootstrap.Close(ctx)
		t.Fatalf(
			"create traffic altitude test schema: %v",
			err,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf(
			"parse traffic altitude test pool config: %v",
			err,
		)
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(
			map[string]string,
		)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf(
			"create traffic altitude test pool: %v",
			err,
		)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatalf(
			"ping traffic altitude test pool: %v",
			err,
		)
	}

	createTrafficAltitudeSchema(t, pool)

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
				"drop traffic altitude test schema: %v",
				err,
			)
		}
		if err := bootstrap.Close(cleanupCtx); err != nil {
			t.Errorf(
				"close traffic altitude bootstrap connection: %v",
				err,
			)
		}
	})

	return &trafficAltitudeFixture{
		pool:       pool,
		repository: NewTrafficRepository(pool),
	}
}

func createTrafficAltitudeSchema(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	mustExecTrafficAltitudeSQL(
		t,
		pool,
		`
			CREATE TABLE ingestion_runs (
				id uuid PRIMARY KEY,
				finished_at timestamptz,
				status text NOT NULL,
				created_at timestamptz NOT NULL
			);

			CREATE TABLE airlines (
				id uuid PRIMARY KEY,
				name text
			);

			CREATE TABLE aircraft_models (
				id uuid PRIMARY KEY,
				model text
			);

			CREATE TABLE aircraft (
				id uuid PRIMARY KEY,
				model_id uuid,
				airline_id uuid
			);

			CREATE TABLE flight_states (
				aircraft_id uuid,
				ingestion_run_id uuid NOT NULL,
				icao24 text NOT NULL,
				callsign text,
				latitude double precision,
				longitude double precision,
				geometric_altitude_m integer,
				geometric_altitude_status text NOT NULL,
				barometric_altitude_m integer,
				barometric_altitude_status text NOT NULL,
				velocity_mps double precision,
				heading_degrees double precision,
				on_ground boolean,
				observed_at timestamptz NOT NULL,
				origin_country text
			)
		`,
	)
}

func insertTrafficAltitudeState(
	t *testing.T,
	pool *pgxpool.Pool,
	runID string,
	icao24 string,
	latitude float64,
	longitude float64,
	geometricAltitude *int,
	geometricStatus string,
	barometricAltitude *int,
	barometricStatus string,
	onGround bool,
	observedAt time.Time,
) {
	t.Helper()

	mustExecTrafficAltitudeSQL(
		t,
		pool,
		`
			INSERT INTO flight_states (
				ingestion_run_id,
				icao24,
				callsign,
				latitude,
				longitude,
				geometric_altitude_m,
				geometric_altitude_status,
				barometric_altitude_m,
				barometric_altitude_status,
				velocity_mps,
				heading_degrees,
				on_ground,
				observed_at,
				origin_country
			)
			VALUES (
				$1,
				$2,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				200,
				90,
				$9,
				$10,
				'Azerbaijan'
			)
		`,
		runID,
		icao24,
		latitude,
		longitude,
		geometricAltitude,
		geometricStatus,
		barometricAltitude,
		barometricStatus,
		onGround,
		observedAt,
	)
}

func assertTrafficAltitudeItem(
	t *testing.T,
	item traffic.CurrentTrafficItem,
	expectedICAO24 string,
	expectedValuePresent bool,
	expectedValue float64,
	expectedStatus flightstate.AltitudeStatus,
	expectedSource traffic.AltitudeSource,
) {
	t.Helper()

	if item.ICAO24 != expectedICAO24 {
		t.Fatalf(
			"icao24 = %s, want %s",
			item.ICAO24,
			expectedICAO24,
		)
	}
	if (item.AltitudeM != nil) != expectedValuePresent {
		t.Fatalf(
			"%s altitude value presence = %v, want %v",
			item.ICAO24,
			item.AltitudeM != nil,
			expectedValuePresent,
		)
	}
	if expectedValuePresent &&
		*item.AltitudeM != expectedValue {
		t.Fatalf(
			"%s altitude value = %v, want %v",
			item.ICAO24,
			*item.AltitudeM,
			expectedValue,
		)
	}
	if item.AltitudeStatus != expectedStatus {
		t.Fatalf(
			"%s altitude status = %q, want %q",
			item.ICAO24,
			item.AltitudeStatus,
			expectedStatus,
		)
	}
	if item.AltitudeSource != expectedSource {
		t.Fatalf(
			"%s altitude source = %q, want %q",
			item.ICAO24,
			item.AltitudeSource,
			expectedSource,
		)
	}
}

func mustExecTrafficAltitudeSQL(
	t *testing.T,
	pool *pgxpool.Pool,
	query string,
	arguments ...any,
) {
	t.Helper()

	if _, err := pool.Exec(
		context.Background(),
		query,
		arguments...,
	); err != nil {
		t.Fatalf(
			"execute traffic altitude SQL: %v",
			err,
		)
	}
}

func intPointer(value int) *int {
	return &value
}
