package server

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestTransponderEvidenceHTTPBoundaryHasNoPostgresDependency(
	t *testing.T,
) {
	files := []string{
		filepath.Join(
			"..",
			"http",
			"handlers",
			"transponder_evidence.go",
		),
		filepath.Join(
			"..",
			"http",
			"dto",
			"transponder_evidence.go",
		),
		filepath.Join(
			"..",
			"analytics",
			"transponderalert",
			"service.go",
		),
	}

	for _, relative := range files {
		for _, imported := range transponderSourceImports(
			t,
			relative,
		) {
			for _, forbidden := range []string{
				"/repository/postgres",
				"github.com/jackc/pgx",
			} {
				if strings.Contains(
					imported,
					forbidden,
				) {
					t.Fatalf(
						"%s imports infrastructure %q",
						relative,
						imported,
					)
				}
			}
		}
	}
}

func TestTransponderEvidenceResponseContainsSafetyFields(
	t *testing.T,
) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve server test source path")
	}
	path := filepath.Join(
		filepath.Dir(currentFile),
		"..",
		"http",
		"dto",
		"transponder_evidence.go",
	)
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(
		fileSet,
		path,
		nil,
		parser.AllErrors,
	)
	if err != nil {
		t.Fatalf("parse transponder DTO: %v", err)
	}

	required := map[string]bool{
		"EvidenceOnly":         false,
		"ConfirmedEmergency":   false,
		"OperationalAlert":     false,
		"MaximumClaimStrength": false,
		"Limitations":          false,
	}

	ast.Inspect(file, func(node ast.Node) bool {
		field, ok := node.(*ast.Field)
		if !ok {
			return true
		}
		for _, name := range field.Names {
			if _, exists := required[name.Name]; exists {
				required[name.Name] = true
			}
		}
		return true
	})

	for name, found := range required {
		if !found {
			t.Fatalf(
				"required safety field %s was not found",
				name,
			)
		}
	}
}

func transponderSourceImports(
	t *testing.T,
	relative string,
) []string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve server test source path")
	}
	path := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			relative,
		),
	)
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(
		fileSet,
		path,
		nil,
		parser.ImportsOnly,
	)
	if err != nil {
		t.Fatalf(
			"parse imports from %s: %v",
			path,
			err,
		)
	}

	imports := make([]string, 0, len(file.Imports))
	for _, item := range file.Imports {
		value, err := strconv.Unquote(
			item.Path.Value,
		)
		if err != nil {
			t.Fatalf(
				"unquote import in %s: %v",
				path,
				err,
			)
		}
		imports = append(imports, value)
	}
	return imports
}
