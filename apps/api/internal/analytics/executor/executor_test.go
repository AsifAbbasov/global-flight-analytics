package executor

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/calculator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/registry"
)

func TestNew(
	t *testing.T,
) {
	reg := registry.New()

	calc := calculator.New(reg)

	executor := New(calc)

	if executor == nil {
		t.Fatal("executor is nil")
	}

	if executor.Calculator() != calc {
		t.Fatal("calculator was not stored")
	}
}
