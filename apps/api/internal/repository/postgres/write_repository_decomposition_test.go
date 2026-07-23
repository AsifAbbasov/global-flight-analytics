package postgres

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestWriteRepositoryCoordinatorsRemainNarrow(t *testing.T) {
	t.Parallel()

	for _, check := range []struct {
		path      string
		function  string
		maxLines  int
		forbidden []string
	}{
		{
			path:     "airport_import_repository.go",
			function: "UpsertImported",
			maxLines: 80,
			forbidden: []string{
				"CREATE TEMP TABLE",
				"UPDATE airports AS target",
				"INSERT INTO airports (",
				"stageAirportImportRecords(",
			},
		},
		{
			path:     "flightstate_repository.go",
			function: "SaveFlightStates",
			maxLines: 80,
			forbidden: []string{
				"INSERT INTO flight_states",
				"NormalizeSquawkCode(",
				"NormalizePositionSource(",
				"ResolveAircraftCategory(",
				"altitudeDatabaseValue(",
			},
		},
	} {
		functionSource, lineCount := goFunctionSource(
			t,
			check.path,
			check.function,
		)
		if lineCount > check.maxLines {
			t.Fatalf(
				"%s.%s line count=%d want<=%d",
				check.path,
				check.function,
				lineCount,
				check.maxLines,
			)
		}

		for _, forbidden := range check.forbidden {
			if strings.Contains(functionSource, forbidden) {
				t.Fatalf(
					"%s retains delegated responsibility %q",
					check.path,
					forbidden,
				)
			}
		}
	}
}

func TestWriteRepositoryResponsibilitiesHaveDedicatedOwners(t *testing.T) {
	t.Parallel()

	for _, check := range []struct {
		path     string
		required []string
	}{
		{
			path: "airport_import_write_steps.go",
			required: []string{
				"func executeAirportImport(",
				"createAirportImportStagingTable(",
				"stageAirportImportRecords(",
				"updateAirportsByICAO(",
				"updateAirportsBySourceIdentity(",
				"insertRemainingAirports(",
			},
		},
		{
			path: "airport_import_staging_write.go",
			required: []string{
				"CREATE TEMP TABLE airport_import_staging",
				"INSERT INTO airport_import_staging",
				"func createAirportImportStagingTable(",
				"func stageAirportImportRecords(",
			},
		},
		{
			path: "airport_import_merge_write.go",
			required: []string{
				"UPDATE airports AS target",
				"INSERT INTO airports (",
				"func updateAirportsByICAO(",
				"func updateAirportsBySourceIdentity(",
				"func insertRemainingAirports(",
			},
		},
		{
			path: "flightstate_write.go",
			required: []string{
				"INSERT INTO flight_states",
				"func saveFlightStateBatch(",
				"func prepareFlightStateInsertArguments(",
				"flightstate.NormalizeSquawkCode(",
				"item.ResolveAircraftCategory()",
			},
		},
	} {
		content := readRepositorySource(t, check.path)
		for _, required := range check.required {
			if !strings.Contains(content, required) {
				t.Fatalf(
					"%s is missing ownership token %q",
					check.path,
					required,
				)
			}
		}
	}
}

func TestWriteRepositoryCoordinatorsPreserveTransactionBoundary(t *testing.T) {
	t.Parallel()

	for _, check := range []struct {
		path       string
		function   string
		delegate   string
		commitCall string
	}{
		{
			path:       "airport_import_repository.go",
			function:   "UpsertImported",
			delegate:   "executeAirportImport(",
			commitCall: "tx.Commit(",
		},
		{
			path:       "flightstate_repository.go",
			function:   "SaveFlightStatesCounted",
			delegate:   "saveFlightStateBatch(",
			commitCall: "tx.Commit(",
		},
	} {
		functionSource, _ := goFunctionSource(
			t,
			check.path,
			check.function,
		)
		for _, required := range []string{
			"BeginTx(",
			"rollbackRepositoryTransaction(tx)",
			check.delegate,
			check.commitCall,
		} {
			if !strings.Contains(functionSource, required) {
				t.Fatalf(
					"%s.%s is missing transaction token %q",
					check.path,
					check.function,
					required,
				)
			}
		}

		if strings.Index(functionSource, check.delegate) >
			strings.Index(functionSource, check.commitCall) {
			t.Fatalf(
				"%s.%s commits before delegated write completion",
				check.path,
				check.function,
			)
		}
	}
}

func TestFlightStateLegacySaveDelegatesToCountedTransactionOwner(
	t *testing.T,
) {
	t.Parallel()

	functionSource, _ := goFunctionSource(
		t,
		"flightstate_repository.go",
		"SaveFlightStates",
	)
	if !strings.Contains(
		functionSource,
		"SaveFlightStatesCounted(",
	) {
		t.Fatal(
			"SaveFlightStates must delegate to SaveFlightStatesCounted",
		)
	}
}

func goFunctionSource(
	t *testing.T,
	path string,
	functionName string,
) (string, int) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, content, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != functionName {
			continue
		}
		start := fileSet.Position(function.Pos())
		end := fileSet.Position(function.End())
		return string(content[start.Offset:end.Offset]), end.Line - start.Line + 1
	}

	t.Fatalf("%s does not define %s", path, functionName)
	return "", 0
}

func readRepositorySource(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
