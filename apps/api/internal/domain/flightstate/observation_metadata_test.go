package flightstate

import (
	"errors"
	"testing"
)

func TestNormalizeSquawkCodeAcceptsOctalEvidence(t *testing.T) {
	value, err := NormalizeSquawkCode(" 7700 ")
	if err != nil {
		t.Fatalf("normalize squawk: %v", err)
	}
	if value != "7700" {
		t.Fatalf("squawk = %q, want 7700", value)
	}
}

func TestNormalizeSquawkCodeRejectsNonOctalValue(t *testing.T) {
	_, err := NormalizeSquawkCode("7800")
	if !errors.Is(err, ErrSquawkCodeInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrSquawkCodeInvalid)
	}
}

func TestAircraftCategoryValueObjectPreservesAvailabilityMeaning(t *testing.T) {
	if _, err := NewAircraftCategory(0); err != nil {
		t.Fatalf("category zero can be observed: %v", err)
	}
	if err := UnavailableAircraftCategory().Validate(); err != nil {
		t.Fatalf("unavailable category zero: %v", err)
	}
	if _, err := NewAircraftCategory(6); err != nil {
		t.Fatalf("observed heavy category: %v", err)
	}
	state := FlightState{
		AircraftCategory:          6,
		AircraftCategoryAvailable: false,
	}
	if _, err := state.ResolveAircraftCategory(); !errors.Is(
		err,
		ErrAircraftCategoryInvalid,
	) {
		t.Fatal("expected unavailable non-zero category to be rejected")
	}
}
