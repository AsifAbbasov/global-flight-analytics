package weathercontract

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type ValidationSeverity string

const (
	ValidationSeverityError   ValidationSeverity = "error"
	ValidationSeverityWarning ValidationSeverity = "warning"
)

type ValidationIssue struct {
	Code     string
	Message  string
	Path     string
	Severity ValidationSeverity
}

type ValidationReport struct {
	Status ValidationStatus
	Issues []ValidationIssue
}

func (report ValidationReport) HasCode(
	code string,
) bool {
	for _, issue := range report.Issues {
		if issue.Code == code {
			return true
		}
	}

	return false
}

func (report ValidationReport) Clone() ValidationReport {
	cloned := report
	cloned.Issues = append(
		[]ValidationIssue(nil),
		report.Issues...,
	)

	return cloned
}

const (
	IssueSchemaVersionInvalid       = "schema_version_invalid"
	IssueStatusInvalid              = "status_invalid"
	IssueTrajectoryIDRequired       = "trajectory_id_required"
	IssueAsOfTimeRequired           = "as_of_time_required"
	IssueGeneratedAtInvalid         = "generated_at_invalid"
	IssueScopeGuardInvalid          = "scope_guard_invalid"
	IssueConfidenceInvalid          = "confidence_invalid"
	IssueConfidenceReasonInvalid    = "confidence_reason_invalid"
	IssueUnavailableContractInvalid = "unavailable_contract_invalid"
	IssueAvailableSamplesRequired   = "available_samples_required"
	IssueLimitedContractInvalid     = "limited_contract_invalid"
	IssueSampleSequenceInvalid      = "sample_sequence_invalid"
	IssueSamplePositionInvalid      = "sample_position_invalid"
	IssueVerticalReferenceInvalid   = "vertical_reference_invalid"
	IssueSourceInvalid              = "source_invalid"
	IssueEvidenceKindInvalid        = "evidence_kind_invalid"
	IssueSourceResolutionInvalid    = "source_resolution_invalid"
	IssueSampleTimeInvalid          = "sample_time_invalid"
	IssueFutureEvidenceAvailability = "future_evidence_availability"
	IssueFutureNonForecastEvidence  = "future_non_forecast_evidence"
	IssueForecastTimeInvalid        = "forecast_time_invalid"
	IssueFeatureVectorEmpty         = "feature_vector_empty"
	IssueFeatureValueInvalid        = "feature_value_invalid"
	IssueSampleOrderInvalid         = "sample_order_invalid"
	IssueLimitationInvalid          = "limitation_invalid"
	IssueExplanationInvalid         = "explanation_invalid"
	IssueExplanationRequired        = "explanation_required"
	IssueFingerprintInvalid         = "fingerprint_invalid"
	IssueProvenanceSourceInvalid    = "provenance_source_invalid"
	IssueProvenanceSourceMismatch   = "provenance_source_mismatch"
	IssueLatestAvailableAtMismatch  = "latest_available_at_mismatch"
)

var fingerprintPattern = regexp.MustCompile(
	`^sha256:[0-9a-f]{64}$`,
)

func Validate(result Result) ValidationReport {
	collector := validationCollector{}

	collector.validateIdentity(result)
	collector.validateConfidence(
		result.Status,
		result.Confidence,
	)
	collector.validateStatusContract(result)
	collector.validateSamples(result)
	collector.validateLimitations(
		result.Limitations,
	)
	collector.validateExplanations(
		result.Explanations,
	)
	collector.validateProvenance(result)

	status := ValidationStatusValid
	if collector.hasErrors() {
		status = ValidationStatusInvalid
	}

	return ValidationReport{
		Status: status,
		Issues: append(
			[]ValidationIssue(nil),
			collector.issues...,
		),
	}
}

type validationCollector struct {
	issues []ValidationIssue
}

func (
	collector *validationCollector,
) addError(
	code string,
	path string,
	message string,
) {
	collector.issues = append(
		collector.issues,
		ValidationIssue{
			Code:     code,
			Message:  message,
			Path:     path,
			Severity: ValidationSeverityError,
		},
	)
}

