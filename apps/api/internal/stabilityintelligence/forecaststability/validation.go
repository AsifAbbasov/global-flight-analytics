package forecaststability

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

var fingerprintPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var versionIDPattern = regexp.MustCompile(`^forecast-version-[0-9a-f]{32}$`)

func ValidateVersionRecord(record ForecastVersionRecord, policy VersionPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if record.SchemaVersion != SchemaVersionV1 ||
		!versionIDPattern.MatchString(record.VersionID) ||
		record.Ordinal < 1 ||
		strings.TrimSpace(record.TrajectoryID) == "" ||
		record.ProjectionSchemaVersion != record.Projection.SchemaVersion ||
		methodIdentity(record.Method) != methodIdentity(record.Projection.Method) ||
		strings.TrimSpace(record.PolicyVersion) == "" ||
		strings.TrimSpace(record.ImplementationVersion) == "" ||
		!fingerprintPattern.MatchString(record.InputFingerprint) ||
		!fingerprintPattern.MatchString(record.OutputFingerprint) ||
		!fingerprintPattern.MatchString(record.DecisionFingerprint) ||
		record.CreatedAt.IsZero() ||
		record.CreatedAt.Before(record.Projection.GeneratedAt) ||
		record.ScopeGuard != ScopeGuardResearchOnly {
		return fmt.Errorf("%w: identity or chronology", ErrInvalidVersionRecord)
	}
	if record.Ordinal == 1 && record.ParentVersionID != "" {
		return fmt.Errorf("%w: initial version cannot have parent", ErrInvalidVersionRecord)
	}
	if record.Ordinal > 1 && !versionIDPattern.MatchString(record.ParentVersionID) {
		return fmt.Errorf("%w: successor parent required", ErrInvalidVersionRecord)
	}
	if record.TrajectoryID != record.Projection.TrajectoryID ||
		record.InputFingerprint != record.Projection.Provenance.InputFingerprint ||
		record.OutputFingerprint != projectionOutputFingerprint(record.Projection) ||
		record.DecisionFingerprint != decisionFingerprint(record.Projection, record.PolicyVersion, record.ImplementationVersion) ||
		record.VersionID != forecastVersionID(record.Ordinal, record.TrajectoryID, record.ParentVersionID, record.DecisionFingerprint) {
		return fmt.Errorf("%w: deterministic identity mismatch", ErrInvalidVersionRecord)
	}
	if len(record.Projection.Points) > policy.MaximumProjectionPointCount {
		return fmt.Errorf("%w: point capacity", ErrInvalidVersionRecord)
	}
	projectionReport := projectioncontract.Validate(record.Projection)
	if projectionReport.Status != projectioncontract.ValidationStatusValid {
		return fmt.Errorf("%w: projection issues=%v", ErrInvalidVersionRecord, projectionReport.Issues)
	}
	return nil
}

func ValidateRegistrationResult(result RegistrationResult, policy VersionPolicy) error {
	if result.SchemaVersion != SchemaVersionV1 ||
		!result.Status.IsKnown() ||
		!result.Decision.IsKnown() ||
		result.ScopeGuard != ScopeGuardResearchOnly ||
		result.GeneratedAt.IsZero() ||
		result.GeneratedAt.Before(result.Record.CreatedAt) ||
		len(result.Limitations) == 0 ||
		len(result.Explanations) == 0 {
		return fmt.Errorf("%w: contract", ErrInvalidRegistrationResult)
	}
	if err := ValidateVersionRecord(result.Record, policy); err != nil {
		return fmt.Errorf("%w: record: %v", ErrInvalidRegistrationResult, err)
	}
	for _, change := range result.Changes {
		if !change.Kind.IsKnown() || change.Previous == change.Current {
			return fmt.Errorf("%w: change", ErrInvalidRegistrationResult)
		}
	}
	switch result.Decision {
	case RegistrationDecisionInitial:
		if result.Record.Ordinal != 1 || result.Record.ParentVersionID != "" || len(result.Changes) != 0 {
			return fmt.Errorf("%w: initial decision", ErrInvalidRegistrationResult)
		}
	case RegistrationDecisionReused:
		if len(result.Changes) != 0 {
			return fmt.Errorf("%w: reused decision", ErrInvalidRegistrationResult)
		}
	case RegistrationDecisionCreated:
		if result.Record.Ordinal < 2 || result.Record.ParentVersionID == "" || len(result.Changes) == 0 {
			return fmt.Errorf("%w: successor decision", ErrInvalidRegistrationResult)
		}
	}
	return validateNarrative(result.Limitations, result.Explanations, ErrInvalidRegistrationResult)
}

