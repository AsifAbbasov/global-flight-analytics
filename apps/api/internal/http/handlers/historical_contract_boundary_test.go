package handlers

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

func TestHistoricalHTTPBoundaryDoesNotImportPostgresStore(
	t *testing.T,
) {
	files := []string{
		"historical_intelligence.go",
		filepath.Join(
			"..",
			"dto",
			"historical_intelligence.go",
		),
	}

	for _, relative := range files {
		imports := parseImports(
			t,
			relative,
		)
		for _, imported := range imports {
			if imported ==
				"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate" {
				t.Fatalf(
					"%s imports PostgreSQL store package",
					relative,
				)
			}
			if strings.Contains(
				imported,
				"github.com/jackc/pgx",
			) {
				t.Fatalf(
					"%s imports pgx infrastructure %q",
					relative,
					imported,
				)
			}
		}
	}
}

func TestHistoricalAggregateContractHasNoInfrastructureImports(
	t *testing.T,
) {
	relative := filepath.Join(
		"..",
		"..",
		"historicalintelligence",
		"historicalaggregatecontract",
		"contracts.go",
	)
	imports := parseImports(t, relative)

	for _, imported := range imports {
		for _, forbidden := range []string{
			"github.com/jackc/pgx",
			"/repository/",
			"/integrations/",
			"github.com/gofiber/fiber",
		} {
			if strings.Contains(
				imported,
				forbidden,
			) {
				t.Fatalf(
					"pure historical aggregate contract imports infrastructure %q",
					imported,
				)
			}
		}
	}
}

func TestHistoricalQueryParserHasNoBooleanModeParameter(
	t *testing.T,
) {
	_, currentFile, _, ok :=
		runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve handler source path",
		)
	}

	path := filepath.Join(
		filepath.Dir(currentFile),
		"historical_intelligence.go",
	)
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(
		fileSet,
		path,
		nil,
		parser.AllErrors,
	)
	if err != nil {
		t.Fatalf(
			"parse historical handler: %v",
			err,
		)
	}

	for _, declaration := range file.Decls {
		function, ok :=
			declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if !strings.HasPrefix(
			function.Name.Name,
			"parseHistorical",
		) {
			continue
		}
		if function.Type.Params == nil {
			continue
		}

		for _, field := range function.Type.Params.List {
			identifier, ok :=
				field.Type.(*ast.Ident)
			if ok &&
				identifier.Name == "bool" {
				t.Fatalf(
					"%s uses a boolean mode parameter",
					function.Name.Name,
				)
			}
		}
	}
}

func parseImports(
	t *testing.T,
	relative string,
) []string {
	t.Helper()

	_, currentFile, _, ok :=
		runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve handler source directory",
		)
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

	imports := make(
		[]string,
		0,
		len(file.Imports),
	)
	for _, specification := range file.Imports {
		value, err := strconv.Unquote(
			specification.Path.Value,
		)
		if err != nil {
			t.Fatalf(
				"unquote import in %s: %v",
				path,
				err,
			)
		}
		imports = append(
			imports,
			value,
		)
	}

	return imports
}