func (
	collector validationCollector,
) hasErrors() bool {
	for _, issue := range collector.issues {
		if issue.Severity ==
			ValidationSeverityError {
			return true
		}
	}

	return false
}

func (
	collector *validationCollector,
) validateIdentity(result Result) {
	if result.SchemaVersion !=
		SchemaVersionV1 {
		collector.addError(
			IssueSchemaVersionInvalid,
			"schema_version",
			"schema version must be weather-feature-v1",
		)
	}
	if !result.Status.IsKnown() {
		collector.addError(
			IssueStatusInvalid,
			"status",
			"weather feature result status is unsupported",
		)
	}
	if strings.TrimSpace(
		result.TrajectoryID,
	) == "" {
		collector.addError(
			IssueTrajectoryIDRequired,
			"trajectory_id",
			"trajectory identifier is required",
		)
	}
	if result.AsOfTime.IsZero() {
		collector.addError(
			IssueAsOfTimeRequired,
			"as_of_time",
			"weather feature as-of time is required",
		)
	}
	if result.GeneratedAt.IsZero() ||
		(!result.AsOfTime.IsZero() &&
			result.GeneratedAt.Before(
				result.AsOfTime,
			)) {
		collector.addError(
			IssueGeneratedAtInvalid,
			"generated_at",
			"generated-at time must be at or after the as-of time",
		)
	}
	if result.ScopeGuard !=
		ScopeGuardContextOnly {
		collector.addError(
			IssueScopeGuardInvalid,
			"scope_guard",
			"weather output must be context only and must not claim proof of cause",
		)
	}
}

func (
	collector *validationCollector,
) validateConfidence(
	status ResultStatus,
	confidence Confidence,
) {
	if !unitInterval(confidence.Score) ||
		!confidence.Level.IsKnown() ||
		confidence.Level !=
			LevelForScore(confidence.Score) {
		collector.addError(
			IssueConfidenceInvalid,
			"confidence",
			"confidence score and level are invalid or inconsistent",
		)
	}

	if status != ResultStatusUnavailable &&
		len(confidence.Reasons) == 0 {
		collector.addError(
			IssueConfidenceReasonInvalid,
			"confidence.reasons",
			"available weather evidence requires at least one confidence reason",
		)
	}

	for index, reason := range confidence.Reasons {
		path := fmt.Sprintf(
			"confidence.reasons[%d]",
			index,
		)
		if strings.TrimSpace(
			reason.Code,
		) == "" ||
			strings.TrimSpace(
				reason.Message,
			) == "" ||
			!finite(
				reason.Contribution,
			) ||
			reason.Contribution < -1 ||
			reason.Contribution > 1 {
			collector.addError(
				IssueConfidenceReasonInvalid,
				path,
				"confidence reason requires code, message, and contribution between minus one and one",
			)
		}
	}
}

func (
	collector *validationCollector,
) validateStatusContract(result Result) {
	switch result.Status {
	case ResultStatusUnavailable:
		if len(result.Samples) != 0 ||
			result.Confidence.Score != 0 ||
			result.Confidence.Level !=
				ConfidenceLevelNone ||
			len(result.Limitations) == 0 ||
			len(result.Explanations) == 0 {
			collector.addError(
				IssueUnavailableContractInvalid,
				"status",
				"unavailable weather result requires no samples, no confidence, limitations, and explanations",
			)
		}
	case ResultStatusLimited:
		if len(result.Samples) == 0 {
			collector.addError(
				IssueAvailableSamplesRequired,
				"samples",
				"limited weather result requires at least one sample",
			)
		}
		if len(result.Limitations) == 0 {
			collector.addError(
				IssueLimitedContractInvalid,
				"limitations",
				"limited weather result requires at least one limitation",
			)
		}
	case ResultStatusComplete:
		if len(result.Samples) == 0 {
			collector.addError(
				IssueAvailableSamplesRequired,
				"samples",
				"complete weather result requires at least one sample",
			)
		}
	}

	if result.Status != ResultStatusUnavailable &&
		len(result.Explanations) == 0 {
		collector.addError(
			IssueExplanationRequired,
			"explanations",
			"available weather result requires an explanation",
		)
	}
}

