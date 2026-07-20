package postgres

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIngestionRunRepositoryGuardsTerminalTransitions(
	t *testing.T,
) {
	t.Parallel()

	sourceBytes, err := os.ReadFile(
		"ingestionrun_repository.go",
	)
	if err != nil {
		t.Fatalf(
			"read ingestion run repository source: %v",
			err,
		)
	}

	source := string(sourceBytes)
	for _, required := range []string{
		"ErrIngestionRunTransitionRejected",
		"WITH updated AS (",
		"AND status = $8",
		"transition_rejected",
		"ingestionrun.StatusRunning",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf(
				"ingestion run repository is missing terminal transition guard %q",
				required,
			)
		}
	}
}

func TestIngestionRunTerminalMigrationDefinesDatabaseProtection(
	t *testing.T,
) {
	t.Parallel()

	migrationPath := filepath.Clean(
		filepath.Join(
			"../../../../../database/migrations",
			"017_ingestion_run_terminal_integrity.sql",
		),
	)

	migrationBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf(
			"read ingestion run terminal migration: %v",
			err,
		)
	}

	migration := string(migrationBytes)
	for _, required := range []string{
		"ingestion_runs_lifecycle_check",
		"enforce_ingestion_run_terminal_immutability",
		"OLD.status IN ('success', 'failed', 'partial')",
		"NEW IS DISTINCT FROM OLD",
		"ERRCODE = '23514'",
		"CREATE TRIGGER ingestion_runs_terminal_immutability",
	} {
		if !strings.Contains(migration, required) {
			t.Fatalf(
				"ingestion run terminal migration is missing %q",
				required,
			)
		}
	}
}
