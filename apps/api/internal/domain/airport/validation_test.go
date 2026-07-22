package airport

import (
	"errors"
	"math"
	"testing"
)

func TestAirportValidateRejectsMissingICAOAndInvalidCoordinates(t *testing.T) {
	if err := (Airport{}).Validate(); !errors.Is(err, ErrAirportICAORequired) {
		t.Fatalf("missing ICAO error = %v", err)
	}
	if err := (Airport{ICAOCode: "UBBB", Latitude: math.NaN(), Longitude: 49}).Validate(); !errors.Is(err, ErrAirportCoordinatesInvalid) {
		t.Fatalf("coordinate error = %v", err)
	}
}
