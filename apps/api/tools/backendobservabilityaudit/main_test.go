package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAuditRejectsMissingObservabilityContract(
	t *testing.T,
) {
	root := t.TempDir()
	if err := os.MkdirAll(
		filepath.Join(root, "apps", "api"),
		0o755,
	); err != nil {
		t.Fatalf("create API directory: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(root, "apps", "api", "go.mod"),
		[]byte("module example.test/audit\n"),
		0o644,
	); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	findings := runAudit(root)
	if len(findings) == 0 {
		t.Fatal("expected missing observability findings")
	}
}
