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

func TestInspectRequirementsReportsMissingAndForbiddenFragments(t *testing.T) {
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

func TestReviewRequirementsProtectFreshnessCorrectnessContracts(t *testing.T) {
	protected := map[string]bool{
		"internal/projectionintelligence/projectionfreshness/model.go":                 false,
		"internal/projectionintelligence/projectionfreshness/config.go":                false,
		"internal/projectionintelligence/projectionfreshness/policy.go":                false,
		"internal/projectionintelligence/projectionfreshness/evaluator.go":             false,
		"internal/projectionintelligence/projectionfreshness/measurement.go":           false,
		"internal/projectionintelligence/projectionfreshness/components.go":            false,
		"internal/projectionintelligence/projectionfreshness/decision.go":              false,
		"internal/projectionintelligence/projectionfreshness/fingerprint.go":           false,
		"internal/projectionintelligence/projectionfreshness/hardening_test.go":        false,
		"internal/projectionintelligence/projectionfreshness/policy_integrity_test.go": false,
	}
	markProtectedRequirements(t, protected)
}

func TestReviewRequirementsProtectProductionAndContinuousIntegration(t *testing.T) {
	protected := map[string]bool{
		"internal/projectionintelligence/projectionproduction/fixtures_test.go":                   false,
		"internal/projectionintelligence/projectionproduction/freshness_fixture_contract_test.go": false,
		"../../.github/workflows/backend-ci.yml":                                                  false,
	}
	markProtectedRequirements(t, protected)
}

func markProtectedRequirements(t *testing.T, protected map[string]bool) {
	t.Helper()
	for _, item := range reviewRequirements() {
		if _, exists := protected[item.path]; exists {
			protected[item.path] = true
		}
	}
	for path, found := range protected {
		if !found {
			t.Fatalf("permanent review requirement is missing for %s", path)
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
