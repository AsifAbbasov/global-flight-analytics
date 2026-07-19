package main

import (
	"strings"
	"testing"
)

func TestValidateFormulaBenchmarkBoundaryAcceptsOfflineSeparation(
	t *testing.T,
) {
	benchmark := map[string]struct{}{
		formulaBenchmarkPackage:     {},
		researchBenchmarkPackage:    {},
		researchDatasetPackage:      {},
		projectionEvaluationPackage: {},
	}
	runtime := map[string]struct{}{
		modulePath + "/internal/server": {},
	}

	if err := validateFormulaBenchmarkBoundary(
		benchmark,
		runtime,
		"FROM scratch\n",
	); err != nil {
		t.Fatalf("validate boundary: %v", err)
	}
}

func TestValidateFormulaBenchmarkBoundaryRejectsRuntimeLeak(
	t *testing.T,
) {
	benchmark := map[string]struct{}{
		formulaBenchmarkPackage:     {},
		researchBenchmarkPackage:    {},
		researchDatasetPackage:      {},
		projectionEvaluationPackage: {},
	}
	runtime := map[string]struct{}{
		formulaBenchmarkPackage: {},
	}

	err := validateFormulaBenchmarkBoundary(
		benchmark,
		runtime,
		"FROM scratch\n",
	)
	if err == nil ||
		!strings.Contains(
			err.Error(),
			"leaked into production runtime",
		) {
		t.Fatalf("error = %v", err)
	}
}

func TestValidateFormulaBenchmarkBoundaryRejectsDockerInclusion(
	t *testing.T,
) {
	benchmark := map[string]struct{}{
		formulaBenchmarkPackage:     {},
		researchBenchmarkPackage:    {},
		researchDatasetPackage:      {},
		projectionEvaluationPackage: {},
	}

	err := validateFormulaBenchmarkBoundary(
		benchmark,
		map[string]struct{}{},
		"RUN go build ./cmd/benchmark-projection-formulas\n",
	)
	if err == nil ||
		!strings.Contains(
			err.Error(),
			"production Docker image",
		) {
		t.Fatalf("error = %v", err)
	}
}
