package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateMigrationsDirReturnsAbsoluteCleanDirectory(
	t *testing.T,
) {
	migrationsDir := t.TempDir()

	resolved, err := validateMigrationsDir(
		migrationsDir,
	)
	if err != nil {
		t.Fatalf(
			"validate migrations directory: %v",
			err,
		)
	}

	expected, err := filepath.Abs(
		migrationsDir,
	)
	if err != nil {
		t.Fatalf(
			"resolve expected absolute path: %v",
			err,
		)
	}

	expected = filepath.Clean(
		expected,
	)

	if resolved != expected {
		t.Fatalf(
			"expected resolved directory %q, got %q",
			expected,
			resolved,
		)
	}
}

func TestValidateMigrationsDirRejectsEmptyPath(
	t *testing.T,
) {
	resolved, err := validateMigrationsDir(
		"   ",
	)

	if err == nil {
		t.Fatal(
			"expected empty migrations directory path to be rejected",
		)
	}

	if resolved != "" {
		t.Fatalf(
			"expected empty resolved path, got %q",
			resolved,
		)
	}

	if !strings.Contains(
		err.Error(),
		"migrations directory path is required",
	) {
		t.Fatalf(
			"expected required path error, got %q",
			err.Error(),
		)
	}
}

func TestValidateMigrationsDirRejectsMissingDirectory(
	t *testing.T,
) {
	missingDir := filepath.Join(
		t.TempDir(),
		"missing",
	)

	resolved, err := validateMigrationsDir(
		missingDir,
	)

	if err == nil {
		t.Fatal(
			"expected missing migrations directory to be rejected",
		)
	}

	if resolved != "" {
		t.Fatalf(
			"expected empty resolved path, got %q",
			resolved,
		)
	}

	if !strings.Contains(
		err.Error(),
		"stat migrations directory",
	) {
		t.Fatalf(
			"expected stat error context, got %q",
			err.Error(),
		)
	}
}

func TestValidateMigrationsDirRejectsFile(
	t *testing.T,
) {
	tempDir := t.TempDir()
	filePath := filepath.Join(
		tempDir,
		"not-a-directory.sql",
	)

	if err := os.WriteFile(
		filePath,
		[]byte("SELECT 1;"),
		0o600,
	); err != nil {
		t.Fatalf(
			"write test migration file: %v",
			err,
		)
	}

	resolved, err := validateMigrationsDir(
		filePath,
	)

	if err == nil {
		t.Fatal(
			"expected migration file path to be rejected",
		)
	}

	if resolved != "" {
		t.Fatalf(
			"expected empty resolved path, got %q",
			resolved,
		)
	}

	if !strings.Contains(
		err.Error(),
		"is not a directory",
	) {
		t.Fatalf(
			"expected directory type error, got %q",
			err.Error(),
		)
	}
}
