package validator

import (
	"regexp"
	"strings"
	"time"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/qualitypolicy"
)

var icao24Pattern = regexp.MustCompile(`^[A-F0-9]{6}$`)

func EvaluateFlightState(item flightstate.FlightState, now time.Time) dataquality.DataQuality {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	score := qualitypolicy.StartingScore
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
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.MissingICAO24Penalty)
	} else if !icao24Pattern.MatchString(icao24) {
		addWarning("invalid_icao24", "ICAO24 must contain exactly 6 hexadecimal characters.", "icao24")
		hasCriticalError = true
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.InvalidICAO24Penalty)
	}

	if !aviationconstraints.IsLatitude(item.Latitude) {
		addWarning("invalid_latitude", "Latitude must be finite and between -90 and 90.", "latitude")
		hasCriticalError = true
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.InvalidLatitudePenalty)
	}

	if !aviationconstraints.IsLongitude(item.Longitude) {
		addWarning("invalid_longitude", "Longitude must be finite and between -180 and 180.", "longitude")
		hasCriticalError = true
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.InvalidLongitudePenalty)
	}

	if item.ObservedAt.IsZero() {
		addMissingField("observed_at")
		addWarning("missing_observed_at", "ObservedAt is required.", "observed_at")
		hasCriticalError = true
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.MissingObservedAtPenalty)
	} else if item.ObservedAt.After(now) {
		addWarning("future_observed_at", "ObservedAt cannot be in the future.", "observed_at")
		hasCriticalError = true
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.FutureObservedAtPenalty)
	}

	if strings.TrimSpace(item.Callsign) == "" {
		addMissingField("callsign")
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.MissingCallsignPenalty)
	}

	if strings.TrimSpace(item.OriginCountry) == "" {
		addMissingField("origin_country")
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.MissingOriginCountryPenalty)
	}

	if strings.TrimSpace(item.SourceName) == "" {
		addMissingField("source_name")
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.MissingSourceNamePenalty)
	}

	if !aviationconstraints.IsNonNegativeFloat64(item.BarometricAltitudeM) {
		addWarning("invalid_barometric_altitude", "Barometric altitude must be finite and non-negative.", "barometric_altitude_m")
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.InvalidAltitudePenalty)
	}

	if !aviationconstraints.IsNonNegativeFloat64(item.GeometricAltitudeM) {
		addWarning("invalid_geometric_altitude", "Geometric altitude must be finite and non-negative.", "geometric_altitude_m")
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.InvalidAltitudePenalty)
	}

	if !aviationconstraints.IsNonNegativeFloat64(item.VelocityMPS) {
		addWarning("invalid_velocity", "Velocity must be finite and non-negative.", "velocity_mps")
		hasMovementError = true
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.InvalidVelocityPenalty)
	}

	if !aviationconstraints.IsFiniteFloat64(item.VerticalRateMPS) {
		addWarning("invalid_vertical_rate", "Vertical rate must be finite.", "vertical_rate_mps")
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.InvalidVerticalRatePenalty)
	}

	if !aviationconstraints.IsHeadingDegreesExclusive(item.HeadingDegrees) {
		addWarning("invalid_heading", "Heading must be finite and between 0 inclusive and 360 exclusive.", "heading_degrees")
		hasMovementError = true
		score = qualitypolicy.ApplyPenalty(score, qualitypolicy.InvalidHeadingPenalty)
	}

	score = qualitypolicy.ClampScore(score)

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
		Confidence:       qualitypolicy.ConfidenceFromScore(score),
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
