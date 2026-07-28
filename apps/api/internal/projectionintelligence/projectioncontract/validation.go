package projectioncontract

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

const ValidationVersion = "projection-intelligence-contract-validation-v2"

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
	Version string
	Status  ValidationStatus
	Issues  []ValidationIssue
}

func (report ValidationReport) HasCode(code string) bool {
	for _, issue := range report.Issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func (report ValidationReport) Clone() ValidationReport {
	cloned := report
	cloned.Issues = append([]ValidationIssue(nil), report.Issues...)
	return cloned
}

const (
	IssueSchemaVersionInvalid       = "schema_version_invalid"
	IssueStatusInvalid              = "status_invalid"
	IssueTrajectoryIDRequired       = "trajectory_id_required"
	IssueICAO24Invalid              = "icao24_invalid"
	IssueMethodNameRequired         = "method_name_required"
	IssueMethodVersionRequired      = "method_version_required"
	IssueDecisionClassInvalid       = "decision_class_invalid"
	IssueHorizonAsOfTimeRequired    = "horizon_as_of_time_required"
	IssueHorizonEndTimeInvalid      = "horizon_end_time_invalid"
	IssueHorizonStepInvalid         = "horizon_step_invalid"
	IssueHorizonGridInvalid         = "horizon_grid_invalid"
	IssueGeneratedAtInvalid         = "generated_at_invalid"
	IssueScopeGuardInvalid          = "scope_guard_invalid"
	IssueConfidenceInvalid          = "confidence_invalid"
	IssueConfidenceReasonRequired   = "confidence_reason_required"
	IssueConfidenceReasonInvalid    = "confidence_reason_invalid"
	IssueConfidenceReasonDuplicate  = "confidence_reason_duplicate"
	IssueConfidenceExceedsEvidence  = "confidence_exceeds_evidence"
	IssueUnavailableContractInvalid = "unavailable_contract_invalid"
	IssueLimitedContractInvalid     = "limited_contract_invalid"
	IssueUsablePointsRequired       = "usable_points_required"
	IssuePointSequenceInvalid       = "point_sequence_invalid"
	IssuePointTimeInvalid           = "point_time_invalid"
	IssuePointGridInvalid           = "point_grid_invalid"
	IssuePointPositionInvalid       = "point_position_invalid"
	IssuePointUncertaintyInvalid    = "point_uncertainty_invalid"
	IssueCompleteHorizonNotReached  = "complete_horizon_not_reached"
	IssueArrivalAirportInvalid      = "arrival_airport_invalid"
	IssueArrivalIntervalInvalid     = "arrival_interval_invalid"
	IssueLimitationInvalid          = "limitation_invalid"
	IssueLimitationDuplicate        = "limitation_duplicate"
	IssueExplanationInvalid         = "explanation_invalid"
	IssueExplanationDuplicate       = "explanation_duplicate"
	IssueExplanationRequired        = "explanation_required"
	IssueFingerprintRequired        = "fingerprint_required"
	IssueFingerprintInvalid         = "fingerprint_invalid"
	IssueInputRequired              = "input_required"
	IssueInputInvalid               = "input_invalid"
	IssueInputDuplicate             = "input_duplicate"
	IssueInputChronologyInvalid     = "input_chronology_invalid"
	IssueFutureInputEvidence        = "future_input_evidence"
	IssueLatestInputMismatch        = "latest_input_mismatch"
)

var (
	airportICAOPattern = regexp.MustCompile(`^[A-Z]{4}$`)
	icao24Pattern      = regexp.MustCompile(`^[0-9A-Fa-f]{6}$`)
	fingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	codePattern        = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

func Validate(result Result) ValidationReport {
	collector := validationCollector{}

	collector.validateIdentity(result)
	collector.validateMethod(result.Method)
	collector.validateHorizon(result.Horizon, result.GeneratedAt)
	collector.validateConfidence("confidence", result.Confidence)
	collector.validateLimitations("limitations", result.Limitations)
	collector.validateExplanations(result)
	collector.validateProvenance(result)
	collector.validateStatusContract(result)
	collector.validatePoints(result)
	collector.validateArrival(result)
	collector.validateConfidenceGraph(result)

	sort.SliceStable(
		collector.issues,
		func(left int, right int) bool {
			leftIssue := collector.issues[left]
			rightIssue := collector.issues[right]
			if leftIssue.Path != rightIssue.Path {
				return leftIssue.Path < rightIssue.Path
			}
			if leftIssue.Code != rightIssue.Code {
				return leftIssue.Code < rightIssue.Code
			}
			return leftIssue.Message < rightIssue.Message
		},
	)

	status := ValidationStatusValid
	if collector.hasErrors() {
		status = ValidationStatusInvalid
	}

	return ValidationReport{
		Version: ValidationVersion,
		Status:  status,
		Issues:  append([]ValidationIssue(nil), collector.issues...),
	}
}

type validationCollector struct {
	issues []ValidationIssue
}

func (collector *validationCollector) addError(
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

func (collector validationCollector) hasErrors() bool {
	for _, issue := range collector.issues {
		if issue.Severity == ValidationSeverityError {
			return true
		}
	}
	return false
}

func (collector *validationCollector) validateIdentity(result Result) {
	if result.SchemaVersion != SchemaVersionV1 {
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
	if strings.TrimSpace(result.TrajectoryID) == "" {
		collector.addError(
			IssueTrajectoryIDRequired,
			"trajectory_id",
			"trajectory id is required",
		)
	}
	if result.ICAO24 != "" &&
		(!icao24Pattern.MatchString(result.ICAO24) ||
			result.ICAO24 != strings.TrimSpace(result.ICAO24)) {
		collector.addError(
			IssueICAO24Invalid,
			"icao24",
			"ICAO24 must contain exactly six hexadecimal characters",
		)
	}
	if result.ScopeGuard != ScopeGuardResearchOnly {
		collector.addError(
			IssueScopeGuardInvalid,
			"scope_guard",
			"projection output must carry the research-only operational scope guard",
		)
	}
}

func (collector *validationCollector) validateMethod(method Method) {
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

func (collector *validationCollector) validateHorizon(
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
		!horizon.EndTime.After(horizon.AsOfTime) {
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
	} else if duration := horizon.Duration(); duration > 0 &&
		duration%horizon.Step != 0 {
		collector.addError(
			IssueHorizonGridInvalid,
			"horizon.step",
			"projection horizon duration must be exactly divisible by the forecast step",
		)
	}
	if generatedAt.IsZero() ||
		(!horizon.AsOfTime.IsZero() &&
			generatedAt.Before(horizon.AsOfTime)) {
		collector.addError(
			IssueGeneratedAtInvalid,
			"generated_at",
			"generated-at time must be present and must not be before the as-of time",
		)
	}
}

func (collector *validationCollector) validateConfidence(
	path string,
	value Confidence,
) {
	if !isUnitInterval(value.Score) ||
		!value.Level.IsKnown() ||
		(value.Score == 0 && value.Level != ConfidenceLevelNone) ||
		(value.Score > 0 && value.Level == ConfidenceLevelNone) {
		collector.addError(
			IssueConfidenceInvalid,
			path,
			"confidence score and ordinal level must be internally consistent",
		)
	}
	if value.Score > 0 && len(value.Reasons) == 0 {
		collector.addError(
			IssueConfidenceReasonRequired,
			path+".reasons",
			"positive confidence requires at least one reason",
		)
	}

	seenCodes := make(map[string]struct{}, len(value.Reasons))
	for index, reason := range value.Reasons {
		reasonPath := fmt.Sprintf("%s.reasons[%d]", path, index)
		code := strings.TrimSpace(reason.Code)
		message := strings.TrimSpace(reason.Message)

		if code == "" ||
			code != reason.Code ||
			!codePattern.MatchString(code) ||
			message == "" ||
			message != reason.Message ||
			!isFinite(reason.Contribution) ||
			reason.Contribution < -1 ||
			reason.Contribution > 1 {
			collector.addError(
				IssueConfidenceReasonInvalid,
				reasonPath,
				"confidence reason requires a normalized code, normalized message, and contribution between negative one and one",
			)
		}
		if _, exists := seenCodes[code]; exists {
			collector.addError(
				IssueConfidenceReasonDuplicate,
				reasonPath+".code",
				"confidence reason codes must be unique",
			)
		}
		seenCodes[code] = struct{}{}
	}
}

func (collector *validationCollector) validateLimitations(
	path string,
	items []Limitation,
) {
	seen := make(map[string]struct{}, len(items))
	for index, item := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, index)
		code := strings.TrimSpace(item.Code)
		message := strings.TrimSpace(item.Message)
		scope := strings.TrimSpace(item.Scope)

		if code == "" ||
			code != item.Code ||
			!codePattern.MatchString(code) ||
			message == "" ||
			message != item.Message ||
			scope == "" ||
			scope != item.Scope ||
			!codePattern.MatchString(scope) {
			collector.addError(
				IssueLimitationInvalid,
				itemPath,
				"limitation requires normalized code, message, and scope",
			)
		}

		key := scope + "\x00" + code
		if _, exists := seen[key]; exists {
			collector.addError(
				IssueLimitationDuplicate,
				itemPath,
				"limitation scope and code combinations must be unique",
			)
		}
		seen[key] = struct{}{}
	}
}

func (collector *validationCollector) validateExplanations(result Result) {
	if result.Status != ResultStatusUnavailable &&
		len(result.Explanations) == 0 {
		collector.addError(
			IssueExplanationRequired,
			"explanations",
			"usable projection output requires at least one explanation",
		)
	}

	seen := make(map[string]struct{}, len(result.Explanations))
	for index, item := range result.Explanations {
		path := fmt.Sprintf("explanations[%d]", index)
		code := strings.TrimSpace(item.Code)
		message := strings.TrimSpace(item.Message)

		if code == "" ||
			code != item.Code ||
			!codePattern.MatchString(code) ||
			message == "" ||
			message != item.Message {
			collector.addError(
				IssueExplanationInvalid,
				path,
				"explanation requires a normalized code and message",
			)
		}
		if _, exists := seen[code]; exists {
			collector.addError(
				IssueExplanationDuplicate,
				path+".code",
				"explanation codes must be unique",
			)
		}
		seen[code] = struct{}{}
	}
}

func (collector *validationCollector) validateProvenance(result Result) {
	usable := result.Status != ResultStatusUnavailable
	fingerprint := strings.TrimSpace(
		result.Provenance.InputFingerprint,
	)
	if usable && fingerprint == "" {
		collector.addError(
			IssueFingerprintRequired,
			"provenance.input_fingerprint",
			"usable projection output requires an input fingerprint",
		)
	}
	if fingerprint != "" &&
		(fingerprint != result.Provenance.InputFingerprint ||
			!fingerprintPattern.MatchString(fingerprint)) {
		collector.addError(
			IssueFingerprintInvalid,
			"provenance.input_fingerprint",
			"input fingerprint must use sha256 followed by sixty-four lowercase hexadecimal characters",
		)
	}
	if usable && len(result.Provenance.Inputs) == 0 {
		collector.addError(
			IssueInputRequired,
			"provenance.inputs",
			"usable projection output requires classified input references",
		)
	}

	latestObservedAt := time.Time{}
	seenNames := make(map[string]struct{}, len(result.Provenance.Inputs))
	for index, input := range result.Provenance.Inputs {
		path := fmt.Sprintf("provenance.inputs[%d]", index)
		collector.validateInputReference(
			result,
			input,
			path,
		)

		name := strings.TrimSpace(input.Name)
		if _, exists := seenNames[name]; exists {
			collector.addError(
				IssueInputDuplicate,
				path+".name",
				"provenance input names must be unique",
			)
		}
		seenNames[name] = struct{}{}

		if !input.ObservedAt.IsZero() {
			observedAt := input.ObservedAt.UTC()
			if latestObservedAt.IsZero() ||
				observedAt.After(latestObservedAt) {
				latestObservedAt = observedAt
			}
		}
	}

	declaredLatest := result.Provenance.LatestInputObservedAt
	if !latestObservedAt.IsZero() &&
		!declaredLatest.UTC().Equal(latestObservedAt) {
		collector.addError(
			IssueLatestInputMismatch,
			"provenance.latest_input_observed_at",
			"latest input observed time must equal the latest classified input observation",
		)
	}
	if latestObservedAt.IsZero() && !declaredLatest.IsZero() {
		collector.addError(
			IssueLatestInputMismatch,
			"provenance.latest_input_observed_at",
			"latest input observed time must be absent when no input observation exists",
		)
	}
	if !declaredLatest.IsZero() &&
		declaredLatest.After(result.Horizon.AsOfTime) {
		collector.addError(
			IssueFutureInputEvidence,
			"provenance.latest_input_observed_at",
			"latest input observed time must not be after the projection as-of time",
		)
	}
}

func (collector *validationCollector) validateInputReference(
	result Result,
	input InputReference,
	path string,
) {
	name := strings.TrimSpace(input.Name)
	source := strings.TrimSpace(input.SourceName)

	if name == "" ||
		name != input.Name ||
		!input.Classification.IsKnown() {
		collector.addError(
			IssueInputInvalid,
			path,
			"input reference requires a non-empty trimmed name and known classification",
		)
	}

	sourceRequired := input.Classification ==
		InputClassificationObserved ||
		input.Classification ==
			InputClassificationOpenlySourced ||
		input.Classification ==
			InputClassificationDerived ||
		input.Classification ==
			InputClassificationEstimated
	if sourceRequired &&
		(source == "" || source != input.SourceName) {
		collector.addError(
			IssueInputInvalid,
			path+".source_name",
			"observed, openly sourced, derived, and estimated inputs require a normalized source name",
		)
	}

	observationRequired := sourceRequired
	if observationRequired && input.ObservedAt.IsZero() {
		collector.addError(
			IssueInputInvalid,
			path+".observed_at",
			"observed, openly sourced, derived, and estimated inputs require an observation or analytical-basis time",
		)
	}
	if !input.ObservedAt.IsZero() &&
		input.ObservedAt.After(result.Horizon.AsOfTime) {
		collector.addError(
			IssueFutureInputEvidence,
			path+".observed_at",
			"input observation must not be after the projection as-of time",
		)
	}

	if !input.RetrievedAt.IsZero() {
		if !result.GeneratedAt.IsZero() &&
			input.RetrievedAt.After(result.GeneratedAt) {
			collector.addError(
				IssueInputInvalid,
				path+".retrieved_at",
				"input retrieval time must not be after generated-at time",
			)
		}
		if !input.ObservedAt.IsZero() &&
			input.RetrievedAt.Before(input.ObservedAt) {
			collector.addError(
				IssueInputChronologyInvalid,
				path+".retrieved_at",
				"input retrieval time must not precede its observation or analytical-basis time",
			)
		}
	}

	if (input.Classification == InputClassificationEstimated ||
		input.Classification == InputClassificationUnknown) &&
		strings.TrimSpace(input.Limitation) == "" {
		collector.addError(
			IssueInputInvalid,
			path+".limitation",
			"estimated or unknown input requires an explicit limitation",
		)
	}
}

func (collector *validationCollector) validateStatusContract(result Result) {
	switch result.Status {
	case ResultStatusUnavailable:
		if len(result.Points) != 0 ||
			result.Arrival != nil ||
			result.Confidence.Score != 0 ||
			result.Confidence.Level != ConfidenceLevelNone ||
			len(result.Limitations) == 0 {
			collector.addError(
				IssueUnavailableContractInvalid,
				"status",
				"unavailable result must not contain projection values and must explain its limitation",
			)
		}
	case ResultStatusLimited:
		if len(result.Points) == 0 {
			collector.addError(
				IssueUsablePointsRequired,
				"points",
				"limited result requires projection points",
			)
		}
		if !hasLimitedStatusEvidence(result.Limitations) {
			collector.addError(
				IssueLimitedContractInvalid,
				"status",
				"limited result requires a limitation beyond method assumptions and the research-only guard",
			)
		}
	case ResultStatusComplete:
		if len(result.Points) == 0 {
			collector.addError(
				IssueUsablePointsRequired,
				"points",
				"complete result requires projection points",
			)
		}
	}
}

func hasLimitedStatusEvidence(items []Limitation) bool {
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		scope := strings.TrimSpace(item.Scope)
		if code == "" || scope == "" {
			continue
		}
		if scope == "method" || code == "research_only" {
			continue
		}
		return true
	}
	return false
}

func (collector *validationCollector) validatePoints(result Result) {
	var previousTime time.Time
	for index, point := range result.Points {
		path := fmt.Sprintf("points[%d]", index)

		if point.Sequence != index {
			collector.addError(
				IssuePointSequenceInvalid,
				path+".sequence",
				"projection point sequence must be contiguous and zero-based",
			)
		}

		if point.ForecastTime.IsZero() ||
			!point.ForecastTime.After(result.Horizon.AsOfTime) ||
			point.ForecastTime.After(result.Horizon.EndTime) ||
			(!previousTime.IsZero() &&
				!point.ForecastTime.After(previousTime)) {
			collector.addError(
				IssuePointTimeInvalid,
				path+".forecast_time",
				"projection point time must be after as-of time and within the horizon",
			)
		}

		previousTime = point.ForecastTime

		if result.Horizon.Step > 0 &&
			!result.Horizon.AsOfTime.IsZero() {
			expectedTime := result.Horizon.AsOfTime.Add(
				time.Duration(index+1) *
					result.Horizon.Step,
			)
			if !point.ForecastTime.Equal(expectedTime) {
				collector.addError(
					IssuePointGridInvalid,
					path+".forecast_time",
					"projection point must occupy the exact slot defined by horizon as-of time and step",
				)
			}
		}

		if !validLatitude(point.Position.Latitude) ||
			!validLongitude(point.Position.Longitude) ||
			(point.Position.AltitudeM != nil &&
				!isFinite(*point.Position.AltitudeM)) {
			collector.addError(
				IssuePointPositionInvalid,
				path+".position",
				"projection point position must contain finite valid coordinates",
			)
		}

		if !isFinite(point.Uncertainty.HorizontalRadiusM) ||
			point.Uncertainty.HorizontalRadiusM <= 0 ||
			(point.Position.AltitudeM != nil &&
				(point.Uncertainty.VerticalRadiusM == nil ||
					!isFinite(*point.Uncertainty.VerticalRadiusM) ||
					*point.Uncertainty.VerticalRadiusM <= 0)) ||
			(point.Position.AltitudeM == nil &&
				point.Uncertainty.VerticalRadiusM != nil) {
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

	if result.Status == ResultStatusComplete &&
		len(result.Points) > 0 &&
		!result.Points[len(result.Points)-1].
			ForecastTime.Equal(result.Horizon.EndTime) {
		collector.addError(
			IssueCompleteHorizonNotReached,
			"points",
			"complete projection must reach the configured horizon end time",
		)
	}

	if result.Horizon.Step > 0 &&
		result.Horizon.Duration() > 0 &&
		result.Horizon.Duration()%result.Horizon.Step == 0 {
		expectedCount := int(
			result.Horizon.Duration() /
				result.Horizon.Step,
		)
		if len(result.Points) > expectedCount ||
			(result.Status == ResultStatusComplete &&
				len(result.Points) != expectedCount) {
			collector.addError(
				IssueHorizonGridInvalid,
				"points",
				"projection point count must fit the horizon grid and complete results must populate every slot",
			)
		}
	}
}

func (collector *validationCollector) validateArrival(result Result) {
	if result.Arrival == nil {
		return
	}
	arrival := result.Arrival

	if result.Status == ResultStatusUnavailable {
		collector.addError(
			IssueUnavailableContractInvalid,
			"arrival",
			"unavailable projection must not contain an arrival estimate",
		)
	}

	airport := strings.TrimSpace(arrival.AirportICAOCode)
	if airport != arrival.AirportICAOCode ||
		!airportICAOPattern.MatchString(airport) {
		collector.addError(
			IssueArrivalAirportInvalid,
			"arrival.airport_icao_code",
			"arrival airport must be a normalized four-letter ICAO location indicator",
		)
	}

	if arrival.EarliestTime.IsZero() ||
		arrival.EstimatedTime.IsZero() ||
		arrival.LatestTime.IsZero() ||
		arrival.EarliestTime.Before(result.Horizon.AsOfTime) ||
		arrival.EstimatedTime.Before(arrival.EarliestTime) ||
		arrival.LatestTime.Before(arrival.EstimatedTime) {
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

func (collector *validationCollector) validateConfidenceGraph(result Result) {
	if result.Status == ResultStatusUnavailable ||
		len(result.Points) == 0 ||
		!isUnitInterval(result.Confidence.Score) {
		return
	}

	weakestScore := result.Points[0].Confidence.Score
	for _, point := range result.Points[1:] {
		if point.Confidence.Score < weakestScore {
			weakestScore = point.Confidence.Score
		}
	}
	if result.Arrival != nil &&
		result.Arrival.Confidence.Score < weakestScore {
		weakestScore = result.Arrival.Confidence.Score
	}

	if result.Confidence.Score > weakestScore+
		confidenceComparisonTolerance {
		collector.addError(
			IssueConfidenceExceedsEvidence,
			"confidence.score",
			"overall confidence must not exceed the weakest mandatory point or arrival confidence",
		)
	}
}

const confidenceComparisonTolerance = 1e-12

func isFinite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func isUnitInterval(value float64) bool {
	return isFinite(value) &&
		value >= 0 &&
		value <= 1
}

func validLatitude(value float64) bool {
	return isFinite(value) &&
		value >= -90 &&
		value <= 90
}

func validLongitude(value float64) bool {
	return isFinite(value) &&
		value >= -180 &&
		value <= 180
}
