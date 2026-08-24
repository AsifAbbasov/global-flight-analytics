package readsbcompat

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
)

const int64BoundaryFloat64 = float64(1 << 63)

var ErrSourceNameRequired = errors.New(
	"readsb-compatible source name is required",
)

func OptionalGroundSpeed(
	value OptionalFloat64,
) (float64, bool) {
	if !value.Available || value.Value < 0 {
		return 0, false
	}
	return KnotsToMetersPerSecond(value.Value), true
}

func OptionalHeading(
	value OptionalFloat64,
) (float64, bool) {
	if !value.Available || value.Value < 0 || value.Value > 360 {
		return 0, false
	}
	return value.Value, true
}

func OptionalVerticalRate(
	value OptionalFloat64,
) (float64, bool) {
	if !value.Available {
		return 0, false
	}
	return FeetPerMinuteToMetersPerSecond(value.Value), true
}

func SafeUnixMilliseconds(
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

func SafeSeenDuration(
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

func ObservationTime(
	responseTime float64,
	seen OptionalFloat64,
) time.Time {
	base, ok := SafeUnixMilliseconds(responseTime)
	if !ok {
		return time.Time{}
	}
	age, ok := SafeSeenDuration(seen)
	if !ok {
		return base
	}
	return base.Add(-age).UTC()
}

func MapAircraft(
	sourceName string,
	item AircraftItem,
	responseTime float64,
) flightstate.FlightState {
	barometricAltitude := BarometricAltitudeReading(item.AltBaro)
	geometricAltitude := GeometricAltitudeReading(item.AltGeom)
	velocity, velocityAvailable := OptionalGroundSpeed(item.GroundSpeed)
	heading, headingAvailable := OptionalHeading(item.Track)
	verticalRate, verticalRateAvailable := OptionalVerticalRate(item.BaroRate)
	onGroundAvailable := barometricAltitude.Status == flightstate.AltitudeStatusGround ||
		barometricAltitude.Status == flightstate.AltitudeStatusObserved

	return flightstate.FlightState{
		ICAO24:                     strings.ToUpper(strings.TrimSpace(item.Hex)),
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
		ObservedAt:                 ObservationTime(responseTime, item.Seen),
		SourceName:                 strings.TrimSpace(sourceName),
	}
}

func AircraftItemRequiredFieldsValid(
	item AircraftItem,
	responseTime float64,
) bool {
	if _, ok := SafeUnixMilliseconds(responseTime); !ok {
		return false
	}
	if strings.TrimSpace(item.Hex) == "" {
		return false
	}
	if math.IsNaN(item.Latitude) || math.IsInf(item.Latitude, 0) ||
		item.Latitude < -90 || item.Latitude > 90 {
		return false
	}
	if math.IsNaN(item.Longitude) || math.IsInf(item.Longitude, 0) ||
		item.Longitude < -180 || item.Longitude > 180 {
		return false
	}
	return true
}

func MapStateResponseWithEvidence(
	sourceName string,
	response *StateResponse,
) (
	[]flightstate.FlightState,
	providerbatch.Evidence,
	error,
) {
	sourceName = strings.TrimSpace(sourceName)
	if sourceName == "" {
		return nil, providerbatch.Evidence{}, ErrSourceNameRequired
	}
	if response == nil {
		return []flightstate.FlightState{}, providerbatch.Evidence{}, nil
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
		if !AircraftItemRequiredFieldsValid(item, response.Now) {
			evidence.RejectedMalformed++
			continue
		}

		result = append(result, MapAircraft(sourceName, item, response.Now))
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
	sourceName string,
	response *StateResponse,
) []flightstate.FlightState {
	states, _, _ := MapStateResponseWithEvidence(sourceName, response)
	return states
}
