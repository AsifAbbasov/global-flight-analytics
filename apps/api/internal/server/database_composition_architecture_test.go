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

func TestDatabaseRouteCoordinatorRemainsNarrow(
	t *testing.T,
) {
	source := parseServerSource(
		t,
		"database_routes.go",
	)

	allowedImports := map[string]struct{}{
		"fmt":                             {},
		"time":                            {},
		"github.com/gofiber/fiber/v2":     {},
		"github.com/jackc/pgx/v5/pgxpool": {},
	}
	for _, imported := range source.imports {
		if _, allowed := allowedImports[imported]; !allowed {
			t.Fatalf(
				"database route coordinator imports concrete bounded-context dependency %q",
				imported,
			)
		}
	}

	lineCount := source.functionLineCount(
		t,
		"registerDatabaseRoutes",
	)
	if lineCount > 24 {
		t.Fatalf(
			"registerDatabaseRoutes has %d lines, maximum is 24",
			lineCount,
		)
	}

	if source.hasSelectorCall(
		"Get",
		"Post",
		"Put",
		"Patch",
		"Delete",
	) {
		t.Fatal(
			"database route coordinator directly registers HTTP routes",
		)
	}
}

func TestDatabaseCompositionFilesDoNotRegisterHTTPRoutes(
	t *testing.T,
) {
	files := []string{
		"core_database_composition.go",
		"route_intelligence_database_composition.go",
		"projection_database_composition.go",
		"airspace_database_composition.go",
	}

	for _, name := range files {
		source := parseServerSource(t, name)
		if source.hasSelectorCall(
			"Get",
			"Post",
			"Put",
			"Patch",
			"Delete",
		) {
			t.Fatalf(
				"composition file %s directly registers an HTTP route",
				name,
			)
		}
	}
}

func TestDatabaseRouteFilesDoNotConstructInfrastructure(
	t *testing.T,
) {
	files := []string{
		"core_database_routes.go",
		"route_intelligence_database_routes.go",
		"projection_database_routes.go",
		"airspace_database_routes.go",
	}
	forbiddenImportFragments := []string{
		"/repository/postgres",
		"/domain/",
		"/airspaceintelligence/",
		"/routeintelligence/",
		"/stabilityintelligence/",
		"github.com/jackc/pgx",
	}

	for _, name := range files {
		source := parseServerSource(t, name)
		for _, imported := range source.imports {
			for _, fragment := range forbiddenImportFragments {
				if strings.Contains(
					imported,
					fragment,
				) {
					t.Fatalf(
						"route file %s imports infrastructure or domain construction dependency %q",
						name,
						imported,
					)
				}
			}
		}
	}
}

type parsedServerSource struct {
	fileSet *token.FileSet
	file    *ast.File
	imports []string
}

func parseServerSource(
	t *testing.T,
	name string,
) parsedServerSource {
	t.Helper()

	_, currentFile, _, ok :=
		runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve server test source path",
		)
	}

	path := filepath.Join(
		filepath.Dir(currentFile),
		name,
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
			"parse %s: %v",
			name,
			err,
		)
	}

	imports := make(
		[]string,
		0,
		len(file.Imports),
	)
	for _, item := range file.Imports {
		value, unquoteErr :=
			strconv.Unquote(
				item.Path.Value,
			)
		if unquoteErr != nil {
			t.Fatalf(
				"unquote import in %s: %v",
				name,
				unquoteErr,
			)
		}
		imports = append(
			imports,
			value,
		)
	}

	return parsedServerSource{
		fileSet: fileSet,
		file:    file,
		imports: imports,
	}
}

func (
	source parsedServerSource,
) functionLineCount(
	t *testing.T,
	name string,
) int {
	t.Helper()

	for _, declaration := range source.file.Decls {
		function, ok :=
			declaration.(*ast.FuncDecl)
		if !ok ||
			function.Name.Name != name {
			continue
		}

		start := source.fileSet.
			Position(
				function.Pos(),
			).
			Line
		end := source.fileSet.
			Position(
				function.End(),
			).
			Line
		return end - start + 1
	}

	t.Fatalf(
		"function %s was not found",
		name,
	)
	return 0
}

func (
	source parsedServerSource,
) hasSelectorCall(
	names ...string,
) bool {
	expected := make(
		map[string]struct{},
		len(names),
	)
	for _, name := range names {
		expected[name] = struct{}{}
	}

	found := false
	ast.Inspect(
		source.file,
		func(node ast.Node) bool {
			call, ok :=
				node.(*ast.CallExpr)
			if !ok {
				return true
			}

			selector, ok :=
				call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if _, exists := expected[selector.Sel.Name]; exists {
				found = true
				return false
			}

			return true
		},
	)
	return found
}
