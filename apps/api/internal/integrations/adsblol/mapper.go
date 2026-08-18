package adsblol

import (
	"math"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
)

const (
	feetToMetersFactor          = 0.3048
	knotToMetersPerSecondFactor = 0.5144444444444445
	feetPerMinuteToMPSFactor    = 0.00508
)

func MapStateResponseWithEvidence(
	response *StateResponse,
) ([]flightstate.FlightState, providerbatch.Evidence, error) {
	if response == nil {
		return []flightstate.FlightState{}, providerbatch.Evidence{}, nil
	}

	evidence := providerbatch.Evidence{Received: len(response.Aircraft)}
	states := make([]flightstate.FlightState, 0, len(response.Aircraft))

	for _, item := range response.Aircraft {
		state, ok := mapAircraft(item, response.Now)
		if !ok {
			evidence.RejectedMalformed++
			continue
		}
		states = append(states, state)
		evidence.Accepted++
	}

	if evidence.Received > 0 && evidence.Accepted == 0 {
		return states, evidence, providerbatch.NewAllItemsRejectedError(
			sourceName,
			evidence,
		)
	}
	return states, evidence, nil
}

func MapStateResponse(response *StateResponse) []flightstate.FlightState {
	states, _, _ := MapStateResponseWithEvidence(response)
	return states
}

func mapAircraft(
	item AircraftItem,
	responseNow int64,
) (flightstate.FlightState, bool) {
	icao24, ok := canonicalICAO24(item.Hex)
	if !ok ||
		item.Latitude == nil ||
		item.Longitude == nil ||
		item.Seen == nil ||
		!finite(*item.Latitude) ||
		!finite(*item.Longitude) ||
		*item.Latitude < -90 ||
		*item.Latitude > 90 ||
		*item.Longitude < -180 ||
		*item.Longitude > 180 ||
		responseNow <= 0 {
		return flightstate.FlightState{}, false
	}

	seenDuration, ok := safeSeenDuration(*item.Seen)
	if !ok {
		return flightstate.FlightState{}, false
	}
	observedAt := time.UnixMilli(responseNow).
		Add(-seenDuration).
		UTC()

	baroMeters, baroStatus := mapBarometricAltitude(item.AltBaro)
	geomMeters, geomStatus := mapGeometricAltitude(item.AltGeom)

	velocity, velocityAvailable := mapGroundSpeed(item.GroundSpeed)
	heading, headingAvailable := mapHeading(item.Track)
	verticalRate, verticalRateAvailable := mapVerticalRate(item.BaroRate)
	onGroundAvailable := baroStatus == flightstate.AltitudeStatusGround ||
		baroStatus == flightstate.AltitudeStatusObserved

	return flightstate.FlightState{
		ICAO24:                     icao24,
		Callsign:                   optionalString(item.Flight),
		SquawkCode:                 optionalString(item.Squawk),
		Latitude:                   *item.Latitude,
		Longitude:                  *item.Longitude,
		BarometricAltitudeM:        baroMeters,
		BarometricAltitudeStatus:   baroStatus,
		GeometricAltitudeM:         geomMeters,
		GeometricAltitudeStatus:    geomStatus,
		VelocityMPS:                velocity,
		VelocityAvailable:          velocityAvailable,
		HeadingDegrees:             heading,
		HeadingAvailable:           headingAvailable,
		VerticalRateMPS:            verticalRate,
		VerticalRateAvailable:      verticalRateAvailable,
		OnGround:                   baroStatus == flightstate.AltitudeStatusGround,
		OnGroundAvailable:          onGroundAvailable,
		TelemetryAvailabilityKnown: true,
		ObservedAt:                 observedAt,
		SourceName:                 sourceName,
	}, true
}

func mapBarometricAltitude(
	value BarometricAltitude,
) (float64, flightstate.AltitudeStatus) {
	switch value.Kind {
	case BarometricAltitudeGround:
		return 0, flightstate.AltitudeStatusGround
	case BarometricAltitudeObserved:
		if finite(value.Feet) {
			return value.Feet * feetToMetersFactor, flightstate.AltitudeStatusObserved
		}
	}
	return 0, flightstate.AltitudeStatusUnavailable
}

func mapGeometricAltitude(value *float64) (float64, flightstate.AltitudeStatus) {
	if value == nil || !finite(*value) {
		return 0, flightstate.AltitudeStatusUnavailable
	}
	return *value * feetToMetersFactor, flightstate.AltitudeStatusObserved
}

func mapGroundSpeed(value *float64) (float64, bool) {
	if value == nil || !finite(*value) || *value < 0 {
		return 0, false
	}
	return *value * knotToMetersPerSecondFactor, true
}

func mapHeading(value *float64) (float64, bool) {
	if value == nil || !finite(*value) || *value < 0 || *value > 360 {
		return 0, false
	}
	return *value, true
}

func mapVerticalRate(value *float64) (float64, bool) {
	if value == nil || !finite(*value) {
		return 0, false
	}
	return *value * feetPerMinuteToMPSFactor, true
}

func canonicalICAO24(value string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if len(normalized) != 6 {
		return "", false
	}
	for _, character := range normalized {
		if (character < '0' || character > '9') &&
			(character < 'A' || character > 'F') {
			return "", false
		}
	}
	return normalized, true
}

func safeSeenDuration(value float64) (time.Duration, bool) {
	if !finite(value) || value < 0 {
		return 0, false
	}
	nanoseconds := value * float64(time.Second)
	if !finite(nanoseconds) ||
		nanoseconds > float64(math.MaxInt64) {
		return 0, false
	}
	return time.Duration(nanoseconds), true
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
