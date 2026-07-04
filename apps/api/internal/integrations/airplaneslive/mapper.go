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
	return flightstate.FlightState{
		ICAO24:              strings.ToUpper(item.Hex),
		Callsign:            strings.TrimSpace(item.Flight),
		Latitude:            item.Latitude,
		Longitude:           item.Longitude,
		BarometricAltitudeM: barometricAltitudeMeters(item.AltBaro),
		GeometricAltitudeM:  geometricAltitudeMeters(item.AltGeom),
		VelocityMPS:         knotsToMetersPerSecond(item.GroundSpeed),
		HeadingDegrees:      item.Track,
		VerticalRateMPS:     feetPerMinuteToMetersPerSecond(item.BaroRate),
		OnGround:            item.AltBaro == "ground",
		ObservedAt: time.UnixMilli(
			int64(responseTime),
		).Add(
			-time.Duration(item.Seen * float64(time.Second)),
		).UTC(),
		SourceName: sourceName,
	}
}

func geometricAltitudeMeters(altitudeFeet *float64) float64 {
	if altitudeFeet == nil {
		return 0
	}

	return feetToMeters(*altitudeFeet)
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
