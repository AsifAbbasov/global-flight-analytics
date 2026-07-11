package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const reconciliationTestDatabaseURL = "TEST_DATABASE_URL"

var reconciliationSchemaCounter uint64

type reconciliationFixture struct {
	pool       *pgxpool.Pool
	adminPool  *pgxpool.Pool
	schemaName string
	repository *ReconciliationRepository
}

func TestReconciliationRepositoryMarksPendingDerivation(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	observedAt := time.Date(
		2026,
		time.July,
		11,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		reconciliation.PendingDerivation{
			ICAO24:         "ABC123",
			DerivationType: reconciliation.DerivationTypeTrajectory,
			ObservedFrom:   observedAt,
			ObservedTo:     observedAt.Add(time.Minute),
			LastError:      "trajectory insert failed",
		},
	)
	if err != nil {
		t.Fatalf(
			"mark pending derivation: %v",
			err,
		)
	}

	var status string
	var icao24 string
	var lastError string

	err = fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT status, icao24, last_error
			FROM derived_reconciliation_tasks
			WHERE derivation_type = 'trajectory';
		`,
	).Scan(
		&status,
		&icao24,
		&lastError,
	)
	if err != nil {
		t.Fatalf(
			"load pending derivation task: %v",
			err,
		)
	}

	if status != string(reconciliation.TaskStatusPending) {
		t.Fatalf(
			"expected pending status, got %s",
			status,
		)
	}

	if icao24 != "abc123" {
		t.Fatalf(
			"expected normalized icao24 abc123, got %s",
			icao24,
		)
	}

	if lastError != "trajectory insert failed" {
		t.Fatalf(
			"expected last error to be stored, got %s",
			lastError,
		)
	}
}

func TestReconciliationRepositoryUpsertsExistingPendingDerivation(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	observedAt := time.Date(
		2026,
		time.July,
		11,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	task := reconciliation.PendingDerivation{
		ICAO24:         "ABC123",
		DerivationType: reconciliation.DerivationTypeTrajectory,
		ObservedFrom:   observedAt,
		ObservedTo:     observedAt.Add(time.Minute),
		LastError:      "first failure",
	}

	if err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		task,
	); err != nil {
		t.Fatalf(
			"mark first pending derivation: %v",
			err,
		)
	}

	task.LastError = "second failure"

	if err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		task,
	); err != nil {
		t.Fatalf(
			"mark second pending derivation: %v",
			err,
		)
	}

	var count int
	var lastError string

	err := fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*), MAX(last_error)
			FROM derived_reconciliation_tasks;
		`,
	).Scan(
		&count,
		&lastError,
	)
	if err != nil {
		t.Fatalf(
			"load reconciliation task count: %v",
			err,
		)
	}

	if count != 1 {
		t.Fatalf(
			"expected 1 upserted task, got %d",
			count,
		)
	}

	if lastError != "second failure" {
		t.Fatalf(
			"expected updated last error, got %s",
			lastError,
		)
	}
}

func newReconciliationFixture(
	t *testing.T,
) *reconciliationFixture {
	t.Helper()

	databaseURL := os.Getenv(
		reconciliationTestDatabaseURL,
	)
	if databaseURL == "" {
		t.Skipf(
			"%s is not set; skipping PostgreSQL integration test",
			reconciliationTestDatabaseURL,
		)
	}

	ctx := context.Background()
	schemaName := fmt.Sprintf(
		"reconciliation_test_%d_%d",
		time.Now().UnixNano(),
		atomic.AddUint64(
			&reconciliationSchemaCounter,
			1,
		),
	)

	adminPool, err := pgxpool.New(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"connect admin postgres: %v",
			err,
		)
	}

	_, err = adminPool.Exec(
		ctx,
		"CREATE SCHEMA "+pgx.Identifier{schemaName}.Sanitize(),
	)
	if err != nil {
		adminPool.Close()
		t.Fatalf(
			"create reconciliation schema: %v",
			err,
		)
	}

	poolConfig, err := pgxpool.ParseConfig(
		databaseURL,
	)
	if err != nil {
		adminPool.Close()
		t.Fatalf(
			"parse postgres config: %v",
			err,
		)
	}

	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName

	pool, err := pgxpool.NewWithConfig(
		ctx,
		poolConfig,
	)
	if err != nil {
		adminPool.Close()
		t.Fatalf(
			"connect schema postgres: %v",
			err,
		)
	}

	applyReconciliationMigration(
		t,
		pool,
	)

	return &reconciliationFixture{
		pool:       pool,
		adminPool:  adminPool,
		schemaName: schemaName,
		repository: NewReconciliationRepository(
			pool,
		),
	}
}

func (fixture *reconciliationFixture) close(
	t *testing.T,
) {
	t.Helper()

	fixture.pool.Close()

	_, err := fixture.adminPool.Exec(
		context.Background(),
		"DROP SCHEMA IF EXISTS "+pgx.Identifier{fixture.schemaName}.Sanitize()+" CASCADE",
	)
	fixture.adminPool.Close()

	if err != nil {
		t.Fatalf(
			"drop reconciliation schema: %v",
			err,
		)
	}
}

func applyReconciliationMigration(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve reconciliation integration test path",
		)
	}

	migrationPath := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations/008_create_derived_reconciliation_tasks.sql",
		),
	)

	migrationSQL, err := os.ReadFile(
		migrationPath,
	)
	if err != nil {
		t.Fatalf(
			"read reconciliation migration %s: %v",
			migrationPath,
			err,
		)
	}

	_, err = pool.Exec(
		context.Background(),
		string(migrationSQL),
	)
	if err != nil {
		t.Fatalf(
			"execute reconciliation migration: %v",
			err,
		)
	}
}
