package migrator

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMigratorPublicOperationsRejectNilContext(t *testing.T) {
	runner := &Runner{}
	tests := []struct {
		name string
		run  func() error
	}{
		{name: "EnsureSchemaMigrations", run: func() error {
			return runner.EnsureSchemaMigrations(nil)
		}},
		{name: "Status", run: func() error {
			_, err := runner.Status(nil)
			return err
		}},
		{name: "ApplyPending", run: func() error {
			_, err := runner.ApplyPending(nil)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			if !errors.Is(err, ErrMigrationContextRequired) {
				t.Fatalf("error = %v, want %v", err, ErrMigrationContextRequired)
			}
		})
	}
}

func TestWithMigrationLockRejectsNilContextBeforePoolAccess(t *testing.T) {
	runner := &Runner{}
	operationCalled := false
	err := runner.withMigrationLock(nil, func(_ *pgxpool.Conn) error {
		operationCalled = true
		return nil
	})
	if !errors.Is(err, ErrMigrationContextRequired) {
		t.Fatalf("error = %v, want %v", err, ErrMigrationContextRequired)
	}
	if operationCalled {
		t.Fatal("operation must not run for a nil context")
	}
}

func TestMigratorContextSourceContract(t *testing.T) {
	sourceBytes, err := os.ReadFile("runner.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, required := range []string{
		"ErrMigrationContextRequired",
		"func requireMigrationContext(",
		"func (runner *Runner) EnsureSchemaMigrations(",
		"func (runner *Runner) Status(",
		"func (runner *Runner) ApplyPending(",
		"withMigrationLock(",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("runner.go is missing %q", required)
		}
	}
	if strings.Contains(source, "ctx = context.Background()") {
		t.Fatal("runner.go must not replace a nil caller context")
	}
}

func TestMigratorCleanupContextsRemainIndependentAndBounded(t *testing.T) {
	sourceBytes, err := os.ReadFile("runner.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	for _, required := range []string{
		"func releaseMigrationLock(",
		"func destroyLockedConnection(",
		"func rollbackMigrationTransaction(",
		"context.WithTimeout(",
		"context.Background()",
		"migrationLockReleaseTimeout",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("cleanup contract is missing %q", required)
		}
	}
}
