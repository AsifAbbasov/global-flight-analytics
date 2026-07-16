package proximityscanner

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
)

var fingerprintPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type ValidationReport struct {
	Status ValidationStatus
	Issues []string
}

func Validate(result Result, policy Policy) ValidationReport {
	issues := make([]string, 0)
	if err := policy.Validate(); err != nil {
		issues = append(issues, err.Error())
	}
	if result.SchemaVersion != SchemaVersionV1 {
		issues = append(issues, "schema_version")
	}
	if !result.Status.IsKnown() {
		issues = append(issues, "status")
	}
	if strings.TrimSpace(result.RegionCode) == "" ||
		!result.SceneStatus.IsKnown() ||
		result.AsOfTime.IsZero() ||
		result.GeneratedAt.IsZero() ||
		result.GeneratedAt.Before(result.AsOfTime) {
		issues = append(issues, "identity_or_times")
	}
	if result.ScopeGuard != ScopeGuardResearchOnly {
		issues = append(issues, "scope_guard")
	}
	if !unitInterval(result.Confidence.Score) ||
		!result.Confidence.Level.IsKnown() ||
		len(result.Confidence.Components) == 0 ||
		len(result.Confidence.Reasons) == 0 {
		issues = append(issues, "confidence")
	}
	if len(result.Limitations) == 0 || len(result.Explanations) == 0 {
		issues = append(issues, "explainability")
	}
	if !fingerprintPattern.MatchString(result.Provenance.InputFingerprint) ||
		!fingerprintPattern.MatchString(result.Provenance.SceneFingerprint) {
		issues = append(issues, "provenance_fingerprints")
	}

	candidateIDs := make(map[string]struct{}, len(result.Candidates))
	completeCount := 0
	limitedCount := 0
	withheldCount := 0
	for index, candidate := range result.Candidates {
		path := fmt.Sprintf("candidates[%d]", index)
		if candidate.ID != canonicalPairID(candidate.SourceNodeID, candidate.TargetNodeID) ||
			candidate.SourceNodeID == candidate.TargetNodeID ||
			!candidate.Status.IsKnown() ||
			!candidate.Kind.IsKnown() ||
			!nonNegativeFinite(candidate.HorizontalDistanceKilometers) ||
			!positiveFinite(candidate.EffectiveHorizontalRadiusKilometers) ||
			candidate.ObservationTimeDifference < 0 ||
			candidate.EvaluatedAt.IsZero() ||
			candidate.EvaluatedAt.After(result.GeneratedAt) ||
			!finite(candidate.ClosingRateMetersPerSecond) ||
			!unitInterval(candidate.Confidence.Score) ||
			!candidate.Confidence.Level.IsKnown() ||
			len(candidate.Confidence.Components) == 0 ||
			len(candidate.Confidence.Reasons) == 0 ||
			len(candidate.Limitations) == 0 ||
			len(candidate.Explanations) == 0 {
			issues = append(issues, path)
		}
		if candidate.HorizontalDistanceKilometers > candidate.EffectiveHorizontalRadiusKilometers {
			issues = append(issues, path+".horizontal_boundary")
		}
		if candidate.VerticalFilteringApplied {
			if candidate.VerticalSeparationMeters == nil ||
				candidate.EffectiveVerticalRadiusMeters == nil ||
				*candidate.VerticalSeparationMeters > *candidate.EffectiveVerticalRadiusMeters {
				issues = append(issues, path+".vertical_boundary")
			}
		} else {
			withheldCount++
			if candidate.VerticalSeparationMeters != nil ||
				candidate.EffectiveVerticalRadiusMeters != nil {
				issues = append(issues, path+".withheld_vertical_values")
			}
		}
		if _, exists := candidateIDs[candidate.ID]; exists {
			issues = append(issues, path+".duplicate")
		}
		candidateIDs[candidate.ID] = struct{}{}
		if candidate.Status == CandidateStatusComplete {
			completeCount++
		} else {
			limitedCount++
		}
	}

	metrics := result.Metrics
	if metrics.AircraftCount < 0 ||
		metrics.PossiblePairCount != possiblePairCount(metrics.AircraftCount) ||
		metrics.EvaluatedPairCount != metrics.PossiblePairCount ||
		metrics.CandidatePairCount != len(result.Candidates) ||
		metrics.CompleteCandidateCount != completeCount ||
		metrics.LimitedCandidateCount != limitedCount ||
		metrics.VerticalFilteringWithheldPairCount != withheldCount ||
		metrics.TemporalRejectedPairCount+metrics.HorizontalRejectedPairCount+
			metrics.VerticalRejectedPairCount+metrics.CandidatePairCount != metrics.EvaluatedPairCount ||
		!unitInterval(metrics.CandidateShare) {
		issues = append(issues, "metrics")
	}
	if metrics.PossiblePairCount == 0 && metrics.CandidateShare != 0 {
		issues = append(issues, "candidate_share_without_pairs")
	}

	graphReport := interactiongraph.Validate(result.Graph)
	if graphReport.Status != interactiongraph.ValidationStatusValid {
		issues = append(issues, "graph")
	}
	if result.Graph.RegionCode != result.RegionCode ||
		!result.Graph.AsOfTime.Equal(result.AsOfTime) ||
		result.Graph.Metrics.NodeCount != metrics.AircraftCount ||
		result.Graph.Metrics.EdgeCount != metrics.CandidatePairCount {
		issues = append(issues, "graph_consistency")
	}
	for _, edge := range result.Graph.Edges {
		if _, exists := candidateIDs[edge.ID]; !exists {
			issues = append(issues, "graph_edge_without_candidate")
		}
	}
	if result.Status != expectedStatus(result) {
		issues = append(issues, "status_consistency")
	}
	if len(issues) > 0 {
		return ValidationReport{Status: ValidationStatusInvalid, Issues: issues}
	}
	return ValidationReport{Status: ValidationStatusValid}
}

func expectedStatus(result Result) ResultStatus {
	if result.Metrics.AircraftCount == 0 {
		return ResultStatusUnavailable
	}
	if result.Metrics.AircraftCount < 2 ||
		result.SceneStatus != localtrafficscene.ResultStatusComplete ||
		result.Metrics.LimitedCandidateCount > 0 {
		return ResultStatusLimited
	}
	return ResultStatusComplete
}
