package postgres

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRecoverStaleRunningFinalizesOnlyExpiredRuns(
	t *testing.T,
) {
	fixture := newIngestionRunTerminalFixture(t)
	createIngestionRunTerminalSchema(t, fixture.pool)
	applyIngestionRunTerminalMigration(t, fixture.pool)

	ctx := context.Background()
	now := time.Date(2026, time.July, 22, 20, 0, 0, 0, time.UTC)
	staleID := "55555555-5555-5555-5555-555555555555"
	freshID := "66666666-6666-6666-6666-666666666666"

	mustExecIngestionRunTerminalSQL(
		t,
		fixture.pool,
		`INSERT INTO ingestion_runs (id, source_name, started_at, status)
		 VALUES ($1, 'airplanes.live', $2, 'running')`,
		staleID,
		now.Add(-time.Hour),
	)
	mustExecIngestionRunTerminalSQL(
		t,
		fixture.pool,
		`INSERT INTO ingestion_runs (id, source_name, started_at, status)
		 VALUES ($1, 'airplanes.live', $2, 'running')`,
		freshID,
		now.Add(-5*time.Minute),
	)

	recoveredCount, err := fixture.repository.RecoverStaleRunning(
		ctx,
		now.Add(-30*time.Minute),
		now,
		"ingestion process stopped before terminal status was recorded",
	)
	if err != nil {
		t.Fatalf("recover stale ingestion runs: %v", err)
	}
	if recoveredCount != 1 {
		t.Fatalf("recovered count = %d, want 1", recoveredCount)
	}

	assertIngestionRunLifecycle(
		t,
		fixture,
		staleID,
		"failed",
		true,
		"ingestion process stopped before terminal status was recorded",
	)
	assertIngestionRunLifecycle(t, fixture, freshID, "running", false, "")
}

func TestRecoverStaleRunningIsConcurrencySafe(
	t *testing.T,
) {
	fixture := newIngestionRunTerminalFixture(t)
	createIngestionRunTerminalSchema(t, fixture.pool)
	applyIngestionRunTerminalMigration(t, fixture.pool)

	ctx := context.Background()
	now := time.Date(2026, time.July, 22, 20, 0, 0, 0, time.UTC)
	runID := "77777777-7777-7777-7777-777777777777"
	mustExecIngestionRunTerminalSQL(
		t,
		fixture.pool,
		`INSERT INTO ingestion_runs (id, source_name, started_at, status)
		 VALUES ($1, 'opensky', $2, 'running')`,
		runID,
		now.Add(-time.Hour),
	)

	const workers = 4
	var waitGroup sync.WaitGroup
	counts := make(chan int64, workers)
	errorsChannel := make(chan error, workers)
	for index := 0; index < workers; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			count, err := fixture.repository.RecoverStaleRunning(
				ctx,
				now.Add(-30*time.Minute),
				now,
				"ingestion process stopped before terminal status was recorded",
			)
			counts <- count
			errorsChannel <- err
		}()
	}
	waitGroup.Wait()
	close(counts)
	close(errorsChannel)

	var total int64
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent recovery: %v", err)
		}
	}
	for count := range counts {
		total += count
	}
	if total != 1 {
		t.Fatalf("total recovered count = %d, want 1", total)
	}

	assertIngestionRunLifecycle(
		t,
		fixture,
		runID,
		"failed",
		true,
		"ingestion process stopped before terminal status was recorded",
	)
}

func assertIngestionRunLifecycle(
	t *testing.T,
	fixture *ingestionRunTerminalFixture,
	runID string,
	expectedStatus string,
	expectedFinished bool,
	expectedError string,
) {
	t.Helper()

	var status string
	var finished bool
	var errorMessage string
	if err := fixture.pool.QueryRow(
		context.Background(),
		`SELECT status, finished_at IS NOT NULL, COALESCE(error_message, '')
		 FROM ingestion_runs
		 WHERE id = $1`,
		runID,
	).Scan(&status, &finished, &errorMessage); err != nil {
		t.Fatalf("load ingestion run lifecycle: %v", err)
	}

	if status != expectedStatus ||
		finished != expectedFinished ||
		errorMessage != expectedError {
		t.Fatalf(
			"run lifecycle status=%s finished=%t error=%q, want status=%s finished=%t error=%q",
			status,
			finished,
			errorMessage,
			expectedStatus,
			expectedFinished,
			expectedError,
		)
	}
}
