package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectRequirementsAcceptsRequiredFragments(t *testing.T) {
	root := t.TempDir()
	writeAuditFixture(t, root, "example.go", "package example\n\nconst alpha = \"beta\"\n")
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

func TestInspectRequirementsNormalizesGofmtLineWrapping(t *testing.T) {
	root := t.TempDir()
	writeAuditFixture(
		t,
		root,
		"example.go",
		"package example\n\nfunc example() {\n\twriteFingerprintString(\n\t\tdigest,\n\t\tFingerprintVersion,\n\t)\n}\n",
	)
	failures := inspectRequirements(root, []requirement{
		{
			path: "example.go",
			fragments: []string{
				"writeFingerprintString(digest, FingerprintVersion)",
			},
		},
	})
	if len(failures) != 0 {
		t.Fatalf("gofmt-wrapped requirement failed: %#v", failures)
	}
}

func TestInspectRequirementsScopesFragmentsToNamedFunction(t *testing.T) {
	root := t.TempDir()
	writeAuditFixture(
		t,
		root,
		"fixtures_test.go",
		`package example

type result struct {
	Score float64
}

func unrelatedFixture() result {
	return result{Score: 0.85}
}

func validProductionFrequency() result {
	return result{Score: 0.94}
}
`,
	)
	requirements := []requirement{
		{
			path:      "fixtures_test.go",
			function:  "validProductionFrequency",
			fragments: []string{"Score: 0.94"},
			forbidden: []string{"Score: 0.85"},
		},
	}
	if failures := inspectRequirements(root, requirements); len(failures) != 0 {
		t.Fatalf("unrelated fixture caused scoped audit failure: %#v", failures)
	}
}

func TestInspectRequirementsRejectsWrongScoreInsideNamedFunction(t *testing.T) {
	root := t.TempDir()
	writeAuditFixture(
		t,
		root,
		"fixtures_test.go",
		`package example

type result struct {
	Score float64
}

func validProductionFrequency() result {
	return result{Score: 0.85}
}
`,
	)
	failures := inspectRequirements(root, []requirement{
		{
			path:      "fixtures_test.go",
			function:  "validProductionFrequency",
			fragments: []string{"Score: 0.94"},
			forbidden: []string{"Score: 0.85"},
		},
	})
	joined := strings.Join(failures, "\n")
	if !strings.Contains(joined, "missing required fragment") ||
		!strings.Contains(joined, "contains forbidden fragment") {
		t.Fatalf("scoped audit did not reject the wrong score: %s", joined)
	}
}

func TestInspectRequirementsReportsMissingForbiddenAndInvalidGo(t *testing.T) {
	root := t.TempDir()
	writeAuditFixture(t, root, "example.go", "package example\nlegacy\n")
	failures := inspectRequirements(root, []requirement{
		{
			path:      "example.go",
			fragments: []string{"alpha"},
			forbidden: []string{"legacy"},
		},
	})
	if len(failures) == 0 {
		t.Fatal("audit accepted invalid and incomplete Go source")
	}
	joined := strings.Join(failures, "\n")
	if !strings.Contains(joined, "inspect example.go: parse Go source") {
		t.Fatalf("unexpected failures: %s", joined)
	}
}

func TestReviewRequirementsAcceptSyntheticProtectedTree(t *testing.T) {
	repositoryRoot := t.TempDir()
	apiRoot := filepath.Join(repositoryRoot, "apps", "api")
	for _, item := range reviewRequirements() {
		path := filepath.Clean(filepath.Join(apiRoot, filepath.FromSlash(item.path)))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create synthetic requirement directory: %v", err)
		}
		var content string
		if filepath.Ext(path) == ".go" {
			if item.function == "validProductionFrequency" {
				content = `package synthetic

type result struct {
	Score float64
}

func unrelatedFixture() result {
	return result{Score: 0.85}
}

func validProductionFrequency() result {
	return result{Score: 0.94}
}
`
			} else {
				content = "package synthetic\n\n/*\n" + strings.Join(item.fragments, "\n") + "\n*/\n"
			}
		} else {
			content = strings.Join(item.fragments, "\n") + "\n"
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write synthetic requirement %s: %v", item.path, err)
		}
	}
	if failures := inspectRequirements(apiRoot, reviewRequirements()); len(failures) != 0 {
		t.Fatalf("review requirements reject their own protected manifest: %#v", failures)
	}
}

