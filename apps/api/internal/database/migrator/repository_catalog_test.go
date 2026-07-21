package migrator

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRepositoryMigrationCatalogIsCanonicalAndUnique(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve repository migration catalog test path")
	}
	migrationDirectory := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations",
		),
	)

	runner := &Runner{migrationsDir: migrationDirectory}
	migrations, err := runner.ListMigrations()
	if err != nil {
		t.Fatalf("list repository migration catalog: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("repository migration catalog is empty")
	}

	seenFiles := make(map[string]bool, len(migrations))
	for _, migration := range migrations {
		seenFiles[filepath.Base(migration.Path)] = true
	}
	for _, required := range []string{
		"016_add_flight_state_observation_metadata.sql",
		"019_data_quality_parent_integrity.sql",
		"020_stage14_correctness_hardening.sql",
	} {
		if !seenFiles[required] {
			t.Fatalf("repository migration catalog is missing %s", required)
		}
	}

	retiredPath := filepath.Join(migrationDirectory, "016_data_quality_parent_integrity.sql")
	if _, err := os.Stat(retiredPath); err == nil {
		t.Fatalf("retired duplicate migration still exists: %s", retiredPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("inspect retired duplicate migration: %v", err)
	}
}
