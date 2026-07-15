package projectioncontract

import (
	"fmt"
	"math"
	"regexp"
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
	IssueMethodNameRequired         = "method_name_required"
	IssueMethodVersionRequired      = "method_version_required"
	IssueDecisionClassInvalid       = "decision_class_invalid"
	IssueHorizonAsOfTimeRequired    = "horizon_as_of_time_required"
	IssueHorizonEndTimeInvalid      = "horizon_end_time_invalid"
	IssueHorizonStepInvalid         = "horizon_step_invalid"
	IssueGeneratedAtInvalid         = "generated_at_invalid"
	IssueScopeGuardInvalid          = "scope_guard_invalid"
	IssueConfidenceInvalid          = "confidence_invalid"
	IssueConfidenceReasonInvalid    = "confidence_reason_invalid"
	IssueUnavailableContractInvalid = "unavailable_contract_invalid"
	IssueUsablePointsRequired       = "usable_points_required"
	IssuePointSequenceInvalid       = "point_sequence_invalid"
	IssuePointTimeInvalid           = "point_time_invalid"
	IssuePointPositionInvalid       = "point_position_invalid"
	IssuePointUncertaintyInvalid    = "point_uncertainty_invalid"
	IssueCompleteHorizonNotReached  = "complete_horizon_not_reached"
	IssueArrivalAirportInvalid      = "arrival_airport_invalid"
	IssueArrivalIntervalInvalid     = "arrival_interval_invalid"
	IssueLimitationInvalid          = "limitation_invalid"
	IssueExplanationInvalid         = "explanation_invalid"
	IssueExplanationRequired        = "explanation_required"
	IssueFingerprintRequired        = "fingerprint_required"
	IssueInputRequired              = "input_required"
	IssueInputInvalid               = "input_invalid"
	IssueFutureInputEvidence        = "future_input_evidence"
	IssueLatestInputMismatch        = "latest_input_mismatch"
)

var airportICAOPattern = regexp.MustCompile(
	`^[A-Z0-9]{4}$`,
)

func Validate(
	result Result,
) ValidationReport {
	collector := validationCollector{}

	collector.validateIdentity(result)
	collector.validateMethod(result.Method)
	collector.validateHorizon(
		result.Horizon,
		result.GeneratedAt,
	)
	collector.validateConfidence(
		"confidence",
		result.Confidence,
	)
	collector.validateLimitations(
		"limitations",
		result.Limitations,
	)
	collector.validateExplanations(
		result,
	)
	collector.validateProvenance(
		result,
	)
	collector.validateStatusContract(
		result,
	)
	collector.validatePoints(
		result,
	)
	collector.validateArrival(
		result,
	)

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
) validateIdentity(
	result Result,
) {
	if result.SchemaVersion !=
		SchemaVersionV1 {
		collector.addError(
			IssueSchemaVersionInvalid,
			"schema_version",
			"schema version must be projection-intelligence-v1",
		)
	}

	if !result.Status.IsKnown() {
		collector.addError(
			IssueStatusInvalid,
			"status",
			"result status is unsupported",
		)
	}

	if strings.TrimSpace(
		result.TrajectoryID,
	) == "" {
		collector.addError(
			IssueTrajectoryIDRequired,
			"trajectory_id",
			"trajectory id is required",
		)
	}

	if result.ScopeGuard !=
		ScopeGuardResearchOnly {
		collector.addError(
			IssueScopeGuardInvalid,
			"scope_guard",
			"projection output must carry the research-only operational scope guard",
		)
	}
}

func (
	collector *validationCollector,
) validateMethod(
	method Method,
) {
	if strings.TrimSpace(method.Name) == "" {
		collector.addError(
			IssueMethodNameRequired,
			"method.name",
			"projection method name is required",
		)
	}

	if strings.TrimSpace(method.Version) == "" {
		collector.addError(
			IssueMethodVersionRequired,
			"method.version",
			"projection method version is required",
		)
	}

	if !method.DecisionClass.IsKnown() {
		collector.addError(
			IssueDecisionClassInvalid,
			"method.decision_class",
			"projection method decision class is unsupported",
		)
	}
}

func (
	collector *validationCollector,
) validateHorizon(
	horizon Horizon,
	generatedAt time.Time,
) {
	if horizon.AsOfTime.IsZero() {
		collector.addError(
			IssueHorizonAsOfTimeRequired,
			"horizon.as_of_time",
			"projection as-of time is required",
		)
	}

	if horizon.EndTime.IsZero() ||
		!horizon.EndTime.After(
			horizon.AsOfTime,
		) {
		collector.addError(
			IssueHorizonEndTimeInvalid,
			"horizon.end_time",
			"projection end time must be after the as-of time",
		)
	}

	if horizon.Step <= 0 {
		collector.addError(
			IssueHorizonStepInvalid,
			"horizon.step",
			"projection step must be greater than zero",
		)
	}

	if generatedAt.IsZero() ||
		(!horizon.AsOfTime.IsZero() &&
			generatedAt.Before(
				horizon.AsOfTime,
			)) {
		collector.addError(
			IssueGeneratedAtInvalid,
			"generated_at",
			"generated-at time must be present and must not be before the as-of time",
		)
	}
}

