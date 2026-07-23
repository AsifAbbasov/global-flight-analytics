package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIngestionRunActiveLifecycleUpdatesDeletesAndMarksPartial(
	t *testing.T,
) {
	fixture := newIngestionRunTerminalFixture(t)
	createIngestionRunTerminalSchema(t, fixture.pool)
	applyIngestionRunTerminalMigration(t, fixture.pool)
	mustExecIngestionRunTerminalSQL(
		t,
		fixture.pool,
		`
			CREATE TABLE flight_states (
				id uuid PRIMARY KEY,
				ingestion_run_id uuid REFERENCES ingestion_runs(id)
			)
		`,
	)

	ctx := context.Background()
	startedAt := time.Date(
		2026,
		time.July,
		23,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	partialRunID := "55555555-5555-5555-5555-555555555555"
	deleteRunID := "66666666-6666-6666-6666-666666666666"

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
			VALUES
				($1, 'airplanes.live', $3, 'running'),
				($2, 'airplanes.live', $3, 'running')
		`,
		partialRunID,
		deleteRunID,
		startedAt,
	)

	if err := fixture.repository.UpdateRunningSource(
		ctx,
		partialRunID,
		"opensky",
	); err != nil {
		t.Fatalf("update running source: %v", err)
	}
	if err := fixture.repository.MarkPartial(
		ctx,
		partialRunID,
		startedAt.Add(time.Minute),
		2,
		1,
		0,
		"trajectory persistence failed",
	); err != nil {
		t.Fatalf("mark partial: %v", err)
	}
	if err := fixture.repository.DeleteRunning(
		ctx,
		deleteRunID,
	); err != nil {
		t.Fatalf("delete provisional run: %v", err)
	}

	var sourceName string
	var status string
	var recordsInserted int
	if err := fixture.pool.QueryRow(
		ctx,
		`
			SELECT source_name, status, records_inserted
			FROM ingestion_runs
			WHERE id = $1
		`,
		partialRunID,
	).Scan(
		&sourceName,
		&status,
		&recordsInserted,
	); err != nil {
		t.Fatalf("read partial run: %v", err)
	}
	if sourceName != "opensky" ||
		status != "partial" ||
		recordsInserted != 1 {
		t.Fatalf(
			"source=%s status=%s inserted=%d, want opensky partial 1",
			sourceName,
			status,
			recordsInserted,
		)
	}

	var deletedCount int
	if err := fixture.pool.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM ingestion_runs WHERE id = $1`,
		deleteRunID,
	).Scan(&deletedCount); err != nil {
		t.Fatalf("count deleted run: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("deleted run count = %d, want 0", deletedCount)
	}
}
