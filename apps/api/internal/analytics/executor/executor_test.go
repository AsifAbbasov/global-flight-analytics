package executor

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/calculator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/registry"
)

func TestNewAcceptsLegacyCalculatorWithoutExposingIt(
	t *testing.T,
) {
	reg := registry.New()
	calc := calculator.New(reg)

	executor := New(calc)
	if executor == nil {
		t.Fatal("executor is nil")
	}

	if executor.scopeGuard == nil ||
		executor.confidenceEvaluator == nil {
		t.Fatal(
			"executor runtime dependencies are missing",
		)
	}
}
