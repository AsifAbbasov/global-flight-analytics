package migrator

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database/contextaudit"
)

func TestMigratorContextASTContract(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migrator context contract test path")
	}

	violations, err := contextaudit.AuditDirectory(
		filepath.Dir(currentFile),
		contextaudit.MigratorPolicy(),
	)
	if err != nil {
		t.Fatalf("audit migrator context policy: %v", err)
	}
	if len(violations) == 0 {
		return
	}

	var details strings.Builder
	for _, violation := range violations {
		details.WriteString("\n- ")
		details.WriteString(violation.String())
	}
	t.Fatalf("migrator context syntax policy violations:%s", details.String())
}
