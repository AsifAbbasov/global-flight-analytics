package migrator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListMigrationsRejectsNonCanonicalSQLFileName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "10_short.sql")
	if err := os.WriteFile(path, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write invalid migration file: %v", err)
	}

	runner := &Runner{
		migrationsDir: dir,
	}
	if _, err := runner.ListMigrations(); err == nil {
		t.Fatal("expected non-canonical migration file name to be rejected")
	}
}

func TestListMigrationsSortsByVersion(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"002_second.sql": "SELECT 2;",
		"001_first.sql":  "SELECT 1;",
		"notes.txt":      "ignored",
	}

	for name, content := range files {
		path := filepath.Join(dir, name)

		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write file %s: %v", name, err)
		}
	}

	runner := &Runner{
		migrationsDir: dir,
	}

	migrations, err := runner.ListMigrations()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migrations))
	}

	if migrations[0].Version != "001" {
		t.Fatalf("expected first version 001, got %s", migrations[0].Version)
	}

	if migrations[1].Version != "002" {
		t.Fatalf("expected second version 002, got %s", migrations[1].Version)
	}
}

func TestCalculateFileChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "001_test.sql")

	if err := os.WriteFile(path, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	firstChecksum, err := calculateFileChecksum(path)
	if err != nil {
		t.Fatalf("first checksum error: %v", err)
	}

	secondChecksum, err := calculateFileChecksum(path)
	if err != nil {
		t.Fatalf("second checksum error: %v", err)
	}

	if firstChecksum == "" {
		t.Fatal("expected checksum to be non-empty")
	}

	if firstChecksum != secondChecksum {
		t.Fatalf("expected stable checksum, got %s and %s", firstChecksum, secondChecksum)
	}
}