func (
	collector *validationCollector,
) validateConfidence(
	path string,
	confidence Confidence,
) {
	if !isUnitInterval(
		confidence.Score,
	) ||
		!confidence.Level.IsKnown() ||
		(confidence.Score == 0 &&
			confidence.Level !=
				ConfidenceLevelNone) ||
		(confidence.Score > 0 &&
			confidence.Level ==
				ConfidenceLevelNone) {
		collector.addError(
			IssueConfidenceInvalid,
			path,
			"confidence score and level must be internally consistent",
		)
	}

	for index, reason := range confidence.Reasons {
		if strings.TrimSpace(reason.Code) == "" ||
			strings.TrimSpace(
				reason.Message,
			) == "" ||
			!isFinite(
				reason.Contribution,
			) {
			collector.addError(
				IssueConfidenceReasonInvalid,
				fmt.Sprintf(
					"%s.reasons[%d]",
					path,
					index,
				),
				"confidence reason requires code, message, and finite contribution",
			)
		}
	}
}

func (
	collector *validationCollector,
) validateLimitations(
	path string,
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
					"%s[%d]",
					path,
					index,
				),
				"limitation requires code, message, and scope",
			)
		}
	}
}

func (
	collector *validationCollector,
) validateExplanations(
	result Result,
) {
	if result.Status !=
		ResultStatusUnavailable &&
		len(result.Explanations) == 0 {
		collector.addError(
			IssueExplanationRequired,
			"explanations",
			"usable projection output requires at least one explanation",
		)
	}

	for index, explanation := range result.Explanations {
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
				"explanation requires code and message",
			)
		}
	}
}

func (
	collector *validationCollector,
) validateProvenance(
	result Result,
) {
	if result.Status !=
		ResultStatusUnavailable &&
		strings.TrimSpace(
			result.Provenance.
				InputFingerprint,
		) == "" {
		collector.addError(
			IssueFingerprintRequired,
			"provenance.input_fingerprint",
			"usable projection output requires an input fingerprint",
		)
	}

	if result.Status !=
		ResultStatusUnavailable &&
		len(result.Provenance.Inputs) == 0 {
		collector.addError(
			IssueInputRequired,
			"provenance.inputs",
			"usable projection output requires classified input references",
		)
	}

	var latestObservedAt time.Time
	for index, input := range result.Provenance.Inputs {
		path := fmt.Sprintf(
			"provenance.inputs[%d]",
			index,
		)

		if strings.TrimSpace(input.Name) == "" ||
			!input.Classification.IsKnown() {
			collector.addError(
				IssueInputInvalid,
				path,
				"input reference requires name and known classification",
			)
		}

		if input.Classification ==
			InputClassificationOpenlySourced &&
			strings.TrimSpace(
				input.SourceName,
			) == "" {
			collector.addError(
				IssueInputInvalid,
				path+".source_name",
				"openly sourced input requires a source name",
			)
		}

		if (input.Classification ==
			InputClassificationEstimated ||
			input.Classification ==
				InputClassificationUnknown) &&
			strings.TrimSpace(
				input.Limitation,
			) == "" {
			collector.addError(
				IssueInputInvalid,
				path+".limitation",
				"estimated or unknown input requires an explicit limitation",
			)
		}

		if !input.ObservedAt.IsZero() {
			observedAt := input.ObservedAt.UTC()
			if observedAt.After(
				result.Horizon.
					AsOfTime.UTC(),
			) {
				collector.addError(
					IssueFutureInputEvidence,
					path+".observed_at",
					"input observation must not be after the projection as-of time",
				)
			}
			if latestObservedAt.IsZero() ||
				observedAt.After(
					latestObservedAt,
				) {
				latestObservedAt = observedAt
			}
		}

		if !input.RetrievedAt.IsZero() &&
			!result.GeneratedAt.IsZero() &&
			input.RetrievedAt.After(
				result.GeneratedAt,
			) {
			collector.addError(
				IssueInputInvalid,
				path+".retrieved_at",
				"input retrieval time must not be after generated-at time",
			)
		}
	}

	if !latestObservedAt.IsZero() &&
		!result.Provenance.
			LatestInputObservedAt.UTC().
			Equal(latestObservedAt) {
		collector.addError(
			IssueLatestInputMismatch,
			"provenance.latest_input_observed_at",
			"latest input observed time must equal the latest classified input observation",
		)
	}

	if !result.Provenance.
		LatestInputObservedAt.IsZero() &&
		result.Provenance.
			LatestInputObservedAt.After(
			result.Horizon.AsOfTime,
		) {
		collector.addError(
			IssueFutureInputEvidence,
			"provenance.latest_input_observed_at",
			"latest input observed time must not be after the projection as-of time",
		)
	}
}

