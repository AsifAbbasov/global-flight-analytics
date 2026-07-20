package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const trajectoryRelationalTestDatabaseURL = "TEST_DATABASE_URL"

var trajectoryRelationalSchemaCounter uint64

type trajectoryRelationalFixture struct {
	pool *pgxpool.Pool
}

func TestTrajectoryRelationalIntegrityAcceptsCanonicalAggregate(
	t *testing.T,
) {
	fixture := newTrajectoryRelationalFixture(t)
	createTrajectoryRelationalSchema(t, fixture.pool)
	applyTrajectoryRelationalMigration(t, fixture.pool)

	ctx := context.Background()
	trajectoryID := "11111111-1111-1111-1111-111111111111"
	firstSegmentID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	secondSegmentID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	tx, err := fixture.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin canonical trajectory transaction: %v", err)
	}

	mustExecTrajectoryRelationalTx(
		t,
		tx,
		`INSERT INTO flight_trajectories (
			id, icao24, segment_count, point_count,
			coverage_gap_count, source_name
		) VALUES ($1, 'ABC123', 2, 4, 1, 'test')`,
		trajectoryID,
	)
	mustInsertTrajectorySegment(t, tx, firstSegmentID, trajectoryID, 1, "ABC123", 2)
	mustInsertTrajectorySegment(t, tx, secondSegmentID, trajectoryID, 2, "ABC123", 2)
	mustExecTrajectoryRelationalTx(
		t,
		tx,
		`INSERT INTO coverage_gaps (
			trajectory_id, previous_segment_id, next_segment_id,
			icao24, reason
		) VALUES ($1, $2, $3, 'ABC123', 'time_gap')`,
		trajectoryID,
		firstSegmentID,
		secondSegmentID,
	)

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit canonical trajectory aggregate: %v", err)
	}
}

