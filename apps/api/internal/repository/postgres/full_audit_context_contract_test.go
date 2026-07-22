package postgres

import (
	"os"
	"strings"
	"testing"
)

func TestPostgreSQLFullAuditContextContracts(t *testing.T) {
	for _, fileName := range []string{
		"flightstate_reconciliation_repository.go",
		"ingestionrun_repository.go",
		"reconciliation_repository.go",
		"source_http_validator_repository.go",
	} {
		content, err := os.ReadFile(fileName)
		if err != nil {
			t.Fatalf("read %s: %v", fileName, err)
		}
		if strings.Contains(string(content), "ctx = context.Background()") {
			t.Fatalf("%s still replaces nil context with background", fileName)
		}
		if !strings.Contains(string(content), "requireRepositoryContext(ctx)") {
			t.Fatalf("%s does not use the repository context contract", fileName)
		}
	}

	content, err := os.ReadFile("trajectory_read_snapshot.go")
	if err != nil {
		t.Fatalf("read trajectory_read_snapshot.go: %v", err)
	}
	text := string(content)
	if strings.Contains(text, ".Rollback(ctx)") {
		t.Fatal("trajectory read snapshot still rolls back with request context")
	}
	if !strings.Contains(text, ".Rollback(rollbackCtx)") {
		t.Fatal("trajectory read snapshot does not use explicit rollback context")
	}
}
