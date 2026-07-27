package historicalread

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresHistoricalReadIntegration(
	t *testing.T,
) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	defer pool.Close()

	start := time.Now().UTC().Truncate(time.Second).Add(-2 * time.Hour)
	end := start.Add(time.Hour)
	asOf := time.Now().UTC().Add(time.Minute)

	boundaryFlightID := uuid.New()
	overlapFlightID := uuid.New()
	trajectoryID := uuid.New()
	oldRouteID := integrationRouteID("old-" + trajectoryID.String())
	newRouteID := integrationRouteID("new-" + trajectoryID.String())

	defer cleanupHistoricalReadIntegration(
		context.Background(),
		pool,
		[]string{oldRouteID, newRouteID},
		trajectoryID,
		[]uuid.UUID{boundaryFlightID, overlapFlightID},
	)

	_, err = pool.Exec(
		ctx,
		`INSERT INTO flights (
			id, first_seen_at, last_seen_at, status, created_at, updated_at
		) VALUES
			($1, $2, $3, 'completed', $4, $4),
			($5, $6, $7, 'completed', $4, $4)`,
		boundaryFlightID,
		start.Add(-time.Hour),
		start,
		asOf.Add(-time.Minute),
		overlapFlightID,
		start.Add(-30*time.Minute),
		start.Add(30*time.Minute),
	)
	if err != nil {
		t.Fatalf("insert flights: %v", err)
	}

	_, err = pool.Exec(
		ctx,
		`INSERT INTO flight_trajectories (
			id, flight_id, icao24, start_time, end_time,
			duration_seconds, segment_count, point_count,
			coverage_gap_count, quality_score, source_name,
			created_at, updated_at
		) VALUES (
			$1, $2, 'abc123', $3, $4,
			1800, 0, 0, 0, 0.987654321098765,
			'integration', $5, $5
		)`,
		trajectoryID,
		overlapFlightID,
		start.Add(10*time.Minute),
		start.Add(40*time.Minute),
		asOf.Add(-time.Minute),
	)
	if err != nil {
		t.Fatalf("insert trajectory: %v", err)
	}

	oldResult := integrationRouteResult(
		trajectoryID.String(),
		start.Add(10*time.Minute),
		start.Add(40*time.Minute),
		asOf.Add(-40*time.Second),
		"OLD1",
	)
	newResult := integrationRouteResult(
		trajectoryID.String(),
		start.Add(10*time.Minute),
		start.Add(40*time.Minute),
		asOf.Add(-20*time.Second),
		"NEW1",
	)
	oldPayload, _ := json.Marshal(oldResult)
	newPayload, _ := json.Marshal(newResult)

	_, err = pool.Exec(
		ctx,
		`INSERT INTO flight_route_results (
			id, trajectory_id, schema_version, as_of_time,
			as_of_time_unix_nano, input_fingerprint, route_status,
			confidence_level, validation_warning_count, route_json,
			stored_at, stored_at_unix_nano
		) VALUES
			($1, $2, 'route-intelligence-v1', $3, $4, $5, 'complete', 'high', 0, $6, $3, $4),
			($7, $2, 'route-intelligence-v1', $8, $9, $10, 'complete', 'high', 0, $11, $8, $9)`,
		oldRouteID,
		trajectoryID,
		oldResult.Window.AsOfTime,
		oldResult.Window.AsOfTime.UnixNano(),
		integrationFingerprint("old"),
		oldPayload,
		newRouteID,
		newResult.Window.AsOfTime,
		newResult.Window.AsOfTime.UnixNano(),
		integrationFingerprint("new"),
		newPayload,
	)
	if err != nil {
		t.Fatalf("insert route versions: %v", err)
	}

	repository, err := NewPostgres(PostgresConfig{Pool: pool})
	if err != nil {
		t.Fatalf("compose repository: %v", err)
	}

	snapshot, err := repository.Read(
		ctx,
		Query{
			Window: historicalcontract.TimeWindow{
				StartTime: start,
				EndTime:   end,
				AsOfTime:  asOf,
			},
			Limit: 10,
		},
	)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}

	if snapshot.IsolationLevel != SnapshotIsolationRepeatableRead ||
		snapshot.FlightMatchedCount != 1 ||
		len(snapshot.Flights) != 1 ||
		snapshot.Flights[0].ID != overlapFlightID.String() {
		t.Fatalf("half-open flight selection failed: %#v", snapshot)
	}
	if snapshot.TrajectoryMatchedCount != 1 || len(snapshot.Trajectories) != 1 {
		t.Fatalf("trajectory selection failed: %#v", snapshot)
	}
	if snapshot.RouteMatchedCount != 1 || len(snapshot.Routes) != 1 {
		t.Fatalf("latest route deduplication failed: %#v", snapshot)
	}
	result, valid := snapshot.Routes[0].ResultAt(asOf)
	if !valid || result.Callsign != "NEW1" {
		t.Fatalf("latest route version was not selected: %#v %t", result, valid)
	}
	if snapshot.Trajectories[0].QualityScore != 0.987654321099 {
		t.Fatalf("quality rounding policy not applied: %.15f", snapshot.Trajectories[0].QualityScore)
	}

	byteLimited, err := repository.Read(
		ctx,
		Query{
			Window: historicalcontract.TimeWindow{
				StartTime: start,
				EndTime:   end,
				AsOfTime:  asOf,
			},
			Limit:                 10,
			RoutePayloadByteLimit: 1,
		},
	)
	if err != nil {
		t.Fatalf("read byte-limited snapshot: %v", err)
	}
	if !byteLimited.RouteByteLimitReached ||
		byteLimited.RouteMatchedCount != 1 ||
		len(byteLimited.Routes) != 0 {
		t.Fatalf("route byte budget was not enforced: %#v", byteLimited)
	}

	for _, indexName := range []string{
		"historical_read_flight_versions_event_idx",
		"historical_read_trajectory_versions_event_idx",
		"flight_states_historical_read_idx",
		"flight_route_results_historical_read_idx",
	} {
		var exists bool
		if err := pool.QueryRow(
			ctx,
			`SELECT EXISTS (
				SELECT 1
				FROM pg_indexes
				WHERE schemaname = current_schema()
				  AND indexname = $1
			)`,
			indexName,
		).Scan(&exists); err != nil {
			t.Fatalf("inspect index %s: %v", indexName, err)
		}
		if !exists {
			t.Fatalf("required historical read index %s is absent", indexName)
		}
	}
}

