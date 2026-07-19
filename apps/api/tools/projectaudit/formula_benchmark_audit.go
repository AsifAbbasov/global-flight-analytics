package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	formulaBenchmarkCommandPattern = "./cmd/benchmark-projection-formulas"

	formulaBenchmarkPackage = modulePath +
		"/internal/analytics/formulabenchmark"
	researchBenchmarkPackage = modulePath +
		"/internal/analytics/researchbenchmark"
	researchDatasetPackage = modulePath +
		"/internal/analytics/researchdataset"
	projectionEvaluationPackage = modulePath +
		"/internal/projectionintelligence/projectionevaluation"
)

func auditFormulaBenchmarkBoundary(
	repositoryRoot string,
	output io.Writer,
) error {
	apiRoot := filepath.Join(
		repositoryRoot,
		"apps",
		"api",
	)

	benchmarkDependencies, err :=
		listDependencies(
			apiRoot,
			formulaBenchmarkCommandPattern,
		)
	if err != nil {
		return fmt.Errorf(
			"load formula benchmark dependencies: %w",
			err,
		)
	}

	runtimePatterns := []string{
		"./cmd/server",
		"./cmd/ingest",
		"./cmd/reconcile",
		"./cmd/materialize-historical-intelligence",
		"./cmd/materialize-flight-features",
	}
	runtimeDependencies := make(
		map[string]struct{},
	)
	for _, pattern := range runtimePatterns {
		dependencies, dependencyErr :=
			listDependencies(
				apiRoot,
				pattern,
			)
		if dependencyErr != nil {
			return fmt.Errorf(
				"load runtime dependencies for %s: %w",
				pattern,
				dependencyErr,
			)
		}
		for dependency := range dependencies {
			runtimeDependencies[dependency] =
				struct{}{}
		}
	}

	dockerfileBytes, err := os.ReadFile(
		filepath.Join(
			apiRoot,
			"Dockerfile",
		),
	)
	if err != nil {
		return fmt.Errorf(
			"read backend Dockerfile: %w",
			err,
		)
	}

	if err := validateFormulaBenchmarkBoundary(
		benchmarkDependencies,
		runtimeDependencies,
		string(dockerfileBytes),
	); err != nil {
		return err
	}

	fmt.Fprintln(
		output,
		"Formula benchmark boundary audit: PASS",
	)
	return nil
}

func validateFormulaBenchmarkBoundary(
	benchmarkDependencies map[string]struct{},
	runtimeDependencies map[string]struct{},
	dockerfile string,
) error {
	required := []string{
		formulaBenchmarkPackage,
		researchBenchmarkPackage,
		researchDatasetPackage,
		projectionEvaluationPackage,
	}
	missing := make([]string, 0)
	for _, importPath := range required {
		if _, exists :=
			benchmarkDependencies[importPath]; !exists {
			missing = append(
				missing,
				importPath,
			)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf(
			"formula benchmark command is missing required offline dependencies: %s",
			strings.Join(missing, ", "),
		)
	}

	leaked := make([]string, 0)
	for _, importPath := range required {
		if _, exists :=
			runtimeDependencies[importPath]; exists {
			leaked = append(
				leaked,
				importPath,
			)
		}
	}
	if len(leaked) > 0 {
		sort.Strings(leaked)
		return fmt.Errorf(
			"offline formula benchmark dependencies leaked into production runtime: %s",
			strings.Join(leaked, ", "),
		)
	}

	if strings.Contains(
		dockerfile,
		"benchmark-projection-formulas",
	) {
		return fmt.Errorf(
			"offline formula benchmark command must not be included in the production Docker image",
		)
	}

	return nil
}
