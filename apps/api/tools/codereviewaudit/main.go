package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type auditResult struct {
	checks     []string
	violations []string
}

func main() {
	strict := flag.Bool("strict", false, "exit non-zero when a violation is found")
	root := flag.String("root", "", "repository root; auto-detected when empty")
	flag.Parse()

	repositoryRoot, err := resolveRepositoryRoot(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	result := auditCodeReviewPolicy(repositoryRoot)
	for _, check := range result.checks {
		fmt.Println(check + ": PASS")
	}
	if len(result.violations) == 0 {
		fmt.Println("Code review policy: PASS")
		return
	}
	for _, violation := range result.violations {
		fmt.Fprintln(os.Stderr, "VIOLATION:", violation)
	}
	if *strict {
		os.Exit(1)
	}
}

func resolveRepositoryRoot(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return filepath.Abs(explicit)
	}
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fileExists(filepath.Join(current, "apps", "api", "go.mod")) &&
			fileExists(filepath.Join(current, "docs", "DOCUMENT_INDEX.md")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("repository root was not found from %s", current)
		}
		current = parent
	}
}

func auditCodeReviewPolicy(root string) auditResult {
	result := auditResult{}
	check := func(name string, condition bool, message string) {
		if condition {
			result.checks = append(result.checks, name)
			return
		}
		result.violations = append(result.violations, message)
	}

	standard := readText(root, "docs/82_CODE_REVIEW_STANDARD.md", &result)
	engineering := readText(root, "docs/27_ENGINEERING_PRINCIPLES.md", &result)
	index := readText(root, "docs/DOCUMENT_INDEX.md", &result)
	pullRequestTemplate := readText(root, ".github/pull_request_template.md", &result)
	postgresAudit := readText(root, "apps/api/tools/postgreslayeraudit/main.go", &result)
	backendWorkflow := readText(root, ".github/workflows/backend-ci.yml", &result)

	check(
		"Severity classification",
		containsAll(standard, "### Blocker", "### Required change", "### Suggestion", "### Nit"),
		"code review standard does not define the four authoritative finding classifications",
	)
	check(
		"Finding evidence contract",
		containsAll(standard, "**Location**", "**Evidence**", "**Risk**", "**Severity**", "**Required change**", "**Verification**"),
		"blocking findings are not required to provide location, evidence, risk, correction and verification",
	)
	check(
		"Mechanical-rule rejection",
		containsAll(
			standard,
			"Function length is a review signal, not a verdict.",
			"Words such as `And` and `With` are not globally forbidden.",
			"`nil` is not globally forbidden.",
			"diagnostic lenses, not standalone evidence",
		),
		"code review policy still permits line count, vocabulary, nil, or principle labels to act as standalone defects",
	)
	check(
		"Audit scope disclosure",
		containsAll(standard, "the exact commit or diff that was reviewed", "which checks were not executed", "valid only for the reviewed commit"),
		"review summaries are not required to disclose commit scope and unexecuted checks",
	)
	check(
		"Engineering baseline linkage",
		strings.Contains(engineering, "<!-- CODE-REVIEW-STANDARD-V1 -->") &&
			strings.Contains(engineering, "docs/82_CODE_REVIEW_STANDARD.md"),
		"engineering principles do not delegate review interpretation to the authoritative code review standard",
	)
	check(
		"Documentation register",
		strings.Contains(index, "82_CODE_REVIEW_STANDARD.md") &&
			strings.Contains(index, "<!-- CODE-REVIEW-STANDARD-V1:DOCUMENT-INDEX -->"),
		"documentation index does not register the authoritative code review standard",
	)
	check(
		"Pull request evidence template",
		containsAll(
			pullRequestTemplate,
			"## Risk and invariants",
			"## Evidence",
			"## Verification",
			"## Reviewer classification",
			"Location, Evidence, Risk, Required change, and Verification",
		),
		"pull request template does not require risk, evidence, verification and classified review findings",
	)
	check(
		"PostgreSQL audit wording",
		!strings.Contains(postgresAudit, "overloaded And contract") &&
			strings.Contains(postgresAudit, "legacy ambiguous trajectory method identifier remains after the intent-oriented rename"),
		"PostgreSQL audit still describes a specific legacy name as a repository-wide And-word violation",
	)
	check(
		"Continuous integration reachability",
		strings.Contains(backendWorkflow, "Run code review policy audit") &&
			strings.Contains(backendWorkflow, "go run ./tools/codereviewaudit -strict"),
		"backend continuous integration does not execute the code review policy audit",
	)

	return result
}

func containsAll(value string, fragments ...string) bool {
	for _, fragment := range fragments {
		if !strings.Contains(value, fragment) {
			return false
		}
	}
	return true
}

func readText(root string, relative string, result *auditResult) string {
	path := filepath.Join(root, filepath.FromSlash(relative))
	content, err := os.ReadFile(path)
	if err != nil {
		result.violations = append(result.violations, fmt.Sprintf("read %s: %v", relative, err))
		return ""
	}
	return string(content)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
