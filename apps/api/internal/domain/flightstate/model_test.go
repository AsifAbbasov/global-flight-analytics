package flightstate

import (
	"math"
	"testing"
)

func TestResolveAltitudeStatusRejectsUnknownStatus(t *testing.T) {
	status := ResolveAltitudeStatus(100, AltitudeStatus("broken"))
	if status != AltitudeStatusInvalid {
		t.Fatalf("status = %q, want %q", status, AltitudeStatusInvalid)
	}
}

func TestResolveAltitudeStatusRejectsNonFiniteValuesWithoutExplicitStatus(t *testing.T) {
	tests := []float64{math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, value := range tests {
		if status := ResolveAltitudeStatus(value, ""); status != AltitudeStatusInvalid {
			t.Fatalf("value %v resolved to %q, want %q", value, status, AltitudeStatusInvalid)
		}
	}
}

func TestResolveAltitudeStatusRejectsObservedNonFiniteValues(t *testing.T) {
	tests := []float64{math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, value := range tests {
		if status := ResolveAltitudeStatus(value, AltitudeStatusObserved); status != AltitudeStatusInvalid {
			t.Fatalf("observed value %v resolved to %q, want %q", value, status, AltitudeStatusInvalid)
		}
	}
}

func TestResolveAltitudeStatusPreservesExplicitUnavailableSentinel(t *testing.T) {
	status := ResolveAltitudeStatus(math.NaN(), AltitudeStatusUnavailable)
	if status != AltitudeStatusUnavailable {
		t.Fatalf("status = %q, want %q", status, AltitudeStatusUnavailable)
	}
}

func TestResolveAltitudeStatusPreservesExplicitGroundState(t *testing.T) {
	status := ResolveAltitudeStatus(1234.9, AltitudeStatusGround)
	if status != AltitudeStatusGround {
		t.Fatalf("status = %q, want %q", status, AltitudeStatusGround)
	}
}

func TestResolveAltitudeStatusPreservesExplicitObservedZero(t *testing.T) {
	status := ResolveAltitudeStatus(0, AltitudeStatusObserved)
	if status != AltitudeStatusObserved {
		t.Fatalf("status = %q, want %q", status, AltitudeStatusObserved)
	}
}