func ValidateStabilityResult(result StabilityResult, policy StabilityPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if result.SchemaVersion != SchemaVersionV1 ||
		!result.Status.IsKnown() ||
		strings.TrimSpace(result.TrajectoryID) == "" ||
		!versionIDPattern.MatchString(result.BaselineVersionID) ||
		!versionIDPattern.MatchString(result.CandidateVersionID) ||
		!result.Level.IsKnown() ||
		!unitInterval(result.Score) ||
		result.ScopeGuard != ScopeGuardResearchOnly ||
		result.EvaluatedAt.IsZero() ||
		!fingerprintPattern.MatchString(result.Provenance.InputFingerprint) ||
		result.Provenance.BaselineVersionID != result.BaselineVersionID ||
		result.Provenance.CandidateVersionID != result.CandidateVersionID ||
		!fingerprintPattern.MatchString(result.Provenance.BaselineOutputFingerprint) ||
		!fingerprintPattern.MatchString(result.Provenance.CandidateOutputFingerprint) ||
		len(result.Components) == 0 ||
		len(result.Reasons) == 0 ||
		len(result.Limitations) == 0 ||
		len(result.Explanations) == 0 {
		return fmt.Errorf("%w: contract", ErrInvalidStabilityResult)
	}
	if result.Level == StabilityLevelIndeterminate {
		if result.Status != ResultStatusLimited || result.Score != 0 {
			return fmt.Errorf("%w: indeterminate contract", ErrInvalidStabilityResult)
		}
	} else if result.Status != ResultStatusComplete {
		return fmt.Errorf("%w: complete classification status", ErrInvalidStabilityResult)
	}
	weightTotal := 0.0
	for _, component := range result.Components {
		if strings.TrimSpace(component.Name) == "" || !unitInterval(component.Stability) || !nonNegativeFinite(component.Weight) || strings.TrimSpace(component.Explanation) == "" {
			return fmt.Errorf("%w: component", ErrInvalidStabilityResult)
		}
		weightTotal += component.Weight
	}
	if weightTotal < 0.999999 || weightTotal > 1.000001 {
		return fmt.Errorf("%w: component weight total", ErrInvalidStabilityResult)
	}
	for _, reason := range result.Reasons {
		if strings.TrimSpace(reason.Code) == "" || strings.TrimSpace(reason.Message) == "" || !unitInterval(reason.Impact) {
			return fmt.Errorf("%w: reason", ErrInvalidStabilityResult)
		}
	}
	if !unitInterval(result.Metrics.AlignedPointShare) ||
		result.Metrics.AlignedPointCount < 0 ||
		result.Metrics.AlignedPointCount > maxInt(result.Metrics.BaselinePointCount, result.Metrics.CandidatePointCount) ||
		!nonNegativeFinite(result.Metrics.MeanHorizontalShiftKilometers) ||
		!nonNegativeFinite(result.Metrics.MaximumHorizontalShiftKilometers) ||
		!nonNegativeFinite(result.Metrics.MeanAbsolutePointConfidenceDelta) ||
		!nonNegativeFinite(result.Metrics.AggregateConfidenceDelta) ||
		!nonNegativeFinite(result.Metrics.MeanRelativeHorizontalUncertaintyChange) ||
		!nonNegativeFinite(result.Metrics.ArrivalShiftSeconds) {
		return fmt.Errorf("%w: metrics", ErrInvalidStabilityResult)
	}
	if result.Provenance.InputFingerprint != stabilityInputFingerprint(result, policy) {
		return fmt.Errorf("%w: fingerprint mismatch", ErrInvalidStabilityResult)
	}
	return validateNarrative(result.Limitations, result.Explanations, ErrInvalidStabilityResult)
}

func validateNarrative(limitations []Limitation, explanations []Explanation, base error) error {
	for _, limitation := range limitations {
		if strings.TrimSpace(limitation.Code) == "" || strings.TrimSpace(limitation.Message) == "" || strings.TrimSpace(limitation.Scope) == "" {
			return fmt.Errorf("%w: limitation", base)
		}
	}
	for _, explanation := range explanations {
		if strings.TrimSpace(explanation.Code) == "" || strings.TrimSpace(explanation.Message) == "" {
			return fmt.Errorf("%w: explanation", base)
		}
	}
	return nil
}
