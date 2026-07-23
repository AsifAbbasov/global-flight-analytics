package deduplicator

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestRemoveExactDuplicatesKeepsEveryCanonicalPayloadDifference(
	t *testing.T,
) {
	base := flightstate.FlightState{
		ICAO24:                     "ABC123",
		Callsign:                   "AHY101",
		Latitude:                   40.4093,
		Longitude:                  49.8671,
		BarometricAltitudeM:        10000,
		BarometricAltitudeStatus:   flightstate.AltitudeStatusObserved,
		GeometricAltitudeM:         10050,
		GeometricAltitudeStatus:    flightstate.AltitudeStatusObserved,
		VelocityMPS:                230,
		VelocityAvailable:          true,
		HeadingDegrees:             90,
		HeadingAvailable:           true,
		VerticalRateMPS:            2.5,
		VerticalRateAvailable:      true,
		OnGround:                   false,
		OnGroundAvailable:          true,
		TelemetryAvailabilityKnown: true,
		OriginCountry:              "Azerbaijan",
		SquawkCode:                 "1200",
		SpecialPurposeIndicator:    false,
		PositionSource:             flightstate.PositionSourceADSB,
		AircraftCategory:           3,
		AircraftCategoryAvailable:  true,
		ObservedAt: time.Date(
			2026,
			time.July,
			23,
			8,
			0,
			0,
			0,
			time.UTC,
		),
		SourceName: "airplanes.live",
	}

	tests := []struct {
		name   string
		mutate func(*flightstate.FlightState)
	}{
		{name: "callsign", mutate: func(item *flightstate.FlightState) { item.Callsign = "AHY102" }},
		{name: "geometric altitude", mutate: func(item *flightstate.FlightState) { item.GeometricAltitudeM++ }},
		{name: "geometric altitude status", mutate: func(item *flightstate.FlightState) {
			item.GeometricAltitudeStatus = flightstate.AltitudeStatusUnavailable
		}},
		{name: "velocity availability", mutate: func(item *flightstate.FlightState) { item.VelocityAvailable = false }},
		{name: "heading availability", mutate: func(item *flightstate.FlightState) { item.HeadingAvailable = false }},
		{name: "vertical rate", mutate: func(item *flightstate.FlightState) { item.VerticalRateMPS++ }},
		{name: "vertical rate availability", mutate: func(item *flightstate.FlightState) { item.VerticalRateAvailable = false }},
		{name: "on ground", mutate: func(item *flightstate.FlightState) { item.OnGround = true }},
		{name: "on ground availability", mutate: func(item *flightstate.FlightState) { item.OnGroundAvailable = false }},
		{name: "telemetry availability knowledge", mutate: func(item *flightstate.FlightState) { item.TelemetryAvailabilityKnown = false }},
		{name: "origin country", mutate: func(item *flightstate.FlightState) { item.OriginCountry = "Georgia" }},
		{name: "squawk", mutate: func(item *flightstate.FlightState) { item.SquawkCode = "7700" }},
		{name: "special purpose indicator", mutate: func(item *flightstate.FlightState) { item.SpecialPurposeIndicator = true }},
		{name: "position source", mutate: func(item *flightstate.FlightState) { item.PositionSource = flightstate.PositionSourceMLAT }},
		{name: "aircraft category", mutate: func(item *flightstate.FlightState) { item.AircraftCategory++ }},
		{name: "aircraft category availability", mutate: func(item *flightstate.FlightState) { item.AircraftCategoryAvailable = false }},
		{name: "source", mutate: func(item *flightstate.FlightState) { item.SourceName = "opensky" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			second := base
			test.mutate(&second)
			result := RemoveExactDuplicates([]flightstate.FlightState{base, second})
			if result.DuplicateCount != 0 || len(result.UniqueStates) != 2 {
				t.Fatalf(
					"canonical payload difference collapsed: duplicates=%d unique=%d",
					result.DuplicateCount,
					len(result.UniqueStates),
				)
			}
		})
	}
}
