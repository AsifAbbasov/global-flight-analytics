package featurestore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const featureTimestampDatabaseURLEnvironmentVariable = "TEST_DATABASE_URL"

var featureTimestampSchemaCounter uint64

func TestPostgresStorePreservesExactNanosecondsAndRejectsMirrorDrift(
	t *testing.T,
) {
	pool := newFeatureTimestampIntegrationPool(t)
	ctx := context.Background()

	storedAt := time.Date(
		2026,
		time.July,
		20,
		18,
		30,
		0,
		987654321,
		time.UTC,
	)
	store := newPostgresStore(
		pgxPoolClient{pool: pool},
		func() time.Time { return storedAt },
	)

	firstAsOfTime := time.Date(
		2026,
		time.July,
		20,
		18,
		0,
		0,
		123456789,
		time.UTC,
	)
	firstRecord, err := store.Put(
		ctx,
		validPostgresFeatures(
			testTrajectoryID,
			firstAsOfTime,
			"f",
		),
	)
	if err != nil {
		t.Fatalf("put first snapshot: %v", err)
	}
	if !firstRecord.Key.AsOfTime.Equal(firstAsOfTime) ||
		!firstRecord.StoredAt.Equal(storedAt) {
		t.Fatalf(
			"exact timestamps were not preserved: %#v",
			firstRecord,
		)
	}

	if _, err := pool.Exec(
		ctx,
		`UPDATE flight_feature_snapshots
		 SET as_of_time = as_of_time + interval '2 microseconds'
		 WHERE id = $1`,
		firstRecord.ID,
	); err != nil {
		t.Fatalf("corrupt as-of mirror: %v", err)
	}

	_, err = store.Get(ctx, firstRecord.Key)
	assertCorruptTimestampField(t, err, "as_of_time")

	secondAsOfTime := firstAsOfTime.Add(time.Second)
	secondRecord, err := store.Put(
		ctx,
		validPostgresFeatures(
			testTrajectoryID,
			secondAsOfTime,
			"a",
		),
	)
	if err != nil {
		t.Fatalf("put second snapshot: %v", err)
	}

	if _, err := pool.Exec(
		ctx,
		`UPDATE flight_feature_snapshots
		 SET stored_at = stored_at - interval '2 microseconds'
		 WHERE id = $1`,
		secondRecord.ID,
	); err != nil {
		t.Fatalf("corrupt stored mirror: %v", err)
	}

	_, err = store.Get(ctx, secondRecord.Key)
	assertCorruptTimestampField(t, err, "stored_at")
}

func newFeatureTimestampIntegrationPool(
	t *testing.T,
) *pgxpool.Pool {
	t.Helper()

	databaseURL := strings.TrimSpace(
		os.Getenv(featureTimestampDatabaseURLEnvironmentVariable),
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			featureTimestampDatabaseURLEnvironmentVariable,
		)
	}

	setupContext, cancelSetup := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancelSetup()

	bootstrapConnection, err := pgx.Connect(
		setupContext,
		databaseURL,
	)
	if err != nil {
		t.Fatalf("connect PostgreSQL test database: %v", err)
	}

	schemaName := fmt.Sprintf(
		"feature_timestamp_consistency_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&featureTimestampSchemaCounter, 1),
	)
	quotedSchemaName := pgx.Identifier{schemaName}.Sanitize()
	if _, err := bootstrapConnection.Exec(
		setupContext,
		"CREATE SCHEMA "+quotedSchemaName,
	); err != nil {
		_ = bootstrapConnection.Close(setupContext)
		t.Fatalf("create PostgreSQL test schema: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		dropFeatureTimestampSchema(
			t,
			bootstrapConnection,
			quotedSchemaName,
		)
		_ = bootstrapConnection.Close(setupContext)
		t.Fatalf("parse PostgreSQL pool config: %v", err)
	}
	if poolConfig.ConnConfig.RuntimeParams == nil {
		poolConfig.ConnConfig.RuntimeParams = make(map[string]string)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName

	pool, err := pgxpool.NewWithConfig(setupContext, poolConfig)
	if err != nil {
		dropFeatureTimestampSchema(
			t,
			bootstrapConnection,
			quotedSchemaName,
		)
		_ = bootstrapConnection.Close(setupContext)
		t.Fatalf("create PostgreSQL test pool: %v", err)
	}
	if err := pool.Ping(setupContext); err != nil {
		pool.Close()
		dropFeatureTimestampSchema(
			t,
			bootstrapConnection,
			quotedSchemaName,
		)
		_ = bootstrapConnection.Close(setupContext)
		t.Fatalf("ping PostgreSQL test pool: %v", err)
	}

	if _, err := pool.Exec(
		setupContext,
		`CREATE TABLE flight_feature_snapshots (
			id text PRIMARY KEY,
			trajectory_id uuid NOT NULL,
			schema_version text NOT NULL,
			as_of_time timestamptz NOT NULL,
			as_of_time_unix_nano bigint NOT NULL,
			input_fingerprint text NOT NULL,
			validation_status text NOT NULL,
			features_json jsonb NOT NULL,
			stored_at timestamptz NOT NULL,
			stored_at_unix_nano bigint NOT NULL,
			UNIQUE (
				trajectory_id,
				schema_version,
				as_of_time_unix_nano
			)
		)`,
	); err != nil {
		pool.Close()
		dropFeatureTimestampSchema(
			t,
			bootstrapConnection,
			quotedSchemaName,
		)
		_ = bootstrapConnection.Close(setupContext)
		t.Fatalf("create feature snapshot table: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		dropFeatureTimestampSchema(
			t,
			bootstrapConnection,
			quotedSchemaName,
		)
		cleanupContext, cancelCleanup := context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
		defer cancelCleanup()
		if err := bootstrapConnection.Close(cleanupContext); err != nil {
			t.Errorf("close PostgreSQL bootstrap connection: %v", err)
		}
	})

	return pool
}

func dropFeatureTimestampSchema(
	t *testing.T,
	connection *pgx.Conn,
	quotedSchemaName string,
) {
	t.Helper()

	cleanupContext, cancelCleanup := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancelCleanup()

	if _, err := connection.Exec(
		cleanupContext,
		"DROP SCHEMA IF EXISTS "+quotedSchemaName+" CASCADE",
	); err != nil && !errors.Is(err, context.Canceled) {
		t.Errorf("drop PostgreSQL test schema: %v", err)
	}
}
