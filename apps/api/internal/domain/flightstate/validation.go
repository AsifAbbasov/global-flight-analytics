package flightstate

import (
	"errors"
	"math"
	"strings"
)

var (
	ErrFlightStateICAO24Required      = errors.New("flight state ICAO24 is required")
	ErrFlightStateCoordinatesInvalid  = errors.New("flight state coordinates are invalid")
	ErrFlightStateObservedAtRequired  = errors.New("flight state observation timestamp is required")
	ErrFlightStateSourceRequired      = errors.New("flight state source name is required")
	ErrFlightStateAltitudeInvalid     = errors.New("flight state altitude is invalid")
	ErrFlightStateVelocityInvalid     = errors.New("flight state velocity is invalid")
	ErrFlightStateHeadingInvalid      = errors.New("flight state heading is invalid")
	ErrFlightStateVerticalRateInvalid = errors.New("flight state vertical rate is invalid")
)

func (value FlightState) Validate() error {
	if strings.TrimSpace(value.ICAO24) == "" {
		return ErrFlightStateICAO24Required
	}
	if !finiteTelemetry(value.Latitude) || value.Latitude < -90 || value.Latitude > 90 ||
		!finiteTelemetry(value.Longitude) || value.Longitude < -180 || value.Longitude > 180 {
		return ErrFlightStateCoordinatesInvalid
	}
	if value.ObservedAt.IsZero() {
		return ErrFlightStateObservedAtRequired
	}
	if strings.TrimSpace(value.SourceName) == "" {
		return ErrFlightStateSourceRequired
	}
	if ResolveAltitudeStatus(value.BarometricAltitudeM, value.BarometricAltitudeStatus) == AltitudeStatusInvalid ||
		ResolveAltitudeStatus(value.GeometricAltitudeM, value.GeometricAltitudeStatus) == AltitudeStatusInvalid {
		return ErrFlightStateAltitudeInvalid
	}
	if value.VelocityAvailable && (!finiteTelemetry(value.VelocityMPS) || value.VelocityMPS < 0) {
		return ErrFlightStateVelocityInvalid
	}
	if value.HeadingAvailable && (!finiteTelemetry(value.HeadingDegrees) || value.HeadingDegrees < 0 || value.HeadingDegrees >= 360) {
		return ErrFlightStateHeadingInvalid
	}
	if value.VerticalRateAvailable && !finiteTelemetry(value.VerticalRateMPS) {
		return ErrFlightStateVerticalRateInvalid
	}
	return nil
}

func finiteTelemetry(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
