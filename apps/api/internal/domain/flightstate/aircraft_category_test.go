package flightstate

import (
	"errors"
	"testing"
)

func TestAircraftCategoryValueObjectPreservesAvailability(t *testing.T) {
	observedZero, err := NewAircraftCategory(0)
	if err != nil {
		t.Fatalf("NewAircraftCategory(0) error = %v", err)
	}
	if !observedZero.Available() || observedZero.Value() != 0 {
		t.Fatalf("unexpected observed zero category: %+v", observedZero)
	}

	unavailable := UnavailableAircraftCategory()
	if unavailable.Available() || unavailable.Value() != 0 {
		t.Fatalf("unexpected unavailable category: %+v", unavailable)
	}
}

func TestFlightStateRejectsUnavailableNonZeroAircraftCategory(t *testing.T) {
	state := FlightState{
		AircraftCategory:          6,
		AircraftCategoryAvailable: false,
	}
	_, err := state.ResolveAircraftCategory()
	if !errors.Is(err, ErrAircraftCategoryInvalid) {
		t.Fatalf("ResolveAircraftCategory() error = %v", err)
	}
}