func TestReviewRequirementsProtectEvidenceIsolationAndPolicyIntegrity(t *testing.T) {
	protected := map[string]bool{
		"internal/projectionintelligence/projectionroutefrequency/model.go":                       false,
		"internal/projectionintelligence/projectionroutefrequency/config.go":                      false,
		"internal/projectionintelligence/projectionroutefrequency/evaluator.go":                   false,
		"internal/projectionintelligence/projectionroutefrequency/fingerprint.go":                 false,
		"internal/projectionintelligence/projectionread/postgres_queries.go":                      false,
		"internal/projectionintelligence/projectionread/policy.go":                                false,
		"internal/projectionintelligence/projectionroutefrequency/decision_integrity_test.go":     false,
		"internal/projectionintelligence/projectionread/route_frequency_policy_integrity_test.go": false,
		"internal/projectionintelligence/projectionproduction/fixtures_test.go":                   false,
	}
	markProtectedRequirements(t, protected)
}

func TestReviewRequirementsProtectDocumentationAndContinuousIntegration(t *testing.T) {
	protected := map[string]bool{
		"../../docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md": false,
		"../../docs/140_PROJECTION_ROUTE_FREQUENCY_REVIEW_HARDENING.md":                false,
		"../../docs/DOCUMENT_INDEX.md":                                                 false,
		"../../.github/workflows/backend-ci.yml":                                       false,
	}
	markProtectedRequirements(t, protected)
}

func TestReviewRequirementsKeepClosureOpenUntilExactEvidence(t *testing.T) {
	var review requirement
	found := false
	for _, item := range reviewRequirements() {
		if item.path == "../../docs/140_PROJECTION_ROUTE_FREQUENCY_REVIEW_HARDENING.md" {
			review = item
			found = true
			break
		}
	}
	if !found {
		t.Fatal("authoritative review document is not protected")
	}
	required := strings.Join(review.fragments, "\n")
	for _, fragment := range []string{
		"POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=PENDING",
		"PERMANENT_AUDIT_COMMIT=PENDING",
		"OPEN_CONFIRMED_FINDINGS=3",
		"PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=OPEN",
	} {
		if !strings.Contains(required, fragment) {
			t.Fatalf("open-closure requirement misses %q", fragment)
		}
	}
	forbidden := strings.Join(review.forbidden, "\n")
	for _, fragment := range []string{
		"Status: closed",
		"OPEN_CONFIRMED_FINDINGS=0",
		"PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=CLOSED",
	} {
		if !strings.Contains(forbidden, fragment) {
			t.Fatalf("open-closure requirement does not forbid %q", fragment)
		}
	}
}

func TestReviewRequirementsScopesProductionFixtureScore(t *testing.T) {
	for _, item := range reviewRequirements() {
		if item.path != "internal/projectionintelligence/projectionproduction/fixtures_test.go" {
			continue
		}
		if item.function != "validProductionFrequency" {
			t.Fatalf("fixture requirement function = %q", item.function)
		}
		required := strings.Join(item.fragments, "\n")
		forbidden := strings.Join(item.forbidden, "\n")
		if !strings.Contains(required, "Score: 0.94") {
			t.Fatal("fixture requirement does not protect weighted score 0.94")
		}
		if !strings.Contains(forbidden, "Score: 0.85") {
			t.Fatal("fixture requirement does not reject stale score 0.85 in the protected function")
		}
		return
	}
	t.Fatal("production fixture requirement is missing")
}

func TestReviewRequirementsRejectLegacyWeightAndEvidenceContracts(t *testing.T) {
	checks := map[string]struct {
		required  []string
		forbidden []string
	}{
		"internal/projectionintelligence/projectionroutefrequency/model.go": {
			required:  []string{"component.Weight <= 0"},
			forbidden: []string{"component.Weight < 0"},
		},
		"internal/projectionintelligence/projectionroutefrequency/config.go": {
			required: []string{
				"route-frequency component weights must be finite, positive, and sum to one",
				"weight <= 0",
			},
			forbidden: []string{
				"route-frequency component weights must be finite, non-negative, and sum to one",
				"weight < 0",
			},
		},
		"internal/projectionintelligence/projectionread/postgres_queries.go": {
			required: []string{
				"latest_route_per_evidence AS (",
				"route_result.trajectory_id::text <> $7",
				"trajectory.flight_id::text <> $8",
			},
			forbidden: []string{"FROM latest_route_per_trajectory;"},
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
				t.Fatalf("%s misses required fragment %q", item.path, fragment)
			}
		}
		forbiddenText := strings.Join(item.forbidden, "\n")
		for _, fragment := range check.forbidden {
			if !strings.Contains(forbiddenText, fragment) {
				t.Fatalf("%s misses forbidden fragment %q", item.path, fragment)
			}
		}
		delete(checks, item.path)
	}
	for path := range checks {
		t.Fatalf("review requirement is missing for %s", path)
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
