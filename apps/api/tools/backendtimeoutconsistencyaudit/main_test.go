package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestScanRepositoryAcceptsBoundedContextualHTTP(t *testing.T) {
	root := newAuditFixture(t)
	writeAuditFixtureFile(
		t,
		root,
		"internal/provider/client.go",
		`package provider

import (
    "context"
    "net/http"
    "time"
)

func load(ctx context.Context) error {
    client := &http.Client{Timeout: 5 * time.Second}
    request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
    if err != nil { return err }
    response, err := client.Do(request)
    if response != nil { _ = response.Body.Close() }
    return err
}
`,
	)
	writeAuditFixtureFile(
		t,
		root,
		"internal/http/handlers/sample.go",
		`package handlers

import "context"

type requestContext interface { UserContext() context.Context }

func use(c requestContext) context.Context { return c.UserContext() }
`,
	)

	report, err := scanRepository(root)
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected no findings, got %+v", report.Findings)
	}
	if report.ContextualRequests != 1 {
		t.Fatalf("expected one contextual request, got %d", report.ContextualRequests)
	}
	if report.BoundedHTTPClientLiterals != 1 {
		t.Fatalf("expected one bounded client literal, got %d", report.BoundedHTTPClientLiterals)
	}
	if report.HandlerUserContexts != 1 {
		t.Fatalf("expected one handler user context, got %d", report.HandlerUserContexts)
	}
}

func TestScanRepositoryRejectsUnboundedHTTPAndTransportHandlerContext(t *testing.T) {
	root := newAuditFixture(t)
	writeAuditFixtureFile(
		t,
		root,
		"internal/provider/client.go",
		`package provider

import "net/http"

func load() {
    _ = &http.Client{}
    _, _ = http.Get("https://example.com")
    _, _ = http.NewRequest(http.MethodGet, "https://example.com", nil)
    _ = http.DefaultClient
}
`,
	)
	writeAuditFixtureFile(
		t,
		root,
		"internal/http/handlers/sample.go",
		`package handlers

type transportContext interface { Context() any }

func use(c transportContext) any { return c.Context() }
`,
	)

	report, err := scanRepository(root)
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}

	for _, rule := range []string{
		"fiber-transport-context",
		"http-client-timeout-missing",
		"http-package-shortcut",
		"http-request-without-context",
		"http-default-global",
	} {
		if !hasFindingRule(report, rule) {
			t.Fatalf("expected finding rule %q, got %+v", rule, report.Findings)
		}
	}
}

func TestScanRepositoryRejectsExplicitNonPositiveClientTimeout(t *testing.T) {
	root := newAuditFixture(t)
	writeAuditFixtureFile(
		t,
		root,
		"cmd/tool/main.go",
		`package main

import "net/http"

var client = &http.Client{Timeout: 0}
`,
	)

	report, err := scanRepository(root)
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	if !hasFindingRule(report, "http-client-timeout-nonpositive") {
		t.Fatalf("expected nonpositive timeout finding, got %+v", report.Findings)
	}
}

func TestScanRepositoryIgnoresTestFiles(t *testing.T) {
	root := newAuditFixture(t)
	writeAuditFixtureFile(
		t,
		root,
		"internal/provider/client_test.go",
		`package provider

import "net/http"

var client = &http.Client{}
`,
	)

	report, err := scanRepository(root)
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected test files to be excluded, got %+v", report.Findings)
	}
}

func TestRunStrictFailsWhenFindingExists(t *testing.T) {
	root := newAuditFixture(t)
	writeAuditFixtureFile(
		t,
		root,
		"internal/provider/client.go",
		`package provider

import "net/http"

var client = &http.Client{}
`,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	status := run(
		&stdout,
		&stderr,
		[]string{"-root", root, "-strict"},
	)
	if status != 1 {
		t.Fatalf("expected strict failure status 1, got %d stderr=%s", status, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("Backend timeout consistency audit: FAIL")) {
		t.Fatalf("expected failure output, got %s", stdout.String())
	}
}

func hasFindingRule(report auditReport, rule string) bool {
	for _, item := range report.Findings {
		if item.Rule == rule {
			return true
		}
	}
	return false
}

func newAuditFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeAuditFixtureFile(
		t,
		root,
		"go.mod",
		"module example.com/timeoutaudit\n\ngo 1.23\n",
	)
	return root
}

func writeAuditFixtureFile(
	t *testing.T,
	root string,
	relativePath string,
	content string,
) {
	t.Helper()
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
}
