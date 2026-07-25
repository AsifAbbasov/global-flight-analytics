package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRootAcceptsExplicitRoot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "apps", "api")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(path, "go.mod"),
		[]byte("module example.test/project\n"),
		0o644,
	); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := resolveRoot(root)
	if err != nil {
		t.Fatalf("resolveRoot() error = %v", err)
	}
	if got != root {
		t.Fatalf("resolveRoot() = %q, want %q", got, root)
	}
}
