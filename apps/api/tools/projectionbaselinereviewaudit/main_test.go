package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryProjectionBaselineRequirementsPass(t *testing.T) {
	apiRoot, err := resolveAPIRoot("")
	if err != nil {
		t.Fatalf("resolve API root: %v", err)
	}

	failures := inspectRequirements(apiRoot, reviewRequirements())
	if len(failures) != 0 {
		t.Fatalf(
			"projection baseline review requirements failed:\n%s",
			strings.Join(failures, "\n"),
		)
	}
}

func TestInspectRequirementsReportsMissingAndForbiddenFragments(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(path, []byte("present forbidden"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	failures := inspectRequirements(
		root,
		[]requirement{
			{
				path:      "sample.txt",
				fragments: []string{"present", "missing"},
				forbidden: []string{"forbidden"},
			},
		},
	)
	if len(failures) != 2 {
		t.Fatalf("failures = %#v, want two failures", failures)
	}
	combined := strings.Join(failures, "\n")
	if !strings.Contains(combined, `misses "missing"`) {
		t.Fatalf("missing-fragment failure not reported: %#v", failures)
	}
	if !strings.Contains(combined, `retains forbidden "forbidden"`) {
		t.Fatalf("forbidden-fragment failure not reported: %#v", failures)
	}
}

func TestResolveAPIRootRejectsUnrelatedDirectory(t *testing.T) {
	_, err := resolveAPIRoot(t.TempDir())
	if err == nil {
		t.Fatal("resolveAPIRoot() error = nil, want error")
	}
}
