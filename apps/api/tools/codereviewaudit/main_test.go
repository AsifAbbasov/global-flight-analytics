package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditAcceptsAuthoritativePolicyFixture(t *testing.T) {
	root := createPolicyFixture(t)
	result := auditCodeReviewPolicy(root)
	if len(result.violations) != 0 {
		t.Fatalf("policy fixture violations = %v", result.violations)
	}
}

func TestAuditRejectsMechanicalReviewRules(t *testing.T) {
	root := createPolicyFixture(t)
	writeFixture(
		t,
		root,
		"docs/82_CODE_REVIEW_STANDARD.md",
		"### Blocker\n### Required change\n### Suggestion\n### Nit\n"+
			"**Location** **Evidence** **Risk** **Severity** **Required change** **Verification**\n"+
			"the exact commit or diff that was reviewed\nwhich checks were not executed\nvalid only for the reviewed commit\n",
	)
	result := auditCodeReviewPolicy(root)
	if len(result.violations) == 0 {
		t.Fatal("mechanical-rule policy regression was not rejected")
	}
}

func TestAuditRejectsUnclassifiedPullRequestTemplate(t *testing.T) {
	root := createPolicyFixture(t)
	writeFixture(t, root, ".github/pull_request_template.md", "## Change summary\n")
	result := auditCodeReviewPolicy(root)
	if len(result.violations) == 0 {
		t.Fatal("unclassified pull request template was not rejected")
	}
}

func TestAuditRejectsMechanicalPostgreSQLWording(t *testing.T) {
	root := createPolicyFixture(t)
	writeFixture(
		t,
		root,
		"apps/api/tools/postgreslayeraudit/main.go",
		"package main\nconst message = \"legacy trajectory method identifier with an overloaded And contract remains\"\n",
	)
	result := auditCodeReviewPolicy(root)
	if len(result.violations) == 0 {
		t.Fatal("mechanical PostgreSQL audit wording was not rejected")
	}
}

func createPolicyFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"apps/api/go.mod": "module fixture\n",
		"docs/82_CODE_REVIEW_STANDARD.md": strings.Join([]string{
			"### Blocker",
			"### Required change",
			"### Suggestion",
			"### Nit",
			"**Location** **Evidence** **Risk** **Severity** **Required change** **Verification**",
			"Function length is a review signal, not a verdict.",
			"Words such as `And` and `With` are not globally forbidden.",
			"`nil` is not globally forbidden.",
			"diagnostic lenses, not standalone evidence",
			"the exact commit or diff that was reviewed",
			"which checks were not executed",
			"valid only for the reviewed commit",
		}, "\n"),
		"docs/27_ENGINEERING_PRINCIPLES.md": "<!-- CODE-REVIEW-STANDARD-V1 -->\ndocs/82_CODE_REVIEW_STANDARD.md\n",
		"docs/DOCUMENT_INDEX.md":            "<!-- CODE-REVIEW-STANDARD-V1:DOCUMENT-INDEX -->\n82_CODE_REVIEW_STANDARD.md\n",
		".github/pull_request_template.md": strings.Join([]string{
			"## Risk and invariants",
			"## Evidence",
			"## Verification",
			"## Reviewer classification",
			"Location, Evidence, Risk, Required change, and Verification",
		}, "\n"),
		"apps/api/tools/postgreslayeraudit/main.go": "package main\nconst message = \"legacy ambiguous trajectory method identifier remains after the intent-oriented rename\"\n",
		".github/workflows/backend-ci.yml":          "- name: Run code review policy audit\n  run: go run ./tools/codereviewaudit -strict\n",
	}
	for path, content := range files {
		writeFixture(t, root, path, content)
	}
	return root
}

func writeFixture(t *testing.T, root string, relative string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
