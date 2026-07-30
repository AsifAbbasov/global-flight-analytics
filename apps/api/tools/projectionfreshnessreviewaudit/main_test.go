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
		"internal/projectionintelligence/projectionfreshness/config_test.go":           false,
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

func TestReviewRequirementsProtectFormalClosureEvidence(t *testing.T) {
	protected := map[string]bool{
		"../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md": false,
		"../../docs/139_PROJECTION_FRESHNESS_REVIEW_HARDENING.md":                      false,
		"../../docs/DOCUMENT_INDEX.md":                                                 false,
		"../../.github/workflows/backend-ci.yml":                                       false,
	}
	markProtectedRequirements(t, protected)
}

func TestReviewRequirementsProtectStrictPositiveWeightPolicy(t *testing.T) {
	checks := map[string]struct {
		required  []string
		forbidden []string
	}{
		"internal/projectionintelligence/projectionfreshness/config.go": {
			required: []string{
				"freshness component weights must be finite, positive, and sum to one",
				"if !finite(weight) || weight <= 0 {",
			},
			forbidden: []string{
				"freshness component weights must be finite, non-negative, and sum to one",
				"if !finite(weight) || weight < 0 {",
			},
		},
		"internal/projectionintelligence/projectionfreshness/model.go": {
			required:  []string{"component.Weight <= 0"},
			forbidden: []string{"component.Weight < 0"},
		},
		"internal/projectionintelligence/projectionfreshness/config_test.go": {
			required: []string{"TestConfigValidateRejectsZeroComponentWeights"},
		},
		"internal/projectionintelligence/projectionfreshness/policy_integrity_test.go": {
			required: []string{"TestResultValidateRejectsZeroComponentWeight"},
		},
	}

	for _, item := range reviewRequirements() {
		check, exists := checks[item.path]
		if !exists {
			continue
		}
		requiredText := strings.Join(item.fragments, "\n")
		for _, fragment := range check.required {
			if !strings.Contains(requiredText, fragment) {
				t.Fatalf(
					"%s misses required positive-weight fragment %q",
					item.path,
					fragment,
				)
			}
		}
		forbiddenText := strings.Join(item.forbidden, "\n")
		for _, fragment := range check.forbidden {
			if !strings.Contains(forbiddenText, fragment) {
				t.Fatalf(
					"%s misses forbidden legacy-weight fragment %q",
					item.path,
					fragment,
				)
			}
		}
		delete(checks, item.path)
	}
	for path := range checks {
		t.Fatalf(
			"positive-weight audit requirement is missing for %s",
			path,
		)
	}
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
