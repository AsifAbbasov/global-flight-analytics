package migrator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMigrationFileName(t *testing.T) {
	version, name, err := parseMigrationFileName("003_weather_foundation.sql")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if version != "003" {
		t.Fatalf("expected version 003, got %s", version)
	}

	if name != "weather_foundation" {
		t.Fatalf("expected name weather_foundation, got %s", name)
	}
}

func TestParseMigrationFileNameRejectsInvalidName(t *testing.T) {
	_, _, err := parseMigrationFileName("invalid.sql")
	if err == nil {
		t.Fatal("expected error")
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
