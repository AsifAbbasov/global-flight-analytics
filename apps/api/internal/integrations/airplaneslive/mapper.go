package airplaneslive

import (
	"math"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
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

func aircraftItemRequiredFieldsValid(
	item AircraftItem,
	responseTime float64,
) bool {
	if _, ok := safeUnixMilliseconds(responseTime); !ok {
		return false
	}
	if strings.TrimSpace(item.Hex) == "" {
		return false
	}
	if math.IsNaN(item.Latitude) ||
		math.IsInf(item.Latitude, 0) ||
		item.Latitude < -90 ||
		item.Latitude > 90 {
		return false
	}
	if math.IsNaN(item.Longitude) ||
		math.IsInf(item.Longitude, 0) ||
		item.Longitude < -180 ||
		item.Longitude > 180 {
		return false
	}
	return true
}

func MapStateResponseWithEvidence(
	response *StateResponse,
) (
	[]flightstate.FlightState,
	providerbatch.Evidence,
	error,
) {
	if response == nil {
		return []flightstate.FlightState{},
			providerbatch.Evidence{},
			nil
	}

	evidence := providerbatch.Evidence{
		Received: len(response.Aircraft),
	}
	result := make(
		[]flightstate.FlightState,
		0,
		len(response.Aircraft),
	)

	for _, item := range response.Aircraft {
		if !aircraftItemRequiredFieldsValid(
			item,
			response.Now,
		) {
			evidence.RejectedMalformed++
			continue
		}

		result = append(
			result,
			mapAircraft(item, response.Now),
		)
		evidence.Accepted++
	}

	if evidence.Received > 0 && evidence.Accepted == 0 {
		return result,
			evidence,
			providerbatch.NewAllItemsRejectedError(
				sourceName,
				evidence,
			)
	}

	return result, evidence, nil
}

func MapStateResponse(
	response *StateResponse,
) []flightstate.FlightState {
	states, _, _ := MapStateResponseWithEvidence(response)
	return states
}

// OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2
