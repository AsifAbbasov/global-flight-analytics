package calculator

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/registry"
)

func TestNew(
	t *testing.T,
) {
	reg := registry.New()

	calculator := New(reg)

	if calculator == nil {
		t.Fatal("calculator is nil")
	}

	if calculator.Registry() != reg {
		t.Fatal("registry was not stored")
	}
}
