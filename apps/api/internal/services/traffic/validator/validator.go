package validator

import (
	"math"
	"regexp"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

var icao24Pattern = regexp.MustCompile(`^[A-F0-9]{6}$`)

func IsValidFlightState(item flightstate.FlightState, now time.Time) bool {
	if !icao24Pattern.MatchString(item.ICAO24) {
		return false
	}

	if item.Latitude < -90 || item.Latitude > 90 {
		return false
	}

	if item.Longitude < -180 || item.Longitude > 180 {
		return false
	}

	if !isFinite(item.BarometricAltitudeM) || item.BarometricAltitudeM < 0 {
		return false
	}

	if !isFinite(item.GeometricAltitudeM) || item.GeometricAltitudeM < 0 {
		return false
	}

	if !isFinite(item.VelocityMPS) || item.VelocityMPS < 0 {
		return false
	}

	if !isFinite(item.VerticalRateMPS) {
		return false
	}

	if !isFinite(item.HeadingDegrees) || item.HeadingDegrees < 0 || item.HeadingDegrees >= 360 {
		return false
	}

	if item.ObservedAt.IsZero() || item.ObservedAt.After(now) {
		return false
	}

	return true
}

func FilterValidFlightStates(items []flightstate.FlightState, now time.Time) []flightstate.FlightState {
	result := make([]flightstate.FlightState, 0, len(items))

	for _, item := range items {
		if IsValidFlightState(item, now) {
			result = append(result, item)
		}
	}

	return result
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
