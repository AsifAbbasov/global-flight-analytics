package airport

import (
	"math"
	"testing"
)

func TestResolveElevationPreservesObservedZero(t *testing.T) {
	value, status, available := ResolveElevation(0, true)
	if value != 0 || status != ElevationStatusObserved || !available {
		t.Fatalf("unexpected observed zero: value=%v status=%q available=%v", value, status, available)
	}
}

func TestResolveElevationKeepsUnknownZeroUnknown(t *testing.T) {
	value, status, available := ResolveElevation(0, false)
	if value != 0 || status != ElevationStatusUnknown || available {
		t.Fatalf("unexpected unknown zero: value=%v status=%q available=%v", value, status, available)
	}
}

func TestResolveElevationPreservesLegacyNonZeroAndNegativeValues(t *testing.T) {
	for _, input := range []float64{12.5, -38} {
		value, status, available := ResolveElevation(input, false)
		if value != input || status != ElevationStatusObserved || !available {
			t.Fatalf("input %v resolved as value=%v status=%q available=%v", input, value, status, available)
		}
	}
}

func TestResolveElevationRejectsNonFiniteValues(t *testing.T) {
	for _, input := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		value, status, available := ResolveElevation(input, true)
		if value != 0 || status != ElevationStatusInvalid || available {
			t.Fatalf("input %v resolved as value=%v status=%q available=%v", input, value, status, available)
		}
	}
}
