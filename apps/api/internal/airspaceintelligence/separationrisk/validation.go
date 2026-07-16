package separationrisk

import (
	"fmt"
	"strings"
)

type ValidationStatus string

const (
	ValidationStatusInvalid ValidationStatus = "invalid"
	ValidationStatusValid   ValidationStatus = "valid"
)

type ValidationReport struct {
	Status ValidationStatus
	Issues []string
}

func Validate(result Result, policy Policy) ValidationReport {
	issues := make([]string, 0)
	if result.SchemaVersion != SchemaVersionV1 {
		issues = append(issues, "schema version is invalid")
	}
	if !result.Status.IsKnown() || strings.TrimSpace(result.RegionCode) == "" || result.AsOfTime.IsZero() || result.GeneratedAt.IsZero() {
		issues = append(issues, "result identity, status, or times are invalid")
	}
	if result.GeneratedAt.Before(result.AsOfTime) {
		issues = append(issues, "generated-at time precedes as-of time")
	}
	if result.ScopeGuard != ScopeGuardResearchOnly {
		issues = append(issues, "scope guard is invalid")
	}
	if strings.TrimSpace(result.Provenance.InputFingerprint) == "" || strings.TrimSpace(result.Provenance.ScanFingerprint) == "" {
		issues = append(issues, "provenance fingerprints are required")
	}
	if len(result.Assessments) != result.Metrics.CandidateCount {
		issues = append(issues, "candidate metric does not match assessments")
	}
	if result.Metrics.CompleteAssessmentCount+result.Metrics.LimitedAssessmentCount != result.Metrics.CandidateCount {
		issues = append(issues, "assessment status metrics do not reconcile")
	}
	if result.Metrics.IndeterminateCount+result.Metrics.ContextualCount+result.Metrics.ElevatedCount+result.Metrics.HighCount != result.Metrics.CandidateCount {
		issues = append(issues, "risk level metrics do not reconcile")
	}
	if !result.Metrics.HighestDeterminateRiskLevel.IsKnown() {
		issues = append(issues, "highest determinate risk level is invalid")
	}
	if !unitInterval(result.Confidence.Score) || !result.Confidence.Level.IsKnown() {
		issues = append(issues, "result confidence is invalid")
	}
	seen := make(map[string]struct{}, len(result.Assessments))
	for index, assessment := range result.Assessments {
		path := fmt.Sprintf("assessments[%d]", index)
		if strings.TrimSpace(assessment.CandidateID) == "" || strings.TrimSpace(assessment.SourceNodeID) == "" || strings.TrimSpace(assessment.TargetNodeID) == "" {
			issues = append(issues, path+" identity is invalid")
		}
		if _, exists := seen[assessment.CandidateID]; exists {
			issues = append(issues, path+" candidate identifier is duplicated")
		}
		seen[assessment.CandidateID] = struct{}{}
		if !assessment.Status.IsKnown() || !assessment.Level.IsKnown() || !assessment.Kind.IsKnown() {
			issues = append(issues, path+" classification is invalid")
		}
		if !nonNegativeFinite(assessment.HorizontalDistanceKilometers) || assessment.ObservationTimeDifference < 0 || !finite(assessment.ClosingRateMetersPerSecond) {
			issues = append(issues, path+" measurements are invalid")
		}
		if assessment.HorizontalRadiusRatio == nil || !nonNegativeFinite(*assessment.HorizontalRadiusRatio) {
			issues = append(issues, path+" horizontal ratio is invalid")
		}
		if assessment.Status == AssessmentStatusLimited {
			if assessment.Level != RiskLevelIndeterminate || assessment.RiskScore != nil || assessment.VerticalRadiusRatio != nil {
				issues = append(issues, path+" limited assessment must withhold determinate risk values")
			}
		} else {
			if assessment.Level == RiskLevelIndeterminate || assessment.RiskScore == nil || assessment.VerticalRadiusRatio == nil || assessment.VerticalSeparationMeters == nil {
				issues = append(issues, path+" complete assessment requires determinate risk values")
			} else if !unitInterval(*assessment.RiskScore) || !nonNegativeFinite(*assessment.VerticalRadiusRatio) {
				issues = append(issues, path+" determinate risk values are invalid")
			}
		}
		if !unitInterval(assessment.Confidence.Score) || !assessment.Confidence.Level.IsKnown() || assessment.EvaluatedAt.IsZero() {
			issues = append(issues, path+" confidence or evaluated-at time is invalid")
		}
	}
	if len(result.Assessments) == 0 && result.Status != ResultStatusUnavailable {
		issues = append(issues, "empty result must be unavailable")
	}
	if len(result.Assessments) > 0 && result.Metrics.LimitedAssessmentCount > 0 && result.Status != ResultStatusLimited {
		issues = append(issues, "limited assessments require limited result status")
	}
	if len(result.Assessments) > 0 && result.Metrics.LimitedAssessmentCount == 0 && result.Status == ResultStatusUnavailable {
		issues = append(issues, "non-empty complete assessments cannot be unavailable")
	}
	if err := policy.Validate(); err != nil {
		issues = append(issues, "policy validation failed")
	}
	if len(issues) > 0 {
		return ValidationReport{Status: ValidationStatusInvalid, Issues: issues}
	}
	return ValidationReport{Status: ValidationStatusValid}
}