func TestPostgresHistoricalReadRepeatableReadConsistency(
	t *testing.T,
) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	defer pool.Close()

	start := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	end := start.Add(30 * time.Minute)
	asOf := time.Now().UTC().Add(time.Minute)
	flightID := uuid.New()
	trajectoryID := uuid.New()

	defer cleanupHistoricalReadIntegration(
		context.Background(),
		pool,
		nil,
		trajectoryID,
		[]uuid.UUID{flightID},
	)

	initialUpdatedAt := asOf.Add(-30 * time.Second)
	_, err = pool.Exec(
		ctx,
		`INSERT INTO flights (
			id, first_seen_at, last_seen_at, status, created_at, updated_at
		) VALUES ($1, $2, $3, 'completed', $4, $4)`,
		flightID,
		start,
		end,
		initialUpdatedAt,
	)
	if err != nil {
		t.Fatalf("insert consistency flight: %v", err)
	}

	_, err = pool.Exec(
		ctx,
		`INSERT INTO flight_trajectories (
			id, flight_id, icao24, start_time, end_time,
			duration_seconds, segment_count, point_count,
			coverage_gap_count, quality_score, source_name,
			created_at, updated_at
		) VALUES (
			$1, $2, 'abc123', $3, $4,
			1800, 0, 0, 0, 0.25, 'integration', $5, $5
		)`,
		trajectoryID,
		flightID,
		start,
		end,
		initialUpdatedAt,
	)
	if err != nil {
		t.Fatalf("insert consistency trajectory: %v", err)
	}

	beginner := &integrationSnapshotBeginner{
		pool: pool,
		beforeTrajectoryRead: func(hookContext context.Context) error {
			_, hookErr := pool.Exec(
				hookContext,
				`UPDATE flight_trajectories
				 SET quality_score = 0.75, updated_at = $2
				 WHERE id = $1`,
				trajectoryID,
				asOf.Add(-10*time.Second),
			)
			return hookErr
		},
	}
	repository := newManagedPostgresRepository(beginner)
	snapshot, err := repository.Read(
		ctx,
		Query{
			Window: historicalcontract.TimeWindow{
				StartTime: start,
				EndTime:   end,
				AsOfTime:  asOf,
			},
			Limit: 10,
		},
	)
	if err != nil {
		t.Fatalf("read repeatable snapshot: %v", err)
	}
	if beginner.hookCount != 1 {
		t.Fatalf("concurrent mutation hook count = %d, want 1", beginner.hookCount)
	}
	if len(snapshot.Trajectories) != 1 ||
		snapshot.Trajectories[0].QualityScore != 0.25 {
		t.Fatalf("snapshot mixed pre- and post-mutation evidence: %#v", snapshot)
	}

	var currentQuality float64
	if err := pool.QueryRow(
		ctx,
		`SELECT quality_score::double precision
		 FROM flight_trajectories
		 WHERE id = $1`,
		trajectoryID,
	).Scan(&currentQuality); err != nil {
		t.Fatalf("read current trajectory: %v", err)
	}
	if currentQuality != 0.75 {
		t.Fatalf("concurrent mutation did not commit: %f", currentQuality)
	}
}

