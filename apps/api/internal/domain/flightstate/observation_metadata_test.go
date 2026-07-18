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

func TestValidateAircraftCategoryPreservesAvailabilityMeaning(t *testing.T) {
	if err := ValidateAircraftCategory(0, true); err != nil {
		t.Fatalf("category zero can be observed: %v", err)
	}
	if err := ValidateAircraftCategory(0, false); err != nil {
		t.Fatalf("unavailable category zero: %v", err)
	}
	if err := ValidateAircraftCategory(6, true); err != nil {
		t.Fatalf("observed heavy category: %v", err)
	}
	if !errors.Is(
		ValidateAircraftCategory(6, false),
		ErrAircraftCategoryInvalid,
	) {
		t.Fatal("expected unavailable non-zero category to be rejected")
	}
}
