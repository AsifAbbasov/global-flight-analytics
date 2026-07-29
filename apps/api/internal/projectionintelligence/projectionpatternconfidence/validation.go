package projectionpatternconfidence

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

func validatePolicy(policy Policy) error {
	if policy.MinimumNeighborCount < 2 ||
		policy.TargetNeighborCount < policy.MinimumNeighborCount ||
		!positiveUnitInterval(policy.MinimumSimilarityScore) ||
		!positiveFinite(policy.MaximumSimilarityStandardDeviation) ||
		policy.MaximumSimilarityStandardDeviation > maximumUnitIntervalStandardDeviation ||
		!positiveFinite(policy.AnchorDistanceNormalizationKM) ||
		!positiveUnitInterval(policy.MinimumUsableScore) ||
		!positiveFinite(policy.MediumConfidenceMinimum) ||
		!positiveFinite(policy.HighConfidenceMinimum) ||
		policy.MediumConfidenceMinimum > policy.HighConfidenceMinimum ||
		policy.HighConfidenceMinimum > 1 ||
		policy.ContinuationAgreementSampleCount < 1 ||
		policy.ContinuationAgreementSampleCount > 32 ||
		!positiveFinite(policy.ContinuationDivergenceNormalizationMPS) ||
		!positiveFinite(policy.MaximumContinuationDivergenceMPS) ||
		policy.MaximumContinuationDivergenceMPS <
			policy.ContinuationDivergenceNormalizationMPS {
		return fmt.Errorf("pattern confidence policy snapshot is invalid")
	}
	weights := []float64{
		policy.SimilarityStrengthWeight,
		policy.SupportWeight,
		policy.SimilarityConsistencyWeight,
		policy.AnchorProximityWeight,
		policy.ContinuationAgreementWeight,
	}
	total := 0.0
	for _, weight := range weights {
		if !positiveFinite(weight) {
			return fmt.Errorf("pattern confidence policy component weight is invalid")
		}
		total += weight
	}
	if math.Abs(total-1) > scoreComparisonTolerance {
		return fmt.Errorf("pattern confidence policy component weights do not sum to one")
	}
	return nil
}

func validateResultCounts(result Result) error {
	if result.NeighborCount < 0 ||
		result.TargetNeighborCount != result.Policy.TargetNeighborCount ||
		result.NeighborCount != len(result.SelectedTrajectoryIDs) {
		return fmt.Errorf("pattern confidence neighbor counts are invalid")
	}
	return nil
}

func validateAggregateMeasurements(result Result) error {
	if !unitInterval(result.MeanSimilarityScore) ||
		!unitInterval(result.MinimumSimilarityScore) ||
		result.MinimumSimilarityScore > result.MeanSimilarityScore ||
		!finite(result.SimilarityStandardDeviation) ||
		result.SimilarityStandardDeviation < 0 ||
		result.SimilarityStandardDeviation > maximumUnitIntervalStandardDeviation ||
		!finite(result.MeanAnchorDistanceKM) ||
		result.MeanAnchorDistanceKM < 0 ||
		result.MeanCandidateAgeSeconds != 0 {
		return fmt.Errorf("pattern confidence aggregate measurements are invalid")
	}
	if result.NeighborCount == 0 &&
		(result.MeanSimilarityScore != 0 ||
			result.MinimumSimilarityScore != 0 ||
			result.SimilarityStandardDeviation != 0 ||
			result.MeanAnchorDistanceKM != 0) {
		return fmt.Errorf("empty pattern confidence must have zero aggregate measurements")
	}
	if result.NeighborCount > 0 && result.MinimumSimilarityScore <= 0 {
		return fmt.Errorf("non-empty pattern confidence requires positive minimum similarity")
	}
	return nil
}