func (
	collector *validationCollector,
) validateSamples(result Result) {
	var previousValidAt time.Time

	for index, sample := range result.Samples {
		path := fmt.Sprintf(
			"samples[%d]",
			index,
		)

		if sample.Sequence != index {
			collector.addError(
				IssueSampleSequenceInvalid,
				path+".sequence",
				"sample sequence must be zero-based and contiguous",
			)
		}

		collector.validatePosition(
			path+".position",
			sample.Position,
		)
		collector.validateSource(
			path+".source",
			sample.Source,
		)
		collector.validateSampleTimes(
			path,
			sample,
			result.AsOfTime,
			result.GeneratedAt,
		)
		collector.validateFeatures(
			path+".features",
			sample.Features,
		)

		if !previousValidAt.IsZero() &&
			sample.ValidAt.Before(
				previousValidAt,
			) {
			collector.addError(
				IssueSampleOrderInvalid,
				path+".valid_at",
				"samples must be ordered by non-decreasing valid time",
			)
		}
		previousValidAt = sample.ValidAt.UTC()
	}
}

func (
	collector *validationCollector,
) validatePosition(
	path string,
	position Position,
) {
	if !finite(position.Latitude) ||
		position.Latitude < -90 ||
		position.Latitude > 90 ||
		!finite(position.Longitude) ||
		position.Longitude < -180 ||
		position.Longitude > 180 {
		collector.addError(
			IssueSamplePositionInvalid,
			path,
			"weather sample latitude or longitude is invalid",
		)
	}

	if !position.VerticalReference.IsKnown() {
		collector.addError(
			IssueVerticalReferenceInvalid,
			path+".vertical_reference",
			"weather sample vertical reference is unsupported",
		)
	}

	if position.AltitudeMeters != nil &&
		(!finite(*position.AltitudeMeters) ||
			*position.AltitudeMeters < -1000 ||
			*position.AltitudeMeters > 100000) {
		collector.addError(
			IssueSamplePositionInvalid,
			path+".altitude_meters",
			"weather sample altitude is outside the supported contract range",
		)
	}
}

func (
	collector *validationCollector,
) validateSource(
	path string,
	source Source,
) {
	if strings.TrimSpace(source.Provider) == "" ||
		strings.TrimSpace(source.Dataset) == "" {
		collector.addError(
			IssueSourceInvalid,
			path,
			"weather source provider and dataset are required",
		)
	}
	if !source.EvidenceKind.IsKnown() {
		collector.addError(
			IssueEvidenceKindInvalid,
			path+".evidence_kind",
			"weather evidence kind is unsupported",
		)
	}
	if source.HorizontalResolutionKilometers != nil &&
		(!finite(
			*source.
				HorizontalResolutionKilometers,
		) ||
			*source.
				HorizontalResolutionKilometers <= 0) {
		collector.addError(
			IssueSourceResolutionInvalid,
			path+".horizontal_resolution_kilometers",
			"horizontal resolution must be finite and greater than zero",
		)
	}
	if source.TemporalResolution < 0 {
		collector.addError(
			IssueSourceResolutionInvalid,
			path+".temporal_resolution",
			"temporal resolution must not be negative",
		)
	}
}

