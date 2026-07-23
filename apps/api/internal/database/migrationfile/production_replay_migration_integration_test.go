package migrationfile

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

var replayMigrationSchemaCounter uint64

func TestProductionReplayMigrationEnforcesObservationIdentity(
	t *testing.T,
) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}

	schemaName := fmt.Sprintf(
		"flight_state_replay_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&replayMigrationSchemaCounter, 1),
	)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()
	if _, err := bootstrap.Exec(
		ctx,
		"CREATE SCHEMA "+quotedSchema,
	); err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create schema: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("parse pool config: %v", err)
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = bootstrap.Close(ctx)
		t.Fatalf("create pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		cleanupContext, cleanupCancel := context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
		defer cleanupCancel()
		if _, err := bootstrap.Exec(
			cleanupContext,
			"DROP SCHEMA IF EXISTS "+quotedSchema+" CASCADE",
		); err != nil {
			t.Errorf("drop schema: %v", err)
		}
		if err := bootstrap.Close(cleanupContext); err != nil {
			t.Errorf("close bootstrap connection: %v", err)
		}
	})

	if _, err := pool.Exec(
		ctx,
		`
			CREATE TABLE flight_states (
				id uuid PRIMARY KEY,
				source_name text NOT NULL,
				icao24 varchar(10) NOT NULL,
				observed_at timestamptz NOT NULL
			)
		`,
	); err != nil {
		t.Fatalf("create flight_states: %v", err)
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve replay migration test file path")
	}
	migrationPath := filepath.Clean(filepath.Join(
		filepath.Dir(currentFile),
		"../../../../../database/migrations/023_ingestion_durability_replay_partial.sql",
	))
	migrationBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read replay migration: %v", err)
	}
	if _, err := pool.Exec(ctx, string(migrationBytes)); err != nil {
		t.Fatalf("apply replay migration: %v", err)
	}

	observedAt := time.Date(
		2026,
		time.July,
		23,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	if _, err := pool.Exec(
		ctx,
		`
			INSERT INTO flight_states (
				id, source_name, icao24, observed_at
			)
			VALUES ($1, $2, $3, $4)
		`,
		"11111111-1111-1111-1111-111111111111",
		"opensky",
		"ABC123",
		observedAt,
	); err != nil {
		t.Fatalf("insert first observation: %v", err)
	}

	commandTag, err := pool.Exec(
		ctx,
		`
			INSERT INTO flight_states (
				id, source_name, icao24, observed_at
			)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (source_name, icao24, observed_at)
			DO NOTHING
		`,
		"22222222-2222-2222-2222-222222222222",
		"opensky",
		"ABC123",
		observedAt,
	)
	if err != nil {
		t.Fatalf("replay insert with conflict policy: %v", err)
	}
	if commandTag.RowsAffected() != 0 {
		t.Fatalf(
			"replay rows affected = %d, want 0",
			commandTag.RowsAffected(),
		)
	}

	_, err = pool.Exec(
		ctx,
		`
			INSERT INTO flight_states (
				id, source_name, icao24, observed_at
			)
			VALUES ($1, $2, $3, $4)
		`,
		"33333333-3333-3333-3333-333333333333",
		"opensky",
		"ABC123",
		observedAt,
	)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23505" {
		t.Fatalf("expected unique violation 23505, got %v", err)
	}
}
