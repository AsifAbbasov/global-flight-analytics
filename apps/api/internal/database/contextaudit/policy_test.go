package contextaudit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditDirectoryAcceptsExactMigratorPolicy(t *testing.T) {
	directory := writeMigratorFixture(t, "")
	violations, err := AuditDirectory(directory, MigratorPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("violations = %#v", violations)
	}
}

func TestAuditDirectoryRejectsContextSourceVariants(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "background alias", source: "package migrator\nimport ctxpkg \"context\"\nfunc hidden() ctxpkg.Context { return ctxpkg.Background() }\n", want: "Background"},
		{name: "todo dot import", source: "package migrator\nimport . \"context\"\nfunc hidden() Context { return TODO() }\n", want: "TODO"},
		{name: "without cancel", source: "package migrator\nimport \"context\"\nfunc hidden(ctx context.Context) context.Context { return context.WithoutCancel(ctx) }\n", want: "WithoutCancel"},
		{name: "function value", source: "package migrator\nimport \"context\"\nvar hidden = context.Background\n", want: "function value"},
		{name: "parameter reassignment", source: "package migrator\nimport \"context\"\nfunc (runner *Runner) Status(ctx context.Context) ([]int, error) { if err := requireMigrationContext(ctx); err != nil { return nil, err }; ctx = context.WithValue(ctx, \"key\", \"value\"); return nil, nil }\n", want: "must not be reassigned"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := writeMigratorFixture(t, test.source)
			violations, err := AuditDirectory(directory, MigratorPolicy())
			if err != nil {
				t.Fatal(err)
			}
			if !violationsContain(violations, test.want) {
				t.Fatalf("violations = %#v, want text %q", violations, test.want)
			}
		})
	}
}

func TestAuditDirectoryIgnoresCommentsAndStrings(t *testing.T) {
	directory := writeMigratorFixture(t, "package migrator\nconst note = `context.Background() and context.TODO()`\n// context.WithoutCancel(ctx)\n")
	violations, err := AuditDirectory(directory, MigratorPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("violations = %#v", violations)
	}
}

func writeMigratorFixture(t *testing.T, additional string) string {
	t.Helper()
	directory := t.TempDir()
	base := `package migrator
import (
    "context"
    "time"
)
var ErrMigrationContextRequired = error(nil)
const migrationLockReleaseTimeout = time.Second
type Runner struct{}
func requireMigrationContext(ctx context.Context) error { return nil }
func (runner *Runner) EnsureSchemaMigrations(ctx context.Context) error { return requireMigrationContext(ctx) }
func (runner *Runner) ensureSchemaMigrations(ctx context.Context) error { return requireMigrationContext(ctx) }
func (runner *Runner) Status(ctx context.Context) ([]int, error) { if err := requireMigrationContext(ctx); err != nil { return nil, err }; return nil, nil }
func (runner *Runner) ApplyPending(ctx context.Context) ([]int, error) { if err := requireMigrationContext(ctx); err != nil { return nil, err }; return nil, nil }
func (runner *Runner) applyMigrationAtomically(ctx context.Context) error { return requireMigrationContext(ctx) }
func (runner *Runner) withMigrationLock(ctx context.Context) error { return requireMigrationContext(ctx) }
func (runner *Runner) appliedMigrations(ctx context.Context) error { return requireMigrationContext(ctx) }
func (runner *Runner) appliedMigrationsWith(ctx context.Context) error { return requireMigrationContext(ctx) }
func releaseMigrationLock() { ctx, cancel := context.WithTimeout(context.Background(), migrationLockReleaseTimeout); defer cancel(); _ = ctx }
func destroyLockedConnection() { ctx, cancel := context.WithTimeout(context.Background(), migrationLockReleaseTimeout); defer cancel(); _ = ctx }
func rollbackMigrationTransaction() { ctx, cancel := context.WithTimeout(context.Background(), migrationLockReleaseTimeout); defer cancel(); _ = ctx }
`
	if err := os.WriteFile(filepath.Join(directory, "runner.go"), []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}
	if additional != "" {
		if err := os.WriteFile(filepath.Join(directory, "additional.go"), []byte(additional), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func violationsContain(violations []Violation, text string) bool {
	for _, violation := range violations {
		if strings.Contains(violation.String(), text) {
			return true
		}
	}
	return false
}
