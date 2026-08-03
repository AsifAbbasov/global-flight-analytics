package dependency_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestLegacyTrafficRootPackageDoesNotReturn(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve contract test path")
	}

	moduleRoot := filepath.Clean(
		filepath.Join(filepath.Dir(currentFile), "..", "..", ".."),
	)
	trafficRoot := filepath.Join(
		moduleRoot,
		"internal",
		"services",
		"traffic",
	)

	entries, err := os.ReadDir(trafficRoot)
	if err != nil {
		t.Fatalf("read traffic service root: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		t.Fatalf(
			"legacy traffic root package returned through %s",
			filepath.Join(trafficRoot, entry.Name()),
		)
	}

	legacyImport :=
		"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/" +
			"traffic"
	quotedLegacyImport := strconv.Quote(legacyImport)

	err = filepath.WalkDir(
		moduleRoot,
		func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if entry.Name() == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(entry.Name(), ".go") {
				return nil
			}

			source, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			if strings.Contains(string(source), quotedLegacyImport) {
				t.Errorf(
					"legacy traffic root import found in %s",
					path,
				)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatalf("scan Go imports: %v", err)
	}
}
