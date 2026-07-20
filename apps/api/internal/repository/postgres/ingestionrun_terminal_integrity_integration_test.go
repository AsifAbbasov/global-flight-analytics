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

const ingestionRunTerminalTestDatabaseURL = "TEST_DATABASE_URL"

var ingestionRunTerminalSchemaCounter uint64

type ingestionRunTerminalFixture struct {
	pool       *pgxpool.Pool
	repository *IngestionRunRepository
}

func TestIngestionRunTerminalIntegrityRejectsRepeatedFinalization(
	t *testing.T,
) {
	fixture := newIngestionRunTerminalFixture(t)
	createIngestionRunTerminalSchema(t, fixture.pool)
	applyIngestionRunTerminalMigration(t, fixture.pool)

	ctx := context.Background()
	runID := "11111111-1111-1111-1111-111111111111"
	startedAt := time.Date(2026, time.July, 20, 8, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(5 * time.Minute)

	mustExecIngestionRunTerminalSQL(
		t,
		fixture.pool,
		`
			INSERT INTO ingestion_runs (
				id,
				source_name,
				started_at,
				status
			)
			VALUES ($1, 'opensky', $2, 'running')
		`,
		runID,
		startedAt,
	)

	if err := fixture.repository.MarkSuccess(
		ctx,
		runID,
		finishedAt,
		20,
		18,
		2,
	); err != nil {
		t.Fatalf("mark ingestion run success: %v", err)
	}

	err := fixture.repository.MarkFailed(
		ctx,
		runID,
		finishedAt.Add(time.Minute),
		99,
		1,
		98,
		"late failure",
	)
	if !errors.Is(err, ErrIngestionRunTransitionRejected) {
		t.Fatalf(
			"expected transition rejection, got %v",
			err,
		)
	}

	var status string
	var recordsReceived int
	var recordsInserted int
	var recordsUpdated int
	var errorMessage string

	if err := fixture.pool.QueryRow(
		ctx,
		`
			SELECT
				status,
				records_received,
				records_inserted,
				records_updated,
				COALESCE(error_message, '')
			FROM ingestion_runs
			WHERE id = $1
		`,
		runID,
	).Scan(
		&status,
		&recordsReceived,
		&recordsInserted,
		&recordsUpdated,
		&errorMessage,
	); err != nil {
		t.Fatalf("load completed ingestion run: %v", err)
	}

	if status != "success" ||
		recordsReceived != 20 ||
		recordsInserted != 18 ||
		recordsUpdated != 2 ||
		errorMessage != "" {
		t.Fatalf(
			"completed ingestion run changed after rejected transition: status=%s received=%d inserted=%d updated=%d error=%q",
			status,
			recordsReceived,
			recordsInserted,
			recordsUpdated,
			errorMessage,
		)
	}

	_, err = fixture.pool.Exec(
		ctx,
		`
			UPDATE ingestion_runs
			SET records_received = 777
			WHERE id = $1
		`,
		runID,
	)
	assertIngestionRunTerminalPostgresCode(t, err, "23514")

	err = fixture.repository.MarkSuccess(
		ctx,
		"22222222-2222-2222-2222-222222222222",
		finishedAt,
		0,
		0,
		0,
	)
	if !errors.Is(err, ErrIngestionRunNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestIngestionRunTerminalIntegrityEnforcesLifecycleShape(
	t *testing.T,
) {
	fixture := newIngestionRunTerminalFixture(t)
	createIngestionRunTerminalSchema(t, fixture.pool)
	applyIngestionRunTerminalMigration(t, fixture.pool)

	_, err := fixture.pool.Exec(
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
				'33333333-3333-3333-3333-333333333333',
				'opensky',
				now(),
				now(),
				'running'
			)
		`,
	)
	assertIngestionRunTerminalPostgresCode(t, err, "23514")

	_, err = fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO ingestion_runs (
				id,
				source_name,
				started_at,
				status
			)
			VALUES (
				'44444444-4444-4444-4444-444444444444',
				'opensky',
				now(),
				'failed'
			)
		`,
	)
	assertIngestionRunTerminalPostgresCode(t, err, "23514")
}

func newIngestionRunTerminalFixture(
	t *testing.T,
) *ingestionRunTerminalFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(
		os.Getenv(ingestionRunTerminalTestDatabaseURL),
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			ingestionRunTerminalTestDatabaseURL,
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL test database: %v", err)
	}

	schemaName := fmt.Sprintf(
		"ingestion_run_terminal_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&ingestionRunTerminalSchemaCounter, 1),
	)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()

	if _, err := bootstrap.Exec(
		ctx,
		"CREATE SCHEMA "+quotedSchema,
	); err != nil {
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

	return &ingestionRunTerminalFixture{
		pool:       pool,
		repository: NewIngestionRunRepository(pool),
	}
}

func createIngestionRunTerminalSchema(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	mustExecIngestionRunTerminalSQL(
		t,
		pool,
		`
			CREATE TABLE ingestion_runs (
				id uuid PRIMARY KEY,
				source_name text NOT NULL,
				region_id uuid,
				started_at timestamptz NOT NULL,
				finished_at timestamptz,
				status text NOT NULL,
				records_received integer NOT NULL DEFAULT 0,
				records_inserted integer NOT NULL DEFAULT 0,
				records_updated integer NOT NULL DEFAULT 0,
				error_message text,
				created_at timestamptz NOT NULL DEFAULT now(),
				CONSTRAINT ingestion_runs_status_check
					CHECK (status IN ('running', 'success', 'failed', 'partial')),
				CONSTRAINT ingestion_runs_counts_check
					CHECK (
						records_received >= 0
						AND records_inserted >= 0
						AND records_updated >= 0
					),
				CONSTRAINT ingestion_runs_time_check
					CHECK (
						finished_at IS NULL
						OR started_at <= finished_at
					)
			)
		`,
	)
}

func applyIngestionRunTerminalMigration(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve ingestion run integration test file path")
	}

	migrationPath := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations/017_ingestion_run_terminal_integrity.sql",
		),
	)
	migrationBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf(
			"read ingestion run terminal migration %s: %v",
			migrationPath,
			err,
		)
	}

	connection, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire PostgreSQL test connection: %v", err)
	}
	defer connection.Release()

	if _, err := connection.Exec(
		context.Background(),
		string(migrationBytes),
	); err != nil {
		_, _ = connection.Exec(context.Background(), "ROLLBACK")
		t.Fatalf("apply ingestion run terminal migration: %v", err)
	}
}

func mustExecIngestionRunTerminalSQL(
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
		t.Fatalf("execute ingestion run terminal SQL: %v", err)
	}
}

func assertIngestionRunTerminalPostgresCode(
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
