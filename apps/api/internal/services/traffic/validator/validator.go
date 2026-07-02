package validator

import (
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

var icao24Pattern = regexp.MustCompile(`^[A-F0-9]{6}$`)

func EvaluateFlightState(item flightstate.FlightState, now time.Time) dataquality.DataQuality {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	score := 1.0
	hasCriticalError := false
	hasMovementError := false

	missingFields := make([]string, 0)
	warnings := make([]dataquality.Warning, 0)

	addMissingField := func(field string) {
		missingFields = append(missingFields, field)
	}

	addWarning := func(code string, message string, field string) {
		warnings = append(warnings, dataquality.Warning{
			Code:    code,
			Message: message,
			Field:   field,
		})
	}

	icao24 := strings.ToUpper(strings.TrimSpace(item.ICAO24))

	if icao24 == "" {
		addMissingField("icao24")
		addWarning("missing_icao24", "ICAO24 is required.", "icao24")
		hasCriticalError = true
		score -= 0.50
	} else if !icao24Pattern.MatchString(icao24) {
		addWarning("invalid_icao24", "ICAO24 must contain exactly 6 hexadecimal characters.", "icao24")
		hasCriticalError = true
		score -= 0.50
	}

	if !isFinite(item.Latitude) || item.Latitude < -90 || item.Latitude > 90 {
		addWarning("invalid_latitude", "Latitude must be finite and between -90 and 90.", "latitude")
		hasCriticalError = true
		score -= 0.35
	}

	if !isFinite(item.Longitude) || item.Longitude < -180 || item.Longitude > 180 {
		addWarning("invalid_longitude", "Longitude must be finite and between -180 and 180.", "longitude")
		hasCriticalError = true
		score -= 0.35
	}

	if item.ObservedAt.IsZero() {
		addMissingField("observed_at")
		addWarning("missing_observed_at", "ObservedAt is required.", "observed_at")
		hasCriticalError = true
		score -= 0.40
	} else if item.ObservedAt.After(now) {
		addWarning("future_observed_at", "ObservedAt cannot be in the future.", "observed_at")
		hasCriticalError = true
		score -= 0.40
	}

	if strings.TrimSpace(item.Callsign) == "" {
		addMissingField("callsign")
		score -= 0.05
	}

	if strings.TrimSpace(item.OriginCountry) == "" {
		addMissingField("origin_country")
		score -= 0.05
	}

	if strings.TrimSpace(item.SourceName) == "" {
		addMissingField("source_name")
		score -= 0.05
	}

	if !isFinite(item.BarometricAltitudeM) || item.BarometricAltitudeM < 0 {
		addWarning("invalid_barometric_altitude", "Barometric altitude must be finite and non-negative.", "barometric_altitude_m")
		score -= 0.10
	}

	if !isFinite(item.GeometricAltitudeM) || item.GeometricAltitudeM < 0 {
		addWarning("invalid_geometric_altitude", "Geometric altitude must be finite and non-negative.", "geometric_altitude_m")
		score -= 0.10
	}

	if !isFinite(item.VelocityMPS) || item.VelocityMPS < 0 {
		addWarning("invalid_velocity", "Velocity must be finite and non-negative.", "velocity_mps")
		hasMovementError = true
		score -= 0.15
	}

	if !isFinite(item.VerticalRateMPS) {
		addWarning("invalid_vertical_rate", "Vertical rate must be finite.", "vertical_rate_mps")
		score -= 0.05
	}

	if !isFinite(item.HeadingDegrees) || item.HeadingDegrees < 0 || item.HeadingDegrees >= 360 {
		addWarning("invalid_heading", "Heading must be finite and between 0 inclusive and 360 exclusive.", "heading_degrees")
		hasMovementError = true
		score -= 0.10
	}

	score = clampScore(score)

	if hasCriticalError {
		return dataquality.DataQuality{
			ValidationStatus: dataquality.ValidationStatusInvalid,
			Completeness:     dataquality.CompletenessLevelInsufficient,
			Confidence:       dataquality.ConfidenceLevelNone,
			Score:            0,
			MissingFields:    missingFields,
			Warnings:         warnings,
		}
	}

	completeness := dataquality.CompletenessLevelComplete

	if hasMovementError {
		completeness = dataquality.CompletenessLevelPositionOnly
	} else if len(missingFields) > 0 || len(warnings) > 0 {
		completeness = dataquality.CompletenessLevelPartial
	}

	validationStatus := dataquality.ValidationStatusValid

	if completeness != dataquality.CompletenessLevelComplete {
		validationStatus = dataquality.ValidationStatusPartial
	}

	return dataquality.DataQuality{
		ValidationStatus: validationStatus,
		Completeness:     completeness,
		Confidence:       confidenceFromScore(score),
		Score:            score,
		MissingFields:    missingFields,
		Warnings:         warnings,
	}
}

func IsValidFlightState(item flightstate.FlightState, now time.Time) bool {
	quality := EvaluateFlightState(item, now)

	return quality.ValidationStatus != dataquality.ValidationStatusInvalid
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

func confidenceFromScore(score float64) dataquality.ConfidenceLevel {
	switch {
	case score >= 0.85:
		return dataquality.ConfidenceLevelHigh
	case score >= 0.60:
		return dataquality.ConfidenceLevelMedium
	case score > 0:
		return dataquality.ConfidenceLevelLow
	default:
		return dataquality.ConfidenceLevelNone
	}
}

func clampScore(score float64) float64 {
	if score < 0 {
		return 0
	}

	if score > 1 {
		return 1
	}

	return score
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
