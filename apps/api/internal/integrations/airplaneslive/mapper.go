package airplaneslive

import (
	"math"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

const (
	sourceName           = "airplanes.live"
	int64BoundaryFloat64 = float64(1 << 63)
)

func optionalGroundSpeed(
	value OptionalFloat64,
) (float64, bool) {
	if !value.Available || value.Value < 0 {
		return 0, false
	}
	return knotsToMetersPerSecond(value.Value), true
}

func optionalHeading(
	value OptionalFloat64,
) (float64, bool) {
	if !value.Available || value.Value < 0 || value.Value > 360 {
		return 0, false
	}
	return value.Value, true
}

func optionalVerticalRate(
	value OptionalFloat64,
) (float64, bool) {
	if !value.Available {
		return 0, false
	}
	return feetPerMinuteToMetersPerSecond(value.Value), true
}

func safeUnixMilliseconds(
	value float64,
) (time.Time, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) ||
		value < -int64BoundaryFloat64 ||
		value >= int64BoundaryFloat64 ||
		math.Trunc(value) != value {
		return time.Time{}, false
	}
	return time.UnixMilli(int64(value)).UTC(), true
}

func safeSeenDuration(
	value OptionalFloat64,
) (time.Duration, bool) {
	if !value.Available || value.Value < 0 {
		return 0, false
	}
	nanoseconds := value.Value * float64(time.Second)
	if math.IsNaN(nanoseconds) || math.IsInf(nanoseconds, 0) ||
		nanoseconds >= int64BoundaryFloat64 {
		return 0, false
	}
	return time.Duration(nanoseconds), true
}

func observationTime(
	responseTime float64,
	seen OptionalFloat64,
) time.Time {
	base, ok := safeUnixMilliseconds(responseTime)
	if !ok {
		return time.Time{}
	}
	age, ok := safeSeenDuration(seen)
	if !ok {
		return base
	}
	return base.Add(-age).UTC()
}

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
	velocity, velocityAvailable := optionalGroundSpeed(item.GroundSpeed)
	heading, headingAvailable := optionalHeading(item.Track)
	verticalRate, verticalRateAvailable := optionalVerticalRate(item.BaroRate)
	onGroundAvailable := barometricAltitude.Status ==
		flightstate.AltitudeStatusGround ||
		barometricAltitude.Status == flightstate.AltitudeStatusObserved

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
		VelocityMPS:                velocity,
		VelocityAvailable:          velocityAvailable,
		HeadingDegrees:             heading,
		HeadingAvailable:           headingAvailable,
		VerticalRateMPS:            verticalRate,
		VerticalRateAvailable:      verticalRateAvailable,
		OnGround:                   barometricAltitude.Status == flightstate.AltitudeStatusGround,
		OnGroundAvailable:          onGroundAvailable,
		TelemetryAvailabilityKnown: true,
		ObservedAt:                 observationTime(responseTime, item.Seen),
		SourceName:                 sourceName,
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
