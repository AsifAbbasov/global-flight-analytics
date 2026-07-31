package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectRequirementsFailsClosed(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "contract.go")
	if err := os.WriteFile(path, []byte("package contract\nconst marker = \"present\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	requirements := []requirement{{
		path:      "contract.go",
		fragments: []string{"marker = \"present\"", "required-postcondition"},
		forbidden: []string{"forbidden-mutation"},
	}}
	failures := inspectRequirements(root, requirements)
	if len(failures) != 1 {
		t.Fatalf("failures = %#v, want one missing requirement", failures)
	}
	if err := os.WriteFile(path, []byte("required-postcondition\nforbidden-mutation\nmarker = \"present\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	failures = inspectRequirements(root, requirements)
	if len(failures) != 1 {
		t.Fatalf("failures = %#v, want one forbidden requirement", failures)
	}
}
