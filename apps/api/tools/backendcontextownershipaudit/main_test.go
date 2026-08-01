package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeParsedFileClassifiesCallerContextReplacement(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		findings int
		root     string
	}{
		{
			name: "background replacement",
			source: `package sample
import "context"
func run(ctx context.Context) {
	ctx = context.Background()
	_ = ctx
}`,
			findings: 1,
			root:     "Background",
		},
		{
			name: "todo replacement through alias",
			source: `package sample
import lifecycle "context"
func run(ctx lifecycle.Context) {
	ctx = lifecycle.TODO()
	_ = ctx
}`,
			findings: 1,
			root:     "TODO",
		},
		{
			name: "nested root replacement",
			source: `package sample
import "context"
func run(ctx context.Context) {
	ctx = context.WithValue(context.Background(), "key", "value")
	_ = ctx
}`,
			findings: 1,
			root:     "Background",
		},
		{
			name: "caller context preserved",
			source: `package sample
import "context"
func run(ctx context.Context) error {
	return ctx.Err()
}`,
			findings: 0,
		},
		{
			name: "independent local root",
			source: `package sample
import "context"
func run() {
	ctx := context.Background()
	_ = ctx
}`,
			findings: 0,
		},
		{
			name: "nested parameter is inspected once",
			source: `package sample
import "context"
func run(ctx context.Context) {
	_ = func(ctx context.Context) {
		ctx = context.Background()
		_ = ctx
	}
	_ = ctx
}`,
			findings: 1,
			root:     "Background",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSet := token.NewFileSet()
			parsed, err := parser.ParseFile(
				fileSet,
				"sample.go",
				test.source,
				parser.SkipObjectResolution,
			)
			if err != nil {
				t.Fatalf("parse fixture: %v", err)
			}

			report := auditReport{}
			analyzeParsedFile(fileSet, parsed, "sample.go", &report)

			if len(report.Findings) != test.findings {
				t.Fatalf("findings = %#v, want %d", report.Findings, test.findings)
			}
			if test.findings > 0 && report.Findings[0].Root != test.root {
				t.Fatalf("root = %q, want %q", report.Findings[0].Root, test.root)
			}
		})
	}
}

func TestRunStrictRejectsFinding(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(root, "go.mod"),
		[]byte("module auditfixture\n\ngo 1.23\n"),
		0o600,
	); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	internal := filepath.Join(root, "internal", "fixture")
	if err := os.MkdirAll(internal, 0o700); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(internal, "fixture.go"),
		[]byte(`package fixture
import "context"
func Run(ctx context.Context) {
	ctx = context.Background()
	_ = ctx
}`),
		0o600,
	); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var stdout strings.Builder
	var stderr strings.Builder
	exitCode := run(
		&stdout,
		&stderr,
		[]string{"-root", root, "-strict"},
	)

	if exitCode != 1 {
		t.Fatalf(
			"exit code = %d, stdout=%q stderr=%q",
			exitCode,
			stdout.String(),
			stderr.String(),
		)
	}
	if !strings.Contains(
		stdout.String(),
		"Backend caller context ownership audit: FAIL",
	) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRepositoryHasNoCallerContextReplacement(t *testing.T) {
	root := moduleRoot(t)
	report, err := scanRepository(root)
	if err != nil {
		t.Fatalf("scan repository: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("caller context replacements remain: %#v", report.Findings)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	current, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("Go module root was not found")
		}
		current = parent
	}
}
