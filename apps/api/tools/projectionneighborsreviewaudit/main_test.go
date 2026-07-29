package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNormalizeRequirementTextCollapsesWhitespace(t *testing.T) {
	input := "source-attested origin-destination route\n\tscope"
	want := "source-attested origin-destination route scope"
	if got := normalizeRequirementText(input); got != want {
		t.Fatalf("normalizeRequirementText() = %q, want %q", got, want)
	}
}

func TestInspectRequirementsMatchesAcrossLineBreaks(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "document.md")
	content := "formal closure pending permanent\naudit Continuous Integration"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	failures := inspectRequirements(root, []requirement{
		{
			path: "document.md",
			fragments: []string{
				"pending permanent audit Continuous Integration",
			},
		},
	})
	if len(failures) != 0 {
		t.Fatalf("inspectRequirements() failures = %v, want none", failures)
	}
}

func TestInspectRequirementsDetectsForbiddenAcrossWhitespace(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "document.md")
	content := "PERMANENT_AUDIT_COMMIT=\nPENDING"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	failures := inspectRequirements(root, []requirement{
		{
			path: "document.md",
			forbidden: []string{
				"PERMANENT_AUDIT_COMMIT= PENDING",
			},
		},
	})
	want := []string{
		`document.md retains forbidden "PERMANENT_AUDIT_COMMIT= PENDING"`,
	}
	if !reflect.DeepEqual(failures, want) {
		t.Fatalf("inspectRequirements() failures = %v, want %v", failures, want)
	}
}