func (
	collector *validationCollector,
) validateSampleTimes(
	path string,
	sample Sample,
	asOfTime time.Time,
	generatedAt time.Time,
) {
	if sample.ValidAt.IsZero() ||
		sample.AvailableAt.IsZero() ||
		sample.RetrievedAt.IsZero() {
		collector.addError(
			IssueSampleTimeInvalid,
			path,
			"valid-at, available-at, and retrieved-at times are required",
		)
		return
	}

	validAt := sample.ValidAt.UTC()
	availableAt := sample.AvailableAt.UTC()
	retrievedAt := sample.RetrievedAt.UTC()
	asOf := asOfTime.UTC()
	generated := generatedAt.UTC()

	if retrievedAt.Before(availableAt) ||
		(!generated.IsZero() &&
			retrievedAt.After(generated)) {
		collector.addError(
			IssueSampleTimeInvalid,
			path+".retrieved_at",
			"retrieved-at time must be between available-at and generated-at",
		)
	}

	if !asOf.IsZero() &&
		availableAt.After(asOf) {
		collector.addError(
			IssueFutureEvidenceAvailability,
			path+".available_at",
			"weather evidence available after the as-of time would leak future information",
		)
	}

	switch sample.Source.EvidenceKind {
	case EvidenceKindObservation,
		EvidenceKindAnalysis:
		if !asOf.IsZero() &&
			validAt.After(asOf) {
			collector.addError(
				IssueFutureNonForecastEvidence,
				path+".valid_at",
				"observation or analysis evidence must not be valid after the as-of time",
			)
		}
	case EvidenceKindForecast:
		if validAt.Before(availableAt) {
			collector.addError(
				IssueForecastTimeInvalid,
				path+".valid_at",
				"forecast valid time must not precede the time the forecast became available",
			)
		}
	}
}

func (
	collector *validationCollector,
) validateFeatures(
	path string,
	features FeatureVector,
) {
	if features.PresentCount() == 0 {
		collector.addError(
			IssueFeatureVectorEmpty,
			path,
			"weather sample requires at least one present feature",
		)
		return
	}

	collector.validateOptionalRange(
		path+".temperature_celsius",
		features.TemperatureCelsius,
		-150,
		100,
	)
	collector.validateOptionalRange(
		path+".relative_humidity_percent",
		features.RelativeHumidityPercent,
		0,
		100,
	)
	collector.validateOptionalRange(
		path+".precipitation_millimeters",
		features.PrecipitationMillimeters,
		0,
		math.MaxFloat64,
	)
	collector.validateOptionalRange(
		path+".rain_millimeters",
		features.RainMillimeters,
		0,
		math.MaxFloat64,
	)
	collector.validateOptionalRange(
		path+".cloud_cover_percent",
		features.CloudCoverPercent,
		0,
		100,
	)
	collector.validateOptionalRange(
		path+".surface_pressure_hpa",
		features.SurfacePressureHPA,
		100,
		1200,
	)
	collector.validateOptionalRange(
		path+".wind_speed_meters_per_second",
		features.WindSpeedMetersPerSecond,
		0,
		250,
	)
	collector.validateOptionalRangeExclusiveMaximum(
		path+".wind_direction_degrees",
		features.WindDirectionDegrees,
		0,
		360,
	)
	collector.validateOptionalRange(
		path+".wind_gusts_meters_per_second",
		features.WindGustsMetersPerSecond,
		0,
		250,
	)
}

func (
	collector *validationCollector,
) validateOptionalRange(
	path string,
	value *float64,
	minimum float64,
	maximum float64,
) {
	if value == nil {
		return
	}
	if !finite(*value) ||
		*value < minimum ||
		*value > maximum {
		collector.addError(
			IssueFeatureValueInvalid,
			path,
			fmt.Sprintf(
				"weather feature must be between %g and %g",
				minimum,
				maximum,
			),
		)
	}
}

func (
	collector *validationCollector,
) validateOptionalRangeExclusiveMaximum(
	path string,
	value *float64,
	minimum float64,
	maximum float64,
) {
	if value == nil {
		return
	}
	if !finite(*value) ||
		*value < minimum ||
		*value >= maximum {
		collector.addError(
			IssueFeatureValueInvalid,
			path,
			fmt.Sprintf(
				"weather feature must be at least %g and less than %g",
				minimum,
				maximum,
			),
		)
	}
}

