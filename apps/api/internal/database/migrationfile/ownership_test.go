package migrationfile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCanonicalParserOwnsMigrationFileIdentityInterpretation(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migrationfile test path")
	}

	packageDir := filepath.Dir(currentFile)
	testCases := []struct {
		path       string
		required   []string
		prohibited []string
	}{
		{
			path: filepath.Join(packageDir, "../migrator/runner.go"),
			required: []string{
				"migrationfile.Parse(fileName)",
			},
			prohibited: []string{
				"func parseMigrationFileName(",
			},
		},
		{
			path: filepath.Join(packageDir, "../migrationaudit/scanner.go"),
			required: []string{
				"migrationfile.Parse(fileName)",
			},
			prohibited: []string{
				"func parseLocalMigrationFileName(",
			},
		},
		{
			path: filepath.Join(packageDir, "../migrationrepair/contracts.go"),
			required: []string{
				"migrationfile.MustParse(",
				"ExpectedAppliedVersion010FileName",
			},
		},
	}

	for _, testCase := range testCases {
		contentBytes, err := os.ReadFile(filepath.Clean(testCase.path))
		if err != nil {
			t.Fatalf("read %s: %v", testCase.path, err)
		}
		content := string(contentBytes)

		for _, required := range testCase.required {
			if !strings.Contains(content, required) {
				t.Fatalf("%s does not contain %q", testCase.path, required)
			}
		}
		for _, prohibited := range testCase.prohibited {
			if strings.Contains(content, prohibited) {
				t.Fatalf("%s still contains %q", testCase.path, prohibited)
			}
		}
	}
}