func validateContinuationAgreement(result Result) error {
	if !result.ContinuationAgreementKnown {
		if result.ContinuationAgreementSampleCount != 0 ||
			result.ContinuationAgreementPairCount != 0 ||
			result.ContinuationComparisonCount != 0 ||
			result.ContinuationHorizonSeconds != 0 ||
			result.MeanContinuationSpreadM != 0 ||
			result.MaximumContinuationSpreadM != 0 ||
			result.MeanContinuationDivergenceMPS != 0 ||
			result.MaximumContinuationDivergenceMPS != 0 {
			return fmt.Errorf("unknown continuation agreement must have zero measurements")
		}
		return nil
	}

	expectedPairs := result.NeighborCount * (result.NeighborCount - 1) / 2
	expectedComparisons := expectedPairs * result.Policy.ContinuationAgreementSampleCount
	if result.NeighborCount < 2 ||
		result.ContinuationAgreementSampleCount !=
			result.Policy.ContinuationAgreementSampleCount ||
		result.ContinuationAgreementPairCount != expectedPairs ||
		result.ContinuationComparisonCount != expectedComparisons ||
		!positiveFinite(result.ContinuationHorizonSeconds) ||
		!nonNegativeFinite(result.MeanContinuationSpreadM) ||
		!nonNegativeFinite(result.MaximumContinuationSpreadM) ||
		result.MaximumContinuationSpreadM < result.MeanContinuationSpreadM ||
		!nonNegativeFinite(result.MeanContinuationDivergenceMPS) ||
		!nonNegativeFinite(result.MaximumContinuationDivergenceMPS) ||
		result.MaximumContinuationDivergenceMPS <
			result.MeanContinuationDivergenceMPS {
		return fmt.Errorf("continuation agreement measurements are invalid")
	}
	minimumElapsed := result.ContinuationHorizonSeconds /
		float64(result.ContinuationAgreementSampleCount)
	tolerance := scoreComparisonTolerance * math.Max(1, result.MaximumContinuationSpreadM)
	if result.MeanContinuationSpreadM+tolerance <
		result.MeanContinuationDivergenceMPS*minimumElapsed ||
		result.MeanContinuationSpreadM-tolerance >
			result.MeanContinuationDivergenceMPS*result.ContinuationHorizonSeconds ||
		result.MaximumContinuationSpreadM-tolerance >
			result.MaximumContinuationDivergenceMPS*result.ContinuationHorizonSeconds ||
		result.MaximumContinuationDivergenceMPS-tolerance >
			result.MaximumContinuationSpreadM/minimumElapsed {
		return fmt.Errorf("continuation spread and divergence measurements are inconsistent")
	}
	return nil
}

func validateComponents(result Result) error {
	if len(result.Components) != len(canonicalComponentNames) {
		return fmt.Errorf("pattern confidence requires the canonical component catalog")
	}

	expectedScores := []float64{
		clampUnit(result.MeanSimilarityScore),
		clampUnit(float64(result.NeighborCount) / float64(result.Policy.TargetNeighborCount)),
		clampUnit(1 - result.SimilarityStandardDeviation/maximumUnitIntervalStandardDeviation),
		clampUnit(1 - result.MeanAnchorDistanceKM/result.Policy.AnchorDistanceNormalizationKM),
		0,
	}
	if result.ContinuationAgreementKnown {
		expectedScores[4] = clampUnit(
			1 - result.MeanContinuationDivergenceMPS/
				result.Policy.ContinuationDivergenceNormalizationMPS,
		)
	}

	expectedWeights := []float64{
		result.Policy.SimilarityStrengthWeight,
		result.Policy.SupportWeight,
		result.Policy.SimilarityConsistencyWeight,
		result.Policy.AnchorProximityWeight,
		result.Policy.ContinuationAgreementWeight,
	}
	weightTotal := 0.0
	weightedScore := 0.0
	for index, component := range result.Components {
		if component.Name != canonicalComponentNames[index] {
			return fmt.Errorf(
				"pattern confidence component catalog is invalid at index %d",
				index,
			)
		}
		if !unitInterval(component.Score) || !positiveFinite(component.Weight) {
			return fmt.Errorf("pattern confidence component is invalid")
		}
		if math.Abs(component.Weight-expectedWeights[index]) > scoreComparisonTolerance {
			return fmt.Errorf(
				"pattern confidence component %q weight does not match policy",
				component.Name,
			)
		}
		if math.Abs(component.Score-expectedScores[index]) > scoreComparisonTolerance {
			return fmt.Errorf(
				"pattern confidence component %q does not match aggregate evidence",
				component.Name,
			)
		}
		weightTotal += component.Weight
		weightedScore += component.Score * component.Weight
	}
	if math.Abs(weightTotal-1) > scoreComparisonTolerance {
		return fmt.Errorf("pattern confidence component weights do not sum to one")
	}
	if math.Abs(result.Score-clampUnit(weightedScore)) > scoreComparisonTolerance {
		return fmt.Errorf("pattern confidence score does not match weighted components")
	}
	return nil
}

