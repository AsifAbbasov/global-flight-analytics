package postgres

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDataQualityParentIntegrityMigrationMovesRejectedEvidenceAndEnforcesParent(
	t *testing.T,
) {
	fixture := newQualityAssociationFixture(t)
	createLegacyQualitySchema(t, fixture.pool)
	createQualityParentIntegritySupportSchema(t, fixture.pool)

	persistedID := "11111111-1111-1111-1111-111111111111"
	rejectedID := "22222222-2222-2222-2222-222222222222"

	mustExecQualitySQL(
		t,
		fixture.pool,
		`INSERT INTO flight_states (id) VALUES ('11111111-1111-1111-1111-111111111111')`,
	)
	mustExecQualitySQL(
		t,
		fixture.pool,
		`
			INSERT INTO data_quality_reports (
				id,
				object_type,
				object_id,
				validation_status,
				completeness,
				confidence,
				score
			)
			VALUES
				(
					'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
					'flight_state',
					'11111111-1111-1111-1111-111111111111',
					'valid',
					'complete',
					'high',
					1
				),
				(
					'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
					'flight_state',
					'22222222-2222-2222-2222-222222222222',
					'invalid',
					'insufficient',
					'none',
					0
				)
		`,
	)

	if err := applyQualityAssociationMigration(t, fixture.pool); err != nil {
		t.Fatalf("apply quality association migration: %v", err)
	}
	if err := applyDataQualityParentIntegrityMigration(t, fixture.pool); err != nil {
		t.Fatalf("apply data quality parent integrity migration: %v", err)
	}

	assertQualityAssociation(t, fixture.pool, persistedID, persistedID)
	assertRejectedQualityAssociation(t, fixture.pool, rejectedID)
	assertNoQualityAssociation(t, fixture.pool, rejectedID)

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO data_quality_reports (
				state_id,
				flight_state_id,
				validation_status,
				completeness,
				confidence,
				score
			)
			VALUES (
				'33333333-3333-3333-3333-333333333333',
				NULL,
				'valid',
				'complete',
				'high',
				1
			)
		`,
	)
	assertPostgresCode(t, err, "23502")

	_, err = fixture.pool.Exec(
		context.Background(),
		`
			INSERT INTO data_quality_reports (
				state_id,
				flight_state_id,
				validation_status,
				completeness,
				confidence,
				score
			)
			VALUES (
				'44444444-4444-4444-4444-444444444444',
				'44444444-4444-4444-4444-444444444444',
				'valid',
				'complete',
				'high',
				1
			)
		`,
	)
	assertPostgresCode(t, err, "23503")

	mustExecQualitySQL(
		t,
		fixture.pool,
		`DELETE FROM flight_states WHERE id = '11111111-1111-1111-1111-111111111111'`,
	)
	assertNoQualityAssociation(t, fixture.pool, persistedID)
}

func createQualityParentIntegritySupportSchema(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	mustExecQualitySQL(
		t,
		pool,
		`
			CREATE TABLE ingestion_runs (
				id uuid PRIMARY KEY
			)
		`,
	)
}

func applyDataQualityParentIntegrityMigration(
	t *testing.T,
	pool *pgxpool.Pool,
) error {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve data quality parent integrity integration test path")
	}

	migrationPath := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations/019_data_quality_parent_integrity.sql",
		),
	)
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read data quality parent integrity migration %s: %v", migrationPath, err)
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

func assertRejectedQualityAssociation(
	t *testing.T,
	pool *pgxpool.Pool,
	stateID string,
) {
	t.Helper()

	var loadedStateID string
	err := pool.QueryRow(
		context.Background(),
		`
			SELECT state_id::text
			FROM rejected_flight_state_quality_reports
			WHERE state_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`,
		stateID,
	).Scan(&loadedStateID)
	if err != nil {
		t.Fatalf("load rejected quality evidence for state %s: %v", stateID, err)
	}
	if loadedStateID != stateID {
		t.Fatalf("expected rejected state id %s, got %s", stateID, loadedStateID)
	}
}

func assertNoQualityAssociation(
	t *testing.T,
	pool *pgxpool.Pool,
	stateID string,
) {
	t.Helper()

	var count int
	err := pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM data_quality_reports
			WHERE state_id = $1
		`,
		stateID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count quality reports for state %s: %v", stateID, err)
	}
	if count != 0 {
		t.Fatalf("expected no canonical quality report for state %s, got %d", stateID, count)
	}
}
