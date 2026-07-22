package flightstate

import (
	"math"
	"time"
)

type AltitudeStatus string

const (
	AltitudeStatusObserved    AltitudeStatus = "observed"
	AltitudeStatusGround      AltitudeStatus = "ground"
	AltitudeStatusUnknown     AltitudeStatus = "unknown"
	AltitudeStatusUnavailable AltitudeStatus = "unavailable"
	AltitudeStatusInvalid     AltitudeStatus = "invalid"
)

func ResolveAltitudeStatus(
	value float64,
	status AltitudeStatus,
) AltitudeStatus {
	if status != "" {
		if !IsKnownAltitudeStatus(status) {
			return AltitudeStatusInvalid
		}

		if status == AltitudeStatusObserved &&
			(math.IsNaN(value) || math.IsInf(value, 0)) {
			return AltitudeStatusInvalid
		}

		return status
	}

	if math.IsNaN(value) || math.IsInf(value, 0) {
		return AltitudeStatusInvalid
	}

	if value != 0 {
		return AltitudeStatusObserved
	}

	return AltitudeStatusUnavailable
}

func IsKnownAltitudeStatus(
	status AltitudeStatus,
) bool {
	switch status {
	case AltitudeStatusObserved,
		AltitudeStatusGround,
		AltitudeStatusUnknown,
		AltitudeStatusUnavailable,
		AltitudeStatusInvalid:
		return true

	default:
		return false
	}
}

type FlightState struct {
	ID                         string
	FlightID                   string
	AircraftID                 string
	IngestionRunID             string
	ICAO24                     string
	Callsign                   string
	Latitude                   float64
	Longitude                  float64
	BarometricAltitudeM        float64
	BarometricAltitudeStatus   AltitudeStatus
	GeometricAltitudeM         float64
	GeometricAltitudeStatus    AltitudeStatus
	VelocityMPS                float64
	VelocityAvailable          bool
	HeadingDegrees             float64
	HeadingAvailable           bool
	VerticalRateMPS            float64
	VerticalRateAvailable      bool
	OnGround                   bool
	OnGroundAvailable          bool
	TelemetryAvailabilityKnown bool
	OriginCountry              string
	SquawkCode                 string
	SpecialPurposeIndicator    bool
	PositionSource             PositionSource
	AircraftCategory           int
	AircraftCategoryAvailable  bool
	ObservedAt                 time.Time
	SourceName                 string
}

// OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2
