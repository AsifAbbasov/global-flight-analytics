package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryArgumentsDoNotUsePointerNilOrArtificialSourceFallback(t *testing.T) {
	t.Parallel()

	source := readPostgresContractSource(t, "repository_helpers.go")
	for _, required := range []string{
		"type nullableUUIDArgument struct",
		"type nullableTextArgument struct",
		"type requiredSourceNameArgument struct",
		"ErrRepositoryUUIDArgumentInvalid",
		"ErrRepositorySourceNameRequired",
		"func requiredSourceNameValue(",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("repository helper contract is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		") *string",
		"return \"unknown\"",
		"func sourceNameOrUnknown(",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("repository helper contract contains retired fragment %q", forbidden)
		}
	}
}

func TestInternalPostgresQueriesDoNotCastUUIDColumnsToTextForArrayMembership(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve source location")
	}
	internalRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../.."))
	forbidden := "::text = " + "ANY("
	forbiddenSpaced := "::text = " + "ANY ("

	err := filepath.WalkDir(
		internalRoot,
		func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			text := string(content)
			if strings.Contains(text, forbidden) || strings.Contains(text, forbiddenSpaced) {
				t.Fatalf("UUID array query still casts an indexed column to text: %s", path)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("walk internal Go source: %v", err)
	}
}

func TestArtificialUnknownSourceFallbackIsAbsentFromInternalPostgresCode(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve source location")
	}
	internalRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../.."))
	retiredCall := "sourceName" + "OrUnknown("

	err := filepath.WalkDir(
		internalRoot,
		func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			if strings.Contains(string(content), retiredCall) {
				t.Fatalf("artificial source fallback remains in %s", path)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("walk internal Go source: %v", err)
	}
}

func readPostgresContractSource(t *testing.T, fileName string) string {
	t.Helper()
	content, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf("read %s: %v", fileName, err)
	}
	return string(content)
}
