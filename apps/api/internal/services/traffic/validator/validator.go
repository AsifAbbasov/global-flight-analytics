package validator

import (
	"regexp"
	"strings"
	"time"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

var icao24Pattern = regexp.MustCompile(`^[A-F0-9]{6}$`)

type assessmentCounter struct {
	assessed int
	passed   int
}

func (counter *assessmentCounter) Assess(
	condition bool,
) {
	counter.assessed++

	if condition {
		counter.passed++
	}
}

func (counter assessmentCounter) PassRatio() float64 {
	if counter.assessed == 0 {
		return 0
	}

	return float64(counter.passed) /
		float64(counter.assessed)
}

type altitudeAssessment struct {
	valid   bool
	missing bool
	code    string
	message string
}

func evaluateAltitude(
	value float64,
	status flightstate.AltitudeStatus,
	onGround bool,
	fieldName string,
) altitudeAssessment {
	effectiveStatus := flightstate.ResolveAltitudeStatus(
		value,
		status,
	)

	switch effectiveStatus {
	case flightstate.AltitudeStatusObserved:
		if aviationconstraints.IsNonNegativeFloat64(
			value,
		) {
			return altitudeAssessment{
				valid: true,
			}
		}

		return altitudeAssessment{
			code: "invalid_" + fieldName,
			message: fieldName +
				" must be finite and non-negative when status is observed.",
		}

	case flightstate.AltitudeStatusGround:
		if value == 0 && onGround {
			return altitudeAssessment{
				valid: true,
			}
		}

		return altitudeAssessment{
			code: "invalid_" + fieldName,
			message: fieldName +
				" ground status requires zero altitude and on_ground=true.",
		}

	case flightstate.AltitudeStatusUnknown,
		flightstate.AltitudeStatusUnavailable:
		return altitudeAssessment{
			missing: true,
		}

	case flightstate.AltitudeStatusInvalid:
		return altitudeAssessment{
			code: "invalid_" + fieldName,
			message: fieldName +
				" provider value is marked invalid.",
		}

	default:
		return altitudeAssessment{
			code: "invalid_" + fieldName,
			message: fieldName +
				" has an unsupported altitude status.",
		}
	}
}

func EvaluateFlightState(
	item flightstate.FlightState,
	now time.Time,
) dataquality.DataQuality {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	hasCriticalError := false
	hasMovementError := false

	missingFields := make(
		[]string,
		0,
	)

	warnings := make(
		[]dataquality.Warning,
		0,
	)

	assessment := assessmentCounter{}

	addMissingField := func(
		field string,
	) {
		missingFields = append(
			missingFields,
			field,
		)
	}

	addWarning := func(
		code string,
		message string,
		field string,
	) {
		warnings = append(
			warnings,
			dataquality.Warning{
				Code:    code,
				Message: message,
				Field:   field,
			},
		)
	}

	icao24 := strings.ToUpper(
		strings.TrimSpace(
			item.ICAO24,
		),
	)

	icao24Valid := icao24 != "" &&
		icao24Pattern.MatchString(
			icao24,
		)

	assessment.Assess(
		icao24Valid,
	)

	if icao24 == "" {
		addMissingField(
			"icao24",
		)

		addWarning(
			"missing_icao24",
			"ICAO24 is required.",
			"icao24",
		)

		hasCriticalError = true
	} else if !icao24Pattern.MatchString(
		icao24,
	) {
		addWarning(
			"invalid_icao24",
			"ICAO24 must contain exactly 6 hexadecimal characters.",
			"icao24",
		)

		hasCriticalError = true
	}

	latitudeValid := aviationconstraints.IsLatitude(
		item.Latitude,
	)

	assessment.Assess(
		latitudeValid,
	)

	if !latitudeValid {
		addWarning(
			"invalid_latitude",
			"Latitude must be finite and between -90 and 90.",
			"latitude",
		)

		hasCriticalError = true
	}

	longitudeValid := aviationconstraints.IsLongitude(
		item.Longitude,
	)

	assessment.Assess(
		longitudeValid,
	)

	if !longitudeValid {
		addWarning(
			"invalid_longitude",
			"Longitude must be finite and between -180 and 180.",
			"longitude",
		)

		hasCriticalError = true
	}

	observedAtValid := !item.ObservedAt.IsZero() &&
		!item.ObservedAt.After(
			now,
		)

	assessment.Assess(
		observedAtValid,
	)

	if item.ObservedAt.IsZero() {
		addMissingField(
			"observed_at",
		)

		addWarning(
			"missing_observed_at",
			"ObservedAt is required.",
			"observed_at",
		)

		hasCriticalError = true
	} else if item.ObservedAt.After(
		now,
	) {
		addWarning(
			"future_observed_at",
			"ObservedAt cannot be in the future.",
			"observed_at",
		)

		hasCriticalError = true
	}

	callsignPresent := strings.TrimSpace(
		item.Callsign,
	) != ""

	assessment.Assess(
		callsignPresent,
	)

	if !callsignPresent {
		addMissingField(
			"callsign",
		)
	}

	originCountryPresent := strings.TrimSpace(
		item.OriginCountry,
	) != ""

	assessment.Assess(
		originCountryPresent,
	)

	if !originCountryPresent {
		addMissingField(
			"origin_country",
		)
	}

	sourceNamePresent := strings.TrimSpace(
		item.SourceName,
	) != ""

	assessment.Assess(
		sourceNamePresent,
	)

	if !sourceNamePresent {
		addMissingField(
			"source_name",
		)
	}

	barometricAltitude := evaluateAltitude(
		item.BarometricAltitudeM,
		item.BarometricAltitudeStatus,
		item.OnGround,
		"barometric_altitude",
	)

	assessment.Assess(
		barometricAltitude.valid,
	)

	if barometricAltitude.missing {
		addMissingField(
			"barometric_altitude_m",
		)
	}

	if barometricAltitude.code != "" {
		addWarning(
			barometricAltitude.code,
			barometricAltitude.message,
			"barometric_altitude_m",
		)
	}

	geometricAltitude := evaluateAltitude(
		item.GeometricAltitudeM,
		item.GeometricAltitudeStatus,
		item.OnGround,
		"geometric_altitude",
	)

	assessment.Assess(
		geometricAltitude.valid,
	)

	if geometricAltitude.missing {
		addMissingField(
			"geometric_altitude_m",
		)
	}

	if geometricAltitude.code != "" {
		addWarning(
			geometricAltitude.code,
			geometricAltitude.message,
			"geometric_altitude_m",
		)
	}

	onGroundAvailable := item.HasOnGroundState()
	if !onGroundAvailable {
		addMissingField(
			"on_ground",
		)
		hasMovementError = true
	}

	velocityAvailable := item.HasVelocity()
	velocityValid := velocityAvailable &&
		aviationconstraints.IsNonNegativeFloat64(
			item.VelocityMPS,
		)

	assessment.Assess(
		velocityValid,
	)

	if !velocityAvailable {
		addMissingField(
			"velocity_mps",
		)
		hasMovementError = true
	} else if !velocityValid {
		addWarning(
			"invalid_velocity",
			"Velocity must be finite and non-negative.",
			"velocity_mps",
		)

		hasMovementError = true
	}

	verticalRateAvailable := item.HasVerticalRate()
	verticalRateValid := verticalRateAvailable &&
		aviationconstraints.IsFiniteFloat64(
			item.VerticalRateMPS,
		)

	assessment.Assess(
		verticalRateValid,
	)

	if !verticalRateAvailable {
		addMissingField(
			"vertical_rate_mps",
		)
		hasMovementError = true
	} else if !verticalRateValid {
		addWarning(
			"invalid_vertical_rate",
			"Vertical rate must be finite.",
			"vertical_rate_mps",
		)
		hasMovementError = true
	}

	headingAvailable := item.HasHeading()
	headingValid := headingAvailable &&
		aviationconstraints.IsHeadingDegreesExclusive(
			item.HeadingDegrees,
		)

	assessment.Assess(
		headingValid,
	)

	if !headingAvailable {
		addMissingField(
			"heading_degrees",
		)
		hasMovementError = true
	} else if !headingValid {
		addWarning(
			"invalid_heading",
			"Heading must be finite and between 0 inclusive and 360 exclusive.",
			"heading_degrees",
		)

		hasMovementError = true
	}

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
	} else if len(
		missingFields,
	) > 0 ||
		len(
			warnings,
		) > 0 {
		completeness = dataquality.CompletenessLevelPartial
	}

	validationStatus := dataquality.ValidationStatusValid

	if completeness != dataquality.CompletenessLevelComplete {
		validationStatus = dataquality.ValidationStatusPartial
	}

	return dataquality.DataQuality{
		ValidationStatus: validationStatus,
		Completeness:     completeness,
		Confidence: confidenceFromCompleteness(
			completeness,
		),
		Score:         assessment.PassRatio(),
		MissingFields: missingFields,
		Warnings:      warnings,
	}
}

func confidenceFromCompleteness(
	completeness dataquality.CompletenessLevel,
) dataquality.ConfidenceLevel {
	switch completeness {
	case dataquality.CompletenessLevelComplete:
		return dataquality.ConfidenceLevelHigh

	case dataquality.CompletenessLevelPartial:
		return dataquality.ConfidenceLevelMedium

	case dataquality.CompletenessLevelPositionOnly:
		return dataquality.ConfidenceLevelLow

	default:
		return dataquality.ConfidenceLevelNone
	}
}

func IsValidFlightState(
	item flightstate.FlightState,
	now time.Time,
) bool {
	quality := EvaluateFlightState(
		item,
		now,
	)

	return quality.ValidationStatus !=
		dataquality.ValidationStatusInvalid
}

func FilterValidFlightStates(
	items []flightstate.FlightState,
	now time.Time,
) []flightstate.FlightState {
	result := make(
		[]flightstate.FlightState,
		0,
		len(items),
	)

	for _, item := range items {
		if IsValidFlightState(
			item,
			now,
		) {
			result = append(
				result,
				item,
			)
		}
	}

	return result
}
