package executor

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/confidencereport"
)

func TestNewStoresDefaultConfidenceEvaluator(
	t *testing.T,
) {
	executor := New(nil)

	if executor.ConfidenceEvaluator() == nil {
		t.Fatal("expected default confidence evaluator")
	}
}

func TestNewWithScopeGuardStoresDefaultConfidenceEvaluator(
	t *testing.T,
) {
	executor := NewWithScopeGuard(
		nil,
		nil,
	)

	if executor.ConfidenceEvaluator() == nil {
		t.Fatal("expected default confidence evaluator")
	}
}

func TestNewWithDependenciesStoresProvidedConfidenceEvaluator(
	t *testing.T,
) {
	evaluator, err := confidencereport.New(
		confidencereport.Config{
			MediumThreshold:  0.50,
			HighThreshold:    0.70,
			MaximumPenalty:   0.80,
			DecimalPrecision: 4,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected confidence evaluator, got %v",
			err,
		)
	}

	executor := NewWithDependencies(
		nil,
		nil,
		evaluator,
	)

	if executor.ConfidenceEvaluator() != evaluator {
		t.Fatal("expected provided confidence evaluator")
	}
}

func TestNewWithDependenciesReplacesNilConfidenceEvaluator(
	t *testing.T,
) {
	executor := NewWithDependencies(
		nil,
		nil,
		nil,
	)

	if executor.ConfidenceEvaluator() == nil {
		t.Fatal("expected nil confidence evaluator replacement")
	}
}