func validateDecisionSemantics(result Result) error {
	if !unitInterval(result.Score) || !result.Level.IsKnown() {
		return fmt.Errorf("pattern confidence score and level are invalid")
	}

	supportSufficient := result.NeighborCount >= result.Policy.MinimumNeighborCount
	similarityFloorSufficient := result.NeighborCount > 0 &&
		result.MinimumSimilarityScore >= result.Policy.MinimumSimilarityScore
	dispersionAcceptable := result.NeighborCount > 0 &&
		result.SimilarityStandardDeviation <=
			result.Policy.MaximumSimilarityStandardDeviation
	agreementAcceptable := result.ContinuationAgreementKnown &&
		result.MaximumContinuationDivergenceMPS <=
			result.Policy.MaximumContinuationDivergenceMPS
	scoreSufficient := result.Score >= result.Policy.MinimumUsableScore

	expectedUsable := supportSufficient &&
		similarityFloorSufficient &&
		dispersionAcceptable &&
		agreementAcceptable &&
		scoreSufficient
	if result.Usable != expectedUsable {
		return fmt.Errorf("pattern confidence usable decision does not match policy evidence")
	}

	expectedStatus := StatusUnavailable
	if expectedUsable {
		expectedStatus = StatusLimited
		if result.NeighborCount >= result.Policy.TargetNeighborCount &&
			result.SelectionStatus == projectionneighbors.StatusComplete {
			expectedStatus = StatusComplete
		}
	}
	if result.Status != expectedStatus {
		return fmt.Errorf("pattern confidence status does not match policy evidence")
	}

	expectedLevel := confidenceLevelForDecision(
		result.Score,
		result.Status,
		result.Policy.MediumConfidenceMinimum,
		result.Policy.HighConfidenceMinimum,
	)
	if result.Level != expectedLevel {
		return fmt.Errorf("pattern confidence level does not match score thresholds")
	}

	if result.Status == StatusUnavailable && len(result.Limitations) == 0 {
		return fmt.Errorf("unavailable pattern confidence must explain limitations")
	}
	coreNotices := []struct {
		code     string
		required bool
	}{
		{"insufficient_historical_neighbor_support", !supportSufficient},
		{"pattern_similarity_floor_below_minimum", result.NeighborCount > 0 && !similarityFloorSufficient},
		{"pattern_similarity_dispersion_above_maximum", result.NeighborCount > 0 && !dispersionAcceptable},
		{"pattern_continuation_agreement_unavailable", !result.ContinuationAgreementKnown},
		{"pattern_continuation_divergence_above_maximum", result.ContinuationAgreementKnown && !agreementAcceptable},
		{"pattern_confidence_below_minimum", !scoreSufficient},
		{"pattern_support_partial", expectedStatus == StatusLimited},
	}
	for _, notice := range coreNotices {
		present := hasNotice(result.Limitations, notice.code)
		if present != notice.required {
			return fmt.Errorf(
				"pattern confidence limitation %q does not match decision evidence",
				notice.code,
			)
		}
	}
	return nil
}

func validateSelectedTrajectoryIDs(result Result) error {
	seen := make(map[string]struct{}, len(result.SelectedTrajectoryIDs))
	for _, trajectoryID := range result.SelectedTrajectoryIDs {
		normalized := strings.TrimSpace(trajectoryID)
		if normalized == "" || trajectoryID != normalized {
			return fmt.Errorf("selected trajectory identifiers must be normalized")
		}
		if _, exists := seen[trajectoryID]; exists {
			return fmt.Errorf("duplicate selected trajectory identifier")
		}
		seen[trajectoryID] = struct{}{}
	}
	if !sort.StringsAreSorted(result.SelectedTrajectoryIDs) {
		return fmt.Errorf("selected trajectory identifiers must be sorted")
	}
	return nil
}

func validateLimitations(items []Notice) error {
	previousKey := ""
	for index, item := range items {
		code := strings.TrimSpace(item.Code)
		message := strings.TrimSpace(item.Message)
		if code == "" || message == "" || code != item.Code || message != item.Message {
			return fmt.Errorf("pattern confidence limitation is invalid or not normalized")
		}
		key := code + "\x00" + message
		if index > 0 && key <= previousKey {
			return fmt.Errorf("pattern confidence limitations must be sorted and unique")
		}
		previousKey = key
	}
	return nil
}

func hasNotice(items []Notice, code string) bool {
	for _, item := range items {
		if item.Code == code {
			return true
		}
	}
	return false
}
