package main

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

func TestTargetedValidationModulesAreSplitByResponsibility(
	t *testing.T,
) {
	apiRoot := projectAPIRoot(t)

	targets := []struct {
		directory     string
		removedFile   string
		generatedGlob string
		minimumFiles  int
	}{
		{
			directory: filepath.Join(
				apiRoot,
				"internal",
				"historicalintelligence",
				"historicalcontract",
			),
			removedFile:   "validation.go",
			generatedGlob: "historical_validation_*.go",
			minimumFiles:  6,
		},
		{
			directory: filepath.Join(
				apiRoot,
				"internal",
				"routeintelligence",
				"routecontract",
			),
			removedFile:   "validation.go",
			generatedGlob: "route_validation_*.go",
			minimumFiles:  5,
		},
	}

	for _, target := range targets {
		removedPath := filepath.Join(
			target.directory,
			target.removedFile,
		)
		if _, err := os.Stat(removedPath); err == nil {
			t.Fatalf(
				"large validation source remains: %s",
				removedPath,
			)
		} else if !os.IsNotExist(err) {
			t.Fatalf(
				"inspect removed validation source %s: %v",
				removedPath,
				err,
			)
		}

		files, err := filepath.Glob(
			filepath.Join(
				target.directory,
				target.generatedGlob,
			),
		)
		if err != nil {
			t.Fatalf(
				"glob generated validation files: %v",
				err,
			)
		}
		if len(files) < target.minimumFiles {
			t.Fatalf(
				"generated validation file count = %d, want at least %d in %s",
				len(files),
				target.minimumFiles,
				target.directory,
			)
		}

		for _, path := range files {
			assertMaximumSourceLines(
				t,
				path,
				500,
			)
		}
	}
}

func TestProjectionLargeModulesAreDecomposed(
	t *testing.T,
) {
	apiRoot := projectAPIRoot(t)

	directories := []string{
		filepath.Join(
			apiRoot,
			"internal",
			"projectionintelligence",
			"projectioncontinuation",
		),
		filepath.Join(
			apiRoot,
			"internal",
			"projectionintelligence",
			"projectionarrival",
		),
	}

	removed := []string{
		filepath.Join(
			directories[0],
			"continuation.go",
		),
		filepath.Join(
			directories[1],
			"estimator.go",
		),
	}
	for _, path := range removed {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf(
				"large projection source remains: %s",
				path,
			)
		} else if !os.IsNotExist(err) {
			t.Fatalf(
				"inspect removed projection source %s: %v",
				path,
				err,
			)
		}
	}

	for _, directory := range directories {
		files, err := filepath.Glob(
			filepath.Join(directory, "*.go"),
		)
		if err != nil {
			t.Fatalf(
				"glob projection files: %v",
				err,
			)
		}
		for _, path := range files {
			if strings.HasSuffix(
				path,
				"_test.go",
			) {
				continue
			}
			assertMaximumSourceLines(
				t,
				path,
				500,
			)
		}
	}
}

func TestProjectionOrchestratorsRemainNarrow(
	t *testing.T,
) {
	apiRoot := projectAPIRoot(t)

	assertFunctionMaximumLines(
		t,
		filepath.Join(
			apiRoot,
			"internal",
			"projectionintelligence",
			"projectioncontinuation",
		),
		"Project",
		90,
	)
	assertFunctionMaximumLines(
		t,
		filepath.Join(
			apiRoot,
			"internal",
			"projectionintelligence",
			"projectionarrival",
		),
		"Estimate",
		90,
	)
	assertFunctionMaximumLines(
		t,
		filepath.Join(
			apiRoot,
			"internal",
			"projectionintelligence",
			"projectionarrival",
		),
		"computeArrival",
		55,
	)
}

func projectAPIRoot(
	t *testing.T,
) string {
	t.Helper()

	_, currentFile, _, ok :=
		runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve project audit source path",
		)
	}

	return filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"..",
			"..",
		),
	)
}

func assertMaximumSourceLines(
	t *testing.T,
	path string,
	maximum int,
) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"read %s: %v",
			path,
			err,
		)
	}
	lineCount := strings.Count(
		string(content),
		"\n",
	) + 1
	if lineCount > maximum {
		t.Fatalf(
			"%s has %d lines, maximum is %d",
			path,
			lineCount,
			maximum,
		)
	}
}

func assertFunctionMaximumLines(
	t *testing.T,
	directory string,
	functionName string,
	maximum int,
) {
	t.Helper()

	files, err := filepath.Glob(
		filepath.Join(directory, "*.go"),
	)
	if err != nil {
		t.Fatalf(
			"glob %s: %v",
			directory,
			err,
		)
	}

	found := false
	for _, path := range files {
		if strings.HasSuffix(
			path,
			"_test.go",
		) {
			continue
		}

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
				path,
				err,
			)
		}

		for _, declaration := range file.Decls {
			function, ok :=
				declaration.(*ast.FuncDecl)
			if !ok ||
				function.Name.Name !=
					functionName {
				continue
			}
			found = true
			start := fileSet.Position(
				function.Pos(),
			).Line
			end := fileSet.Position(
				function.End(),
			).Line
			lineCount := end - start + 1
			if lineCount > maximum {
				t.Fatalf(
					"%s in %s has %d lines, maximum is %d",
					functionName,
					path,
					lineCount,
					maximum,
				)
			}
		}
	}

	if !found {
		t.Fatalf(
			"function %s was not found in %s",
			functionName,
			directory,
		)
	}
}
