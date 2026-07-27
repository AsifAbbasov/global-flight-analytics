package trajectory

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestTrackPoint4DFromFlightStatePreservesTelemetryAvailability(t *testing.T) {
	observedAt := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	state := flightstate.FlightState{
		ID:                         "state-one",
		FlightID:                   "flight-one",
		ICAO24:                     "ABC123",
		VelocityMPS:                0,
		VelocityAvailable:          false,
		HeadingDegrees:             0,
		HeadingAvailable:           true,
		VerticalRateMPS:            0,
		VerticalRateAvailable:      false,
		OnGround:                   false,
		OnGroundAvailable:          false,
		TelemetryAvailabilityKnown: true,
		ObservedAt:                 observedAt,
	}

	point := TrackPoint4DFromFlightState(state)
	if point.ID != state.ID || point.FlightStateID != state.ID {
		t.Fatalf("point identity = %#v", point)
	}
	if point.HasVelocity() || !point.HasHeading() || point.HasVerticalRate() || point.HasOnGroundState() {
		t.Fatalf("availability contract was not preserved: %#v", point)
	}
	if !point.ObservedAt.Equal(observedAt) {
		t.Fatalf("observed at = %s", point.ObservedAt)
	}
}

func TestLegacyTrackPointTreatsTelemetryAsAvailable(t *testing.T) {
	point := TrackPoint4D{}
	if !point.HasVelocity() || !point.HasHeading() || !point.HasVerticalRate() || !point.HasOnGroundState() {
		t.Fatal("legacy point availability compatibility changed")
	}
}
