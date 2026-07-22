package migrationaudit

import (
	"errors"
	"testing"
)

func TestAuditRejectsNilContext(t *testing.T) {
	auditor := newTestAuditor(
		t,
		Config{
			MigrationsDir: t.TempDir(),
			StateLoader:   fakeStateLoader{},
		},
	)

	_, err := auditor.Audit(nil)
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("Audit(nil) error = %v", err)
	}
}
