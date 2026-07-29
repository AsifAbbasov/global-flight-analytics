package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectRequirementsAcceptsRequiredFragments(t *testing.T) {
	root := t.TempDir()
	writeAuditFixture(t, root, "example.go", "alpha\nbeta\n")

	failures := inspectRequirements(root, []requirement{
		{
			path:      "example.go",
			fragments: []string{"alpha", "beta"},
			forbidden: []string{"legacy"},
		},
	})
	if len(failures) != 0 {
		t.Fatalf("unexpected audit failures: %#v", failures)
	}
}

func TestInspectRequirementsReportsMissingForbiddenFragments(t *testing.T) {
	root := t.TempDir()
	writeAuditFixture(t, root, "example.go", "alpha\nlegacy\n")

	failures := inspectRequirements(root, []requirement{
		{
			path:      "example.go",
			fragments: []string{"alpha", "beta"},
			forbidden: []string{"legacy"},
		},
	})
	if len(failures) != 2 {
		t.Fatalf("failure count = %d, want 2: %#v", len(failures), failures)
	}
	joined := strings.Join(failures, "\n")
	if !strings.Contains(joined, "missing required fragment") ||
		!strings.Contains(joined, "contains forbidden fragment") {
		t.Fatalf("unexpected failures: %s", joined)
	}
}

func TestReviewRequirementsProtectMandatoryContinuationInterfaces(t *testing.T) {
	requirements := reviewRequirements()
	protected := map[string]bool{
		"internal/projectionintelligence/projectionproduction/config.go":               false,
		"internal/projectionintelligence/projectionproduction/pattern_evaluation.go":   false,
		"internal/projectionintelligence/projectioncontinuation/config.go":             false,
		"internal/projectionintelligence/projectioncontinuation/pattern_evaluation.go": false,
		"../../.github/workflows/backend-ci.yml":                                       false,
	}
	for _, item := range requirements {
		if _, exists := protected[item.path]; exists {
			protected[item.path] = true
		}
	}
	for path, found := range protected {
		if !found {
			t.Fatalf("mandatory review requirement is missing for %s", path)
		}
	}
}

func writeAuditFixture(t *testing.T, root string, relativePath string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