func TestTrajectoryRelationalIntegrityRejectsNullParentAndDuplicateSequence(
	t *testing.T,
) {
	fixture := newTrajectoryRelationalFixture(t)
	createTrajectoryRelationalSchema(t, fixture.pool)
	applyTrajectoryRelationalMigration(t, fixture.pool)

	_, err := fixture.pool.Exec(
		context.Background(),
		`INSERT INTO trajectory_segments (
			trajectory_id, icao24, sequence_number, point_count, source_name
		) VALUES (NULL, 'ABC123', 1, 1, 'test')`,
	)
	assertTrajectoryRelationalPostgresCode(t, err, "23502")

	ctx := context.Background()
	trajectoryID := "22222222-2222-2222-2222-222222222222"
	tx, err := fixture.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin duplicate sequence transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	mustExecTrajectoryRelationalTx(
		t,
		tx,
		`INSERT INTO flight_trajectories (
			id, icao24, segment_count, point_count,
			coverage_gap_count, source_name
		) VALUES ($1, 'ABC123', 2, 2, 0, 'test')`,
		trajectoryID,
	)
	mustInsertTrajectorySegment(
		t,
		tx,
		"cccccccc-cccc-cccc-cccc-cccccccccccc",
		trajectoryID,
		1,
		"ABC123",
		1,
	)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO trajectory_segments (
			id, trajectory_id, icao24, sequence_number, point_count, source_name
		) VALUES ($1, $2, 'ABC123', 1, 1, 'test')`,
		"dddddddd-dddd-dddd-dddd-dddddddddddd",
		trajectoryID,
	)
	assertTrajectoryRelationalPostgresCode(t, err, "23505")
}

func TestTrajectoryRelationalIntegrityRejectsCrossTrajectoryGapReference(
	t *testing.T,
) {
	fixture := newTrajectoryRelationalFixture(t)
	createTrajectoryRelationalSchema(t, fixture.pool)
	applyTrajectoryRelationalMigration(t, fixture.pool)

	ctx := context.Background()
	firstTrajectoryID := "33333333-3333-3333-3333-333333333333"
	secondTrajectoryID := "44444444-4444-4444-4444-444444444444"
	firstSegmentID := "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	secondSegmentID := "ffffffff-ffff-ffff-ffff-ffffffffffff"

	mustInsertCompleteOneSegmentTrajectory(
		t,
		fixture.pool,
		firstTrajectoryID,
		firstSegmentID,
		"ABC123",
	)
	mustInsertCompleteOneSegmentTrajectory(
		t,
		fixture.pool,
		secondTrajectoryID,
		secondSegmentID,
		"DEF456",
	)

	tx, err := fixture.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin cross-trajectory gap transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	mustExecTrajectoryRelationalTx(
		t,
		tx,
		`UPDATE flight_trajectories
		 SET coverage_gap_count = 1
		 WHERE id = $1`,
		firstTrajectoryID,
	)
	mustExecTrajectoryRelationalTx(
		t,
		tx,
		`INSERT INTO coverage_gaps (
			trajectory_id, previous_segment_id, next_segment_id,
			icao24, reason
		) VALUES ($1, $2, NULL, 'ABC123', 'time_gap')`,
		firstTrajectoryID,
		secondSegmentID,
	)

	_, err = tx.Exec(
		ctx,
		"SET CONSTRAINTS coverage_gaps_previous_segment_same_trajectory_fk IMMEDIATE",
	)
	assertTrajectoryRelationalPostgresCode(t, err, "23503")
}

func TestTrajectoryRelationalIntegrityRejectsStoredCountAndSequenceGaps(
	t *testing.T,
) {
	fixture := newTrajectoryRelationalFixture(t)
	createTrajectoryRelationalSchema(t, fixture.pool)
	applyTrajectoryRelationalMigration(t, fixture.pool)

	ctx := context.Background()
	tx, err := fixture.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin mismatched count transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	trajectoryID := "55555555-5555-5555-5555-555555555555"
	mustExecTrajectoryRelationalTx(
		t,
		tx,
		`INSERT INTO flight_trajectories (
			id, icao24, segment_count, point_count,
			coverage_gap_count, source_name
		) VALUES ($1, 'ABC123', 2, 2, 0, 'test')`,
		trajectoryID,
	)
	mustInsertTrajectorySegment(
		t,
		tx,
		"12121212-1212-1212-1212-121212121212",
		trajectoryID,
		1,
		"ABC123",
		1,
	)
	mustInsertTrajectorySegment(
		t,
		tx,
		"34343434-3434-3434-3434-343434343434",
		trajectoryID,
		3,
		"ABC123",
		1,
	)

	_, err = tx.Exec(ctx, "SET CONSTRAINTS ALL IMMEDIATE")
	assertTrajectoryRelationalPostgresCode(t, err, "23514")
}

func newTrajectoryRelationalFixture(
	t *testing.T,
) *trajectoryRelationalFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(os.Getenv(trajectoryRelationalTestDatabaseURL))
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			trajectoryRelationalTestDatabaseURL,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL test database: %v", err)
	}

	schemaName := fmt.Sprintf(
		"trajectory_relational_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&trajectoryRelationalSchemaCounter, 1),
	)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()

	if _, err := bootstrap.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create PostgreSQL test schema: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("parse PostgreSQL test pool config: %v", err)
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create PostgreSQL test pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatalf("ping PostgreSQL test pool: %v", err)
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
			t.Errorf("drop PostgreSQL test schema: %v", err)
		}
		if err := bootstrap.Close(cleanupCtx); err != nil {
			t.Errorf("close PostgreSQL bootstrap connection: %v", err)
		}
	})

	return &trajectoryRelationalFixture{pool: pool}
}

func createTrajectoryRelationalSchema(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	mustExecTrajectoryRelationalSQL(
		t,
		pool,
		`
			CREATE TABLE flight_trajectories (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				identity_key text,
				identity_basis text,
				split_reason text,
				flight_id uuid,
				aircraft_id uuid,
				icao24 text NOT NULL,
				callsign text,
				start_time timestamptz NOT NULL DEFAULT now(),
				end_time timestamptz NOT NULL DEFAULT now(),
				duration_seconds bigint NOT NULL DEFAULT 0,
				segment_count integer NOT NULL DEFAULT 0,
				point_count integer NOT NULL DEFAULT 0,
				coverage_gap_count integer NOT NULL DEFAULT 0,
				quality_score numeric NOT NULL DEFAULT 0,
				source_name text NOT NULL,
				reconciliation_task_id uuid,
				created_at timestamptz NOT NULL DEFAULT now(),
				updated_at timestamptz NOT NULL DEFAULT now()
			);

			CREATE TABLE trajectory_segments (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				trajectory_id uuid REFERENCES flight_trajectories(id) ON DELETE CASCADE,
				flight_id uuid,
				aircraft_id uuid,
				icao24 text NOT NULL,
				callsign text,
				sequence_number integer NOT NULL,
				status text NOT NULL DEFAULT 'observed',
				quality_score numeric NOT NULL DEFAULT 0,
				start_time timestamptz NOT NULL DEFAULT now(),
				end_time timestamptz NOT NULL DEFAULT now(),
				duration_seconds bigint NOT NULL DEFAULT 0,
				start_latitude numeric NOT NULL DEFAULT 0,
				start_longitude numeric NOT NULL DEFAULT 0,
				end_latitude numeric NOT NULL DEFAULT 0,
				end_longitude numeric NOT NULL DEFAULT 0,
				point_count integer NOT NULL DEFAULT 0,
				source_name text NOT NULL,
				created_at timestamptz NOT NULL DEFAULT now()
			);

			CREATE TABLE coverage_gaps (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				trajectory_id uuid REFERENCES flight_trajectories(id) ON DELETE CASCADE,
				previous_segment_id uuid REFERENCES trajectory_segments(id) ON DELETE SET NULL,
				next_segment_id uuid REFERENCES trajectory_segments(id) ON DELETE SET NULL,
				icao24 text NOT NULL,
				gap_start_time timestamptz NOT NULL DEFAULT now(),
				gap_end_time timestamptz NOT NULL DEFAULT now(),
				duration_seconds bigint NOT NULL DEFAULT 0,
				distance_km numeric NOT NULL DEFAULT 0,
				reason text NOT NULL,
				filled_by text,
				created_at timestamptz NOT NULL DEFAULT now()
			);
		`,
	)
}

func applyTrajectoryRelationalMigration(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve trajectory relational test file path")
	}

	migrationPath := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations/018_trajectory_relational_integrity.sql",
		),
	)
	migrationBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read trajectory relational migration %s: %v", migrationPath, err)
	}

	connection, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire PostgreSQL test connection: %v", err)
	}
	defer connection.Release()

	if _, err := connection.Exec(context.Background(), string(migrationBytes)); err != nil {
		_, _ = connection.Exec(context.Background(), "ROLLBACK")
		t.Fatalf("apply trajectory relational migration: %v", err)
	}
}

func mustInsertCompleteOneSegmentTrajectory(
	t *testing.T,
	pool *pgxpool.Pool,
	trajectoryID string,
	segmentID string,
	icao24 string,
) {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin one-segment trajectory transaction: %v", err)
	}

	mustExecTrajectoryRelationalTx(
		t,
		tx,
		`INSERT INTO flight_trajectories (
			id, icao24, segment_count, point_count,
			coverage_gap_count, source_name
		) VALUES ($1, $2, 1, 2, 0, 'test')`,
		trajectoryID,
		icao24,
	)
	mustInsertTrajectorySegment(t, tx, segmentID, trajectoryID, 1, icao24, 2)

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit one-segment trajectory: %v", err)
	}
}

func mustInsertTrajectorySegment(
	t *testing.T,
	tx pgx.Tx,
	segmentID string,
	trajectoryID string,
	sequenceNumber int,
	icao24 string,
	pointCount int,
) {
	t.Helper()
	mustExecTrajectoryRelationalTx(
		t,
		tx,
		`INSERT INTO trajectory_segments (
			id, trajectory_id, icao24, sequence_number,
			point_count, source_name
		) VALUES ($1, $2, $3, $4, $5, 'test')`,
		segmentID,
		trajectoryID,
		icao24,
		sequenceNumber,
		pointCount,
	)
}

func mustExecTrajectoryRelationalSQL(
	t *testing.T,
	pool *pgxpool.Pool,
	query string,
	arguments ...any,
) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), query, arguments...); err != nil {
		t.Fatalf("execute trajectory relational SQL: %v", err)
	}
}

func mustExecTrajectoryRelationalTx(
	t *testing.T,
	tx pgx.Tx,
	query string,
	arguments ...any,
) {
	t.Helper()
	if _, err := tx.Exec(context.Background(), query, arguments...); err != nil {
		t.Fatalf("execute trajectory relational transaction SQL: %v", err)
	}
}

func assertTrajectoryRelationalPostgresCode(
	t *testing.T,
	err error,
	expectedCode string,
) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected PostgreSQL error code %s, got nil", expectedCode)
	}

	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("expected PostgreSQL error, got %T: %v", err, err)
	}
	if postgresError.Code != expectedCode {
		t.Fatalf(
			"expected PostgreSQL error code %s, got %s: %v",
			expectedCode,
			postgresError.Code,
			err,
		)
	}
}