type integrationSnapshotBeginner struct {
	pool                 *pgxpool.Pool
	beforeTrajectoryRead func(context.Context) error
	hookCount            int
}

func (beginner *integrationSnapshotBeginner) BeginSnapshot(
	ctx context.Context,
) (managedSnapshot, error) {
	transaction, err := beginner.pool.BeginTx(
		ctx,
		pgx.TxOptions{
			IsoLevel:   pgx.RepeatableRead,
			AccessMode: pgx.ReadOnly,
		},
	)
	if err != nil {
		return nil, err
	}
	return &integrationManagedSnapshot{
		transaction: transaction,
		beginner:    beginner,
	}, nil
}

type integrationManagedSnapshot struct {
	transaction pgx.Tx
	beginner    *integrationSnapshotBeginner
	queryCount  int
}

func (snapshot *integrationManagedSnapshot) Query(
	ctx context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	snapshot.queryCount++
	if snapshot.queryCount == 3 &&
		snapshot.beginner.beforeTrajectoryRead != nil {
		snapshot.beginner.hookCount++
		if err := snapshot.beginner.beforeTrajectoryRead(ctx); err != nil {
			return nil, err
		}
	}
	rows, err := snapshot.transaction.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (snapshot *integrationManagedSnapshot) Commit(
	ctx context.Context,
) error {
	return snapshot.transaction.Commit(ctx)
}

func (snapshot *integrationManagedSnapshot) Rollback(
	ctx context.Context,
) error {
	return snapshot.transaction.Rollback(ctx)
}

func integrationRouteResult(
	trajectoryID string,
	start time.Time,
	end time.Time,
	asOf time.Time,
	callsign string,
) routecontract.Result {
	return routecontract.Result{
		SchemaVersion: routecontract.SchemaVersionV1,
		Status:        routecontract.RouteStatusComplete,
		TrajectoryID:  trajectoryID,
		Callsign:      callsign,
		Window: routecontract.RouteWindow{
			StartTime: start,
			EndTime:   end,
			AsOfTime:  asOf,
		},
		Confidence: routecontract.Confidence{
			Score: 1,
			Level: routecontract.ConfidenceLevelHigh,
		},
		GeneratedAt: asOf,
	}
}

func integrationRouteID(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return "route-record-" + hex.EncodeToString(sum[:])
}

func integrationFingerprint(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func cleanupHistoricalReadIntegration(
	ctx context.Context,
	pool *pgxpool.Pool,
	routeIDs []string,
	trajectoryID uuid.UUID,
	flightIDs []uuid.UUID,
) {
	_, _ = pool.Exec(ctx, `DELETE FROM flight_route_results WHERE id = ANY($1)`, routeIDs)
	_, _ = pool.Exec(ctx, `DELETE FROM flight_trajectories WHERE id = $1`, trajectoryID)
	_, _ = pool.Exec(ctx, `DELETE FROM flights WHERE id = ANY($1)`, flightIDs)
	_, _ = pool.Exec(ctx, `DELETE FROM historical_read_trajectory_versions WHERE trajectory_id = $1`, trajectoryID)
	_, _ = pool.Exec(ctx, `DELETE FROM historical_read_flight_versions WHERE flight_id = ANY($1)`, flightIDs)
}