func (
	collector *validationCollector,
) validateStatusContract(
	result Result,
) {
	switch result.Status {
	case ResultStatusUnavailable:
		if len(result.Points) != 0 ||
			result.Arrival != nil ||
			result.Confidence.Score != 0 ||
			result.Confidence.Level !=
				ConfidenceLevelNone ||
			len(result.Limitations) == 0 {
			collector.addError(
				IssueUnavailableContractInvalid,
				"status",
				"unavailable result must not contain projection values and must explain its limitation",
			)
		}

	case ResultStatusLimited,
		ResultStatusComplete:
		if len(result.Points) == 0 {
			collector.addError(
				IssueUsablePointsRequired,
				"points",
				"limited or complete result requires projection points",
			)
		}
	}
}

func (
	collector *validationCollector,
) validatePoints(
	result Result,
) {
	var previousTime time.Time

	for index, point := range result.Points {
		path := fmt.Sprintf(
			"points[%d]",
			index,
		)

		if point.Sequence != index {
			collector.addError(
				IssuePointSequenceInvalid,
				path+".sequence",
				"projection point sequence must be contiguous and zero-based",
			)
		}

		if point.ForecastTime.IsZero() ||
			!point.ForecastTime.After(
				result.Horizon.AsOfTime,
			) ||
			point.ForecastTime.After(
				result.Horizon.EndTime,
			) ||
			(!previousTime.IsZero() &&
				!point.ForecastTime.After(
					previousTime,
				)) {
			collector.addError(
				IssuePointTimeInvalid,
				path+".forecast_time",
				"projection point time must be strictly ordered after as-of time and within the horizon",
			)
		}
		previousTime = point.ForecastTime

		if !validLatitude(
			point.Position.Latitude,
		) ||
			!validLongitude(
				point.Position.Longitude,
			) ||
			(point.Position.AltitudeM != nil &&
				!isFinite(
					*point.Position.
						AltitudeM,
				)) {
			collector.addError(
				IssuePointPositionInvalid,
				path+".position",
				"projection point position must contain finite valid coordinates",
			)
		}

		if !isFinite(
			point.Uncertainty.
				HorizontalRadiusM,
		) ||
			point.Uncertainty.
				HorizontalRadiusM <= 0 ||
			(point.Position.AltitudeM != nil &&
				(point.Uncertainty.
					VerticalRadiusM == nil ||
					!isFinite(
						*point.Uncertainty.
							VerticalRadiusM,
					) ||
					*point.Uncertainty.
						VerticalRadiusM <= 0)) ||
			(point.Position.AltitudeM == nil &&
				point.Uncertainty.
					VerticalRadiusM != nil) {
			collector.addError(
				IssuePointUncertaintyInvalid,
				path+".uncertainty",
				"estimated projection point requires positive explicit uncertainty matching its position dimensions",
			)
		}

		collector.validateConfidence(
			path+".confidence",
			point.Confidence,
		)
	}

	if result.Status ==
		ResultStatusComplete &&
		len(result.Points) > 0 &&
		!result.Points[len(result.Points)-1].
			ForecastTime.Equal(
			result.Horizon.EndTime,
		) {
		collector.addError(
			IssueCompleteHorizonNotReached,
			"points",
			"complete projection must reach the configured horizon end time",
		)
	}
}

func (
	collector *validationCollector,
) validateArrival(
	result Result,
) {
	if result.Arrival == nil {
		return
	}

	arrival := result.Arrival

	if result.Status ==
		ResultStatusUnavailable {
		collector.addError(
			IssueUnavailableContractInvalid,
			"arrival",
			"unavailable projection must not contain an arrival estimate",
		)
	}

	if !airportICAOPattern.MatchString(
		strings.TrimSpace(
			arrival.AirportICAOCode,
		),
	) {
		collector.addError(
			IssueArrivalAirportInvalid,
			"arrival.airport_icao_code",
			"arrival airport must be a normalized four-character ICAO code",
		)
	}

	if arrival.EarliestTime.IsZero() ||
		arrival.EstimatedTime.IsZero() ||
		arrival.LatestTime.IsZero() ||
		arrival.EarliestTime.Before(
			result.Horizon.AsOfTime,
		) ||
		arrival.EstimatedTime.Before(
			arrival.EarliestTime,
		) ||
		arrival.LatestTime.Before(
			arrival.EstimatedTime,
		) {
		collector.addError(
			IssueArrivalIntervalInvalid,
			"arrival",
			"arrival interval must satisfy as-of <= earliest <= estimated <= latest",
		)
	}

	collector.validateConfidence(
		"arrival.confidence",
		arrival.Confidence,
	)
	collector.validateLimitations(
		"arrival.limitations",
		arrival.Limitations,
	)
}

func isFinite(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func isUnitInterval(
	value float64,
) bool {
	return isFinite(value) &&
		value >= 0 &&
		value <= 1
}

func validLatitude(
	value float64,
) bool {
	return isFinite(value) &&
		value >= -90 &&
		value <= 90
}

func validLongitude(
	value float64,
) bool {
	return isFinite(value) &&
		value >= -180 &&
		value <= 180
}
