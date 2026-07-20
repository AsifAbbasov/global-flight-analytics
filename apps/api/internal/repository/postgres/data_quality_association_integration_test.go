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

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const qualityAssociationTestDatabaseURL = "TEST_DATABASE_URL"

var qualityAssociationSchemaCounter uint64

type qualityAssociationFixture struct {
	pool       *pgxpool.Pool
	repository *DataQualityRepository
}

func TestDataQualityAssociationMigrationPreservesTypedStateIdentity(t *testing.T) {
	fixture := newQualityAssociationFixture(t)
	createLegacyQualitySchema(t, fixture.pool)

	persistedID := "11111111-1111-1111-1111-111111111111"
	rejectedID := "22222222-2222-2222-2222-222222222222"

	if _, err := fixture.pool.Exec(
		context.Background(),
		`INSERT INTO flight_states (id) VALUES ($1)`,
		persistedID,
	); err != nil {
		t.Fatalf("insert persisted flight state: %v", err)
	}

	if _, err := fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO data_quality_reports (
				id, object_type, object_id,
				validation_status, completeness, confidence, score
			)
			VALUES
				(
					'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
					'flight_state', $1, 'valid', 'complete', 'high', 1
				),
				(
					'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
					'flight_state', $2, 'invalid', 'insufficient', 'none', 0
				)
		`,
		persistedID,
		rejectedID,
	); err != nil {
		t.Fatalf("insert legacy quality reports: %v", err)
	}

	if err := applyQualityAssociationMigration(t, fixture.pool); err != nil {
		t.Fatalf("apply quality association migration: %v", err)
	}

	assertLegacyQualityColumnsRemoved(t, fixture.pool)
	assertQualityAssociation(t, fixture.pool, persistedID, persistedID)
	assertQualityAssociation(t, fixture.pool, rejectedID, "")

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO data_quality_reports (
				id, state_id, flight_state_id,
				validation_status, completeness, confidence, score
			)
			VALUES (
				'dddddddd-dddd-dddd-dddd-dddddddddddd',
				'66666666-6666-6666-6666-666666666666',
				'66666666-6666-6666-6666-666666666666',
				'valid', 'complete', 'high', 1
			)
		`,
	)
	assertPostgresCode(t, err, "23503")

	_, err = fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO data_quality_reports (
				id, state_id, flight_state_id,
				validation_status, completeness, confidence, score
			)
			VALUES (
				'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
				'77777777-7777-7777-7777-777777777777',
				$1,
				'valid', 'complete', 'high', 1
			)
		`,
		persistedID,
	)
	assertPostgresCode(t, err, "23514")
}

func TestDataQualityAssociationMigrationRejectsUnsupportedLegacyObjectType(t *testing.T) {
	fixture := newQualityAssociationFixture(t)
	createLegacyQualitySchema(t, fixture.pool)

	mustExecQualitySQL(
		t,
		fixture.pool,
		`
			INSERT INTO data_quality_reports (
				id, object_type, object_id,
				validation_status, completeness, confidence, score
			)
			VALUES (
				'cccccccc-cccc-cccc-cccc-cccccccccccc',
				'trajectory',
				'33333333-3333-3333-3333-333333333333',
				'partial', 'partial', 'medium', 0.5
			)
		`,
	)

	err := applyQualityAssociationMigration(t, fixture.pool)
	if err == nil {
		t.Fatal("expected unsupported legacy object type to reject migration")
	}
	if !strings.Contains(
		err.Error(),
		"unsupported data_quality_reports object_type values exist",
	) {
		t.Fatalf("expected unsupported object type error, got %v", err)
	}
}

func TestDataQualityRepositorySeparatesPersistedAndRejectedStateEvidence(t *testing.T) {
	fixture := newQualityAssociationFixture(t)
	createLegacyQualitySchema(t, fixture.pool)
	createQualityParentIntegritySupportSchema(t, fixture.pool)

	if err := applyQualityAssociationMigration(t, fixture.pool); err != nil {
		t.Fatalf("apply quality association migration: %v", err)
	}
	if err := applyDataQualityParentIntegrityMigration(t, fixture.pool); err != nil {
		t.Fatalf("apply data quality parent integrity migration: %v", err)
	}

	persistedID := "44444444-4444-4444-4444-444444444444"
	rejectedID := "55555555-5555-5555-5555-555555555555"

	if _, err := fixture.pool.Exec(
		context.Background(),
		`INSERT INTO flight_states (id) VALUES ($1)`,
		persistedID,
	); err != nil {
		t.Fatalf("insert repository test flight state: %v", err)
	}

	quality := dataquality.DataQuality{
		ValidationStatus: dataquality.ValidationStatusValid,
		Completeness:     dataquality.CompletenessLevelComplete,
		Confidence:       dataquality.ConfidenceLevelHigh,
		Score:            0.95,
		MissingFields:    []string{},
		Warnings: []dataquality.Warning{
			{
				Code:    "TEST_WARNING",
				Message: "test warning",
				Field:   "latitude",
			},
		},
	}

	if err := fixture.repository.SaveFlightStateQuality(
		context.Background(),
		flightstate.FlightState{ID: persistedID},
		quality,
	); err != nil {
		t.Fatalf("save persisted state quality: %v", err)
	}

	rejectedQuality := quality
	rejectedQuality.ValidationStatus = dataquality.ValidationStatusInvalid
	rejectedQuality.Completeness = dataquality.CompletenessLevelInsufficient
	rejectedQuality.Confidence = dataquality.ConfidenceLevelNone
	rejectedQuality.Score = 0

	if err := fixture.repository.SaveFlightStateQuality(
		context.Background(),
		flightstate.FlightState{
			ID:         rejectedID,
			ICAO24:     "REJECTED",
			ObservedAt: time.Date(2026, time.July, 20, 8, 0, 0, 0, time.UTC),
			SourceName: "test",
		},
		rejectedQuality,
	); err != nil {
		t.Fatalf("save rejected state quality: %v", err)
	}

	assertQualityAssociation(t, fixture.pool, persistedID, persistedID)
	assertRejectedQualityAssociation(t, fixture.pool, rejectedID)
	assertNoQualityAssociation(t, fixture.pool, rejectedID)

	if _, err := fixture.pool.Exec(
		context.Background(),
		`DELETE FROM flight_states WHERE id = $1`,
		persistedID,
	); err != nil {
		t.Fatalf("delete persisted flight state: %v", err)
	}

	assertNoQualityAssociation(t, fixture.pool, persistedID)
}

func newQualityAssociationFixture(t *testing.T) *qualityAssociationFixture {
	t.Helper()

	databaseURL := strings.TrimSpace(os.Getenv(qualityAssociationTestDatabaseURL))
	if databaseURL == "" {
		t.Skipf("%s is not set; skipping PostgreSQL integration test", qualityAssociationTestDatabaseURL)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bootstrap, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to PostgreSQL test database: %v", err)
	}

	schemaName := fmt.Sprintf(
		"quality_association_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(&qualityAssociationSchemaCounter, 1),
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

		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	return &qualityAssociationFixture{
		pool:       pool,
		repository: NewDataQualityRepository(pool),
	}
}

func createLegacyQualitySchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	mustExecQualitySQL(
		t,
		pool,
		`
			CREATE TABLE flight_states (
				id uuid PRIMARY KEY
			);

			CREATE TABLE data_quality_reports (
				id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
				object_type text NOT NULL,
				object_id uuid,
				validation_status text NOT NULL,
				completeness text NOT NULL,
				confidence text NOT NULL,
				score numeric NOT NULL DEFAULT 0,
				missing_fields text[] NOT NULL DEFAULT '{}',
				warnings_json jsonb NOT NULL DEFAULT '[]'::jsonb,
				calculated_at timestamptz NOT NULL DEFAULT now(),
				created_at timestamptz NOT NULL DEFAULT now()
			);

			CREATE INDEX data_quality_reports_object_idx
				ON data_quality_reports (object_type, object_id);
		`,
	)
}

func applyQualityAssociationMigration(t *testing.T, pool *pgxpool.Pool) error {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve quality association integration test file path")
	}

	migrationPath := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations/007_data_quality_association_integrity.sql",
		),
	)
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read quality association migration %s: %v", migrationPath, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	connection, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire PostgreSQL test connection: %v", err)
	}
	defer connection.Release()

	_, err = connection.Exec(ctx, string(sqlBytes))
	if err != nil {
		_, _ = connection.Exec(ctx, "ROLLBACK")
	}
	return err
}

func mustExecQualitySQL(
	t *testing.T,
	pool *pgxpool.Pool,
	query string,
) {
	t.Helper()

	if _, err := pool.Exec(context.Background(), query); err != nil {
		t.Fatalf("execute quality association SQL: %v", err)
	}
}

func assertLegacyQualityColumnsRemoved(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	var count int
	err := pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM information_schema.columns
			WHERE table_schema = current_schema()
				AND table_name = 'data_quality_reports'
				AND column_name IN ('object_type', 'object_id')
		`,
	).Scan(&count)
	if err != nil {
		t.Fatalf("inspect legacy polymorphic columns: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected legacy polymorphic columns removed, got %d", count)
	}
}

func assertQualityAssociation(
	t *testing.T,
	pool *pgxpool.Pool,
	stateID string,
	expectedFlightStateID string,
) {
	t.Helper()

	var loadedStateID string
	var loadedFlightStateID string

	err := pool.QueryRow(
		context.Background(),
		`
			SELECT state_id::text, COALESCE(flight_state_id::text, '')
			FROM data_quality_reports
			WHERE state_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`,
		stateID,
	).Scan(&loadedStateID, &loadedFlightStateID)
	if err != nil {
		t.Fatalf("load quality association for state %s: %v", stateID, err)
	}
	if loadedStateID != stateID {
		t.Fatalf("expected state id %s, got %s", stateID, loadedStateID)
	}
	if loadedFlightStateID != expectedFlightStateID {
		t.Fatalf(
			"expected flight state id %q, got %q",
			expectedFlightStateID,
			loadedFlightStateID,
		)
	}
}

func assertPostgresCode(t *testing.T, err error, expectedCode string) {
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
			"expected PostgreSQL error code %s, got %s",
			expectedCode,
			postgresError.Code,
		)
	}
}
