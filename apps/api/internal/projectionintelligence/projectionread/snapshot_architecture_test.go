package projectionread

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProjectionReadDataSourceExposesOneSnapshotOperation(
	t *testing.T,
) {
	path := projectionReadSourcePath(
		t,
		"contracts.go",
	)
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(
		fileSet,
		path,
		nil,
		parser.AllErrors,
	)
	if err != nil {
		t.Fatalf("parse contracts.go: %v", err)
	}

	methodNames := []string{}
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, specification := range general.Specs {
			typeSpecification, ok := specification.(*ast.TypeSpec)
			if !ok || typeSpecification.Name.Name != "DataSource" {
				continue
			}
			interfaceType, ok := typeSpecification.Type.(*ast.InterfaceType)
			if !ok {
				t.Fatal("DataSource is not an interface")
			}
			for _, method := range interfaceType.Methods.List {
				for _, name := range method.Names {
					methodNames = append(methodNames, name.Name)
				}
			}
		}
	}

	if len(methodNames) != 1 || methodNames[0] != "LoadSnapshot" {
		t.Fatalf(
			"DataSource methods = %v, want only LoadSnapshot",
			methodNames,
		)
	}
}

func TestProjectionReadServiceDoesNotCoordinateIndependentDatabaseReads(
	t *testing.T,
) {
	content, err := os.ReadFile(
		projectionReadSourcePath(t, "service.go"),
	)
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	text := string(content)
	for _, forbidden := range []string{
		"LoadCurrentTrajectory(",
		"LoadRoute(",
		"LoadHistoricalCandidates(",
		"LoadRouteHistory(",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf(
				"service coordinates independent read %q",
				forbidden,
			)
		}
	}
	if !strings.Contains(text, "LoadSnapshot(") {
		t.Fatal("service does not load a Projection Intelligence snapshot")
	}
}

func TestProductionPostgresDataSourceUsesRepeatableReadSnapshotExecutor(
	t *testing.T,
) {
	content, err := os.ReadFile(
		projectionReadSourcePath(t, "postgres_config.go"),
	)
	if err != nil {
		t.Fatalf("read postgres_config.go: %v", err)
	}
	text := string(content)
	for _, required := range []string{
		"repeatableReadSnapshotExecutor",
		"pgxSnapshotTransactionStarter",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf(
				"production snapshot wiring does not contain %q",
				required,
			)
		}
	}
}

func projectionReadSourcePath(
	t *testing.T,
	name string,
) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve Projection Intelligence source path")
	}
	return filepath.Join(
		filepath.Dir(currentFile),
		name,
	)
}
