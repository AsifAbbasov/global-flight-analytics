package airplaneslive

import (
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

const sourceName = "airplanes.live"

func mapAircraft(
	item AircraftItem,
	responseTime float64,
) flightstate.FlightState {
	barometricAltitude := barometricAltitudeReading(
		item.AltBaro,
	)
	geometricAltitude := geometricAltitudeReading(
		item.AltGeom,
	)

	return flightstate.FlightState{
		ICAO24:                     strings.ToUpper(item.Hex),
		Callsign:                   strings.TrimSpace(item.Flight),
		SquawkCode:                 strings.TrimSpace(item.Squawk),
		Latitude:                   item.Latitude,
		Longitude:                  item.Longitude,
		BarometricAltitudeM:        barometricAltitude.Meters,
		BarometricAltitudeStatus:   barometricAltitude.Status,
		GeometricAltitudeM:         geometricAltitude.Meters,
		GeometricAltitudeStatus:    geometricAltitude.Status,
		VelocityMPS:                knotsToMetersPerSecond(item.GroundSpeed),
		VelocityAvailable:          true,
		HeadingDegrees:             item.Track,
		HeadingAvailable:           true,
		VerticalRateMPS:            feetPerMinuteToMetersPerSecond(item.BaroRate),
		VerticalRateAvailable:      true,
		OnGround:                   barometricAltitude.Status == flightstate.AltitudeStatusGround,
		OnGroundAvailable:          true,
		TelemetryAvailabilityKnown: true,
		ObservedAt: time.UnixMilli(
			int64(responseTime),
		).Add(
			-time.Duration(item.Seen * float64(time.Second)),
		).UTC(),
		SourceName: sourceName,
	}
}

func MapStateResponse(response *StateResponse) []flightstate.FlightState {
	if response == nil {
		return []flightstate.FlightState{}
	}

	result := make(
		[]flightstate.FlightState,
		0,
		len(response.Aircraft),
	)

	for _, item := range response.Aircraft {
		result = append(
			result,
			mapAircraft(item, response.Now),
		)
	}

	return result
}

// OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2
