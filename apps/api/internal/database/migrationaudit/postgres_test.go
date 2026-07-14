package migrationaudit

import (
	"errors"
	"testing"
)

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
