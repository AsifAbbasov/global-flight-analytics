package migrationrepair

import (
	"errors"
	"testing"
)

func TestNewPostgresInspectorRequiresPool(t *testing.T) {
	_, err := NewPostgresInspector(nil)
	if !errors.Is(err, ErrPostgresPoolRequired) {
		t.Fatalf(
			"NewPostgresInspector() error = %v, want %v",
			err,
			ErrPostgresPoolRequired,
		)
	}
}

func TestVersionRemainsStable(t *testing.T) {
	if Version !=
		"migration-sequence-repair-preflight-v1" {
		t.Fatalf("Version = %q", Version)
	}
}
