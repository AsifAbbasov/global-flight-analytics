package migrationaudit

import (
	"errors"
	"testing"
)

func TestPostgresStateLoaderRejectsNilContextBeforePoolAccess(
	t *testing.T,
) {
	loader := &PostgresStateLoader{}

	_, err := loader.Load(nil)

	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf(
			"Load() error = %v, want %v",
			err,
			ErrContextRequired,
		)
	}
}

func TestNewPostgresStateLoaderRequiresPool(t *testing.T) {
	_, err := NewPostgresStateLoader(nil)
	if !errors.Is(err, ErrPostgresPoolRequired) {
		t.Fatalf(
			"NewPostgresStateLoader() error = %v, want %v",
			err,
			ErrPostgresPoolRequired,
		)
	}
}

func TestMigrationAuditVersionRemainsStable(t *testing.T) {
	if Version != "migration-history-audit-v1" {
		t.Fatalf("Version = %q", Version)
	}
}
