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

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database/migrator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const stage14CorrectnessDatabaseURL = "TEST_DATABASE_URL"

var stage14CorrectnessSchemaCounter uint64

func TestStage14CorrectnessConstraintsOnProductionCatalog(t *testing.T) {
	pool := newStage14CorrectnessFixture(t)

	t.Run("processed counts cannot exceed received records", func(t *testing.T) {
		_, err := pool.Exec(
			context.Background(),
			`
				INSERT INTO ingestion_runs (
					id,
					source_name,
					started_at,
					status,
					records_received,
					records_inserted,
					records_updated
				)
				VALUES (
					'10000000-0000-0000-0000-000000000001',
					'test',
					now(),
					'running',
					1,
					1,
					1
				)
			`,
		)
		assertStage14CorrectnessPostgresCode(t, err, "23514")
	})

	t.Run("failed and partial runs require an error message", func(t *testing.T) {
		for _, status := range []string{"failed", "partial"} {
			_, err := pool.Exec(
				context.Background(),
				`
					INSERT INTO ingestion_runs (
						id,
						source_name,
						started_at,
						finished_at,
						status
					)
					VALUES (
						gen_random_uuid(),
						'test',
						now(),
						now(),
						$1
					)
				`,
				status,
			)
			assertStage14CorrectnessPostgresCode(t, err, "23514")
		}
	})

	t.Run("success runs reject error messages", func(t *testing.T) {
		_, err := pool.Exec(
			context.Background(),
			`
				INSERT INTO ingestion_runs (
					id,
					source_name,
					started_at,
					finished_at,
					status,
					error_message
				)
				VALUES (
					'10000000-0000-0000-0000-000000000002',
					'test',
					now(),
					now(),
					'success',
					'unexpected error'
				)
			`,
		)
		assertStage14CorrectnessPostgresCode(t, err, "23514")
	})

	t.Run("route timestamp mirrors reject drift", func(t *testing.T) {
		trajectoryID := "20000000-0000-0000-0000-000000000001"
		instant := time.Date(2026, 7, 21, 1, 2, 3, 123456789, time.UTC)

		_, err := pool.Exec(
			context.Background(),
			`
				INSERT INTO flight_trajectories (
					id,
					icao24,
					start_time,
					end_time,
					duration_seconds,
					segment_count,
					point_count,
					coverage_gap_count,
					quality_score,
					source_name
				)
				VALUES ($1, 'abc123', $2, $2, 0, 0, 0, 0, 1, 'test')
			`,
			trajectoryID,
			instant,
		)
		if err != nil {
			t.Fatalf("insert trajectory fixture: %v", err)
		}

		_, err = pool.Exec(
			context.Background(),
			`
				INSERT INTO flight_route_results (
					id,
					trajectory_id,
					schema_version,
					as_of_time,
					as_of_time_unix_nano,
					input_fingerprint,
					route_status,
					confidence_level,
					validation_warning_count,
					route_json,
					stored_at,
					stored_at_unix_nano
				)
				VALUES (
					$1,
					$2,
					'route-intelligence-v1',
					$3,
					$4,
					$5,
					'unavailable',
					'none',
					0,
					'{}'::jsonb,
					$6,
					$7
				)
			`,
			"route-record-"+strings.Repeat("a", 64),
			trajectoryID,
			instant.Add(2*time.Microsecond),
			instant.UnixNano(),
			"sha256:"+strings.Repeat("b", 64),
			instant,
			instant.UnixNano(),
		)
		assertStage14CorrectnessPostgresCode(t, err, "23514")
	})

	t.Run("historical timestamp mirrors reject drift", func(t *testing.T) {
		windowStart := time.Date(2026, 7, 21, 0, 0, 0, 123456789, time.UTC)
		windowEnd := windowStart.Add(time.Hour)
		asOfTime := windowEnd
		storedAt := asOfTime.Add(time.Minute)

		_, err := pool.Exec(
			context.Background(),
			`
				INSERT INTO historical_aggregate_results (
					id,
					schema_version,
					metric_name,
					scope_type,
					scope_key,
					granularity,
					window_start,
					window_start_unix_nano,
					window_end,
					window_end_unix_nano,
					as_of_time,
					as_of_time_unix_nano,
					input_fingerprint,
					series_status,
					confidence_level,
					result_json,
					stored_at,
					stored_at_unix_nano
				)
				VALUES (
					$1,
					'historical-intelligence-v1',
					'flight_count',
					'global',
					'global',
					'hour',
					$2,
					$3,
					$4,
					$5,
					$6,
					$7,
					$8,
					'unavailable',
					'none',
					'{}'::jsonb,
					$9,
					$10
				)
			`,
			"historical-aggregate-record-"+strings.Repeat("c", 64),
			windowStart.Add(2*time.Microsecond),
			windowStart.UnixNano(),
			windowEnd,
			windowEnd.UnixNano(),
			asOfTime,
			asOfTime.UnixNano(),
			"sha256:"+strings.Repeat("d", 64),
			storedAt,
			storedAt.UnixNano(),
		)
		assertStage14CorrectnessPostgresCode(t, err, "23514")
	})
}

func newStage14CorrectnessFixture(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := strings.TrimSpace(os.Getenv(stage14CorrectnessDatabaseURL))
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			stage14CorrectnessDatabaseURL,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL test database: %v", err)
	}

	schemaName := fmt.Sprintf(
		"stage14_correctness_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&stage14CorrectnessSchemaCounter, 1),
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

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatal("resolve correctness integration test path")
	}
	migrationDirectory := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations",
		),
	)
	runner, err := migrator.NewRunner(pool, migrationDirectory)
	if err != nil {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatalf("create production migration runner: %v", err)
	}
	if _, err := runner.ApplyPending(ctx); err != nil {
		pool.Close()
		_ = bootstrap.Close(ctx)
		t.Fatalf("apply production migration catalog: %v", err)
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
			t.Errorf("drop PostgreSQL correctness schema: %v", err)
		}
		if err := bootstrap.Close(cleanupCtx); err != nil {
			t.Errorf("close PostgreSQL correctness bootstrap connection: %v", err)
		}
	})

	return pool
}

func assertStage14CorrectnessPostgresCode(
	t *testing.T,
	err error,
	expectedCode string,
) {
	t.Helper()

	if err == nil {
		t.Fatalf(
			"expected PostgreSQL error code %s, got nil",
			expectedCode,
		)
	}

	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("expected PostgreSQL error, got %T: %v", err, err)
	}
	if postgresError.Code != expectedCode {
		t.Fatalf(
			"expected PostgreSQL error code %s, got %s",
			expectedCode,
			postgresError.Code,
		)
	}
}
