package migrationaudit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanLocalMigrationsSortsAndCalculatesChecksums(
	t *testing.T,
) {
	dir := t.TempDir()
	writeTestFile(
		t,
		dir,
		"010_second.sql",
		"SELECT 2;\n",
	)
	writeTestFile(
		t,
		dir,
		"001_first.sql",
		"SELECT 1;\n",
	)
	writeTestFile(
		t,
		dir,
		"notes.txt",
		"ignored",
	)

	scan, err := scanLocalMigrations(dir)
	if err != nil {
		t.Fatalf(
			"scanLocalMigrations() error = %v",
			err,
		)
	}
	if len(scan.migrations) != 2 ||
		len(scan.invalid) != 0 {
		t.Fatalf(
			"unexpected scan: %#v",
			scan,
		)
	}
	if scan.migrations[0].Version != "001" ||
		scan.migrations[0].Name != "first" ||
		scan.migrations[1].Version != "010" ||
		scan.migrations[1].Name != "second" {
		t.Fatalf(
			"unexpected order: %#v",
			scan.migrations,
		)
	}
	for _, migration := range scan.migrations {
		if len(migration.Checksum) != 64 {
			t.Fatalf(
				"checksum length = %d for %#v",
				len(migration.Checksum),
				migration,
			)
		}
	}
}

func TestScanLocalMigrationsReportsInvalidSQLFileNames(
	t *testing.T,
) {
	dir := t.TempDir()
	for _, fileName := range []string{
		"10_short.sql",
		"ABC_letters.sql",
		"010.sql",
		"010_invalid-name.sql",
	} {
		writeTestFile(
			t,
			dir,
			fileName,
			"SELECT 1;",
		)
	}

	scan, err := scanLocalMigrations(dir)
	if err != nil {
		t.Fatalf(
			"scanLocalMigrations() error = %v",
			err,
		)
	}
	if len(scan.migrations) != 0 ||
		len(scan.invalid) != 4 {
		t.Fatalf(
			"unexpected scan: %#v",
			scan,
		)
	}
}

func TestParseLocalMigrationFileName(t *testing.T) {
	version, name, err :=
		parseLocalMigrationFileName(
			"010_add_identity.sql",
		)
	if err != nil {
		t.Fatalf(
			"parseLocalMigrationFileName() error = %v",
			err,
		)
	}
	if version != "010" ||
		name != "add_identity" {
		t.Fatalf(
			"result = %q, %q",
			version,
			name,
		)
	}
}

func writeTestFile(
	t *testing.T,
	dir string,
	name string,
	content string,
) {
	t.Helper()

	if err := os.WriteFile(
		filepath.Join(dir, name),
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatalf(
			"write %s: %v",
			name,
			err,
		)
	}
}
