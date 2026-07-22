package flightstate

import (
	"errors"
	"math"
	"testing"
	"time"
)

func validFlightStateContractFixture() FlightState {
	return FlightState{ICAO24: "abc123", Latitude: 40, Longitude: 49, ObservedAt: time.Now().UTC(), SourceName: "opensky"}
}

func TestFlightStateValidateRejectsInvalidCoordinatesAndHeading(t *testing.T) {
	state := validFlightStateContractFixture()
	state.Latitude = math.NaN()
	if err := state.Validate(); !errors.Is(err, ErrFlightStateCoordinatesInvalid) {
		t.Fatalf("coordinate error = %v", err)
	}
	state = validFlightStateContractFixture()
	state.HeadingAvailable = true
	state.HeadingDegrees = 360
	if err := state.Validate(); !errors.Is(err, ErrFlightStateHeadingInvalid) {
		t.Fatalf("heading error = %v", err)
	}
}