func (
	collector *validationCollector,
) validateLimitations(
	limitations []Limitation,
) {
	for index, limitation := range limitations {
		if strings.TrimSpace(
			limitation.Code,
		) == "" ||
			strings.TrimSpace(
				limitation.Message,
			) == "" ||
			strings.TrimSpace(
				limitation.Scope,
			) == "" {
			collector.addError(
				IssueLimitationInvalid,
				fmt.Sprintf(
					"limitations[%d]",
					index,
				),
				"weather limitation requires code, message, and scope",
			)
		}
	}
}

func (
	collector *validationCollector,
) validateExplanations(
	explanations []Explanation,
) {
	for index, explanation := range explanations {
		if strings.TrimSpace(
			explanation.Code,
		) == "" ||
			strings.TrimSpace(
				explanation.Message,
			) == "" {
			collector.addError(
				IssueExplanationInvalid,
				fmt.Sprintf(
					"explanations[%d]",
					index,
				),
				"weather explanation requires code and message",
			)
		}
	}
}

func (
	collector *validationCollector,
) validateProvenance(result Result) {
	if !fingerprintPattern.MatchString(
		result.Provenance.InputFingerprint,
	) {
		collector.addError(
			IssueFingerprintInvalid,
			"provenance.input_fingerprint",
			"weather input fingerprint must be a sha256 fingerprint",
		)
	}

	sourceNames := append(
		[]string(nil),
		result.Provenance.SourceNames...,
	)
	if !sort.StringsAreSorted(sourceNames) {
		collector.addError(
			IssueProvenanceSourceInvalid,
			"provenance.source_names",
			"weather provenance source names must be sorted",
		)
	}

	seenSources := make(map[string]struct{})
	for index, sourceName := range sourceNames {
		normalized := strings.TrimSpace(
			sourceName,
		)
		if normalized == "" {
			collector.addError(
				IssueProvenanceSourceInvalid,
				fmt.Sprintf(
					"provenance.source_names[%d]",
					index,
				),
				"weather provenance source name is required",
			)
			continue
		}
		if _, exists := seenSources[normalized]; exists {
			collector.addError(
				IssueProvenanceSourceInvalid,
				fmt.Sprintf(
					"provenance.source_names[%d]",
					index,
				),
				"weather provenance source names must be unique",
			)
		}
		seenSources[normalized] = struct{}{}
	}

	expectedSources := make(map[string]struct{})
	latestAvailableAt := time.Time{}
	for _, sample := range result.Samples {
		provider := strings.TrimSpace(
			sample.Source.Provider,
		)
		if provider != "" {
			expectedSources[provider] = struct{}{}
		}
		if sample.AvailableAt.After(
			latestAvailableAt,
		) {
			latestAvailableAt =
				sample.AvailableAt.UTC()
		}
	}

	if len(expectedSources) != len(seenSources) {
		collector.addError(
			IssueProvenanceSourceMismatch,
			"provenance.source_names",
			"weather provenance source names do not match sample providers",
		)
	} else {
		for provider := range expectedSources {
			if _, exists := seenSources[provider]; !exists {
				collector.addError(
					IssueProvenanceSourceMismatch,
					"provenance.source_names",
					"weather provenance source names do not match sample providers",
				)
				break
			}
		}
	}

	if len(result.Samples) == 0 {
		if !result.Provenance.
			LatestAvailableAt.IsZero() {
			collector.addError(
				IssueLatestAvailableAtMismatch,
				"provenance.latest_available_at",
				"weather result without samples must not publish a latest available time",
			)
		}
	} else if result.Provenance.
		LatestAvailableAt.IsZero() ||
		!result.Provenance.
			LatestAvailableAt.Equal(
			latestAvailableAt,
		) {
		collector.addError(
			IssueLatestAvailableAtMismatch,
			"provenance.latest_available_at",
			"weather provenance latest available time must match the newest sample availability",
		)
	}
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func unitInterval(value float64) bool {
	return finite(value) &&
		value >= 0 &&
		value <= 1
}
