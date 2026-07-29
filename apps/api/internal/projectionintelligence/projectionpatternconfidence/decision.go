package projectionpatternconfidence

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

type patternDecision struct {
	status      Status
	usable      bool
	limitations []Notice
}

func decidePattern(
	selection projectionneighbors.Result,
	evidence patternEvidence,
	score float64,
	config Config,
) patternDecision {
	neighborCount := len(evidence.neighbors)
	limitations := append([]Notice(nil), evidence.limitations...)

	supportSufficient := neighborCount >= config.MinimumNeighborCount
	similarityFloorSufficient := neighborCount > 0 &&
		evidence.minimumSimilarityScore >= config.MinimumSimilarityScore
	dispersionAcceptable := neighborCount > 0 &&
		evidence.similarityStandardDeviation <=
			config.MaximumSimilarityStandardDeviation
	agreementAvailable := evidence.continuation.known
	agreementAcceptable := agreementAvailable &&
		evidence.continuation.maximumDivergenceMPS <=
			config.MaximumContinuationDivergenceMPS
	scoreSufficient := score >= config.MinimumUsableScore

	if !supportSufficient {
		limitations = append(limitations, Notice{
			Code: "insufficient_historical_neighbor_support",
			Message: fmt.Sprintf(
				"Pattern requires at least %d neighbors, but %d were selected.",
				config.MinimumNeighborCount,
				neighborCount,
			),
		})
	}
	if neighborCount > 0 && !similarityFloorSufficient {
		limitations = append(limitations, Notice{
			Code: "pattern_similarity_floor_below_minimum",
			Message: fmt.Sprintf(
				"Minimum selected-neighbor similarity %.6f is below the configured minimum %.6f.",
				evidence.minimumSimilarityScore,
				config.MinimumSimilarityScore,
			),
		})
	}
	if neighborCount > 0 && !dispersionAcceptable {
		limitations = append(limitations, Notice{
			Code: "pattern_similarity_dispersion_above_maximum",
			Message: fmt.Sprintf(
				"Selected-neighbor similarity standard deviation %.6f exceeds the configured maximum %.6f.",
				evidence.similarityStandardDeviation,
				config.MaximumSimilarityStandardDeviation,
			),
		})
	}
	if !agreementAvailable {
		limitations = append(limitations, Notice{
			Code:    "pattern_continuation_agreement_unavailable",
			Message: "Future continuation agreement was not evaluated, so the historical pattern cannot authorize a projection.",
		})
	} else if !agreementAcceptable {
		limitations = append(limitations, Notice{
			Code: "pattern_continuation_divergence_above_maximum",
			Message: fmt.Sprintf(
				"Maximum continuation divergence %.6f m/s exceeds the configured maximum %.6f m/s.",
				evidence.continuation.maximumDivergenceMPS,
				config.MaximumContinuationDivergenceMPS,
			),
		})
	}
	if !scoreSufficient {
		limitations = append(limitations, Notice{
			Code: "pattern_confidence_below_minimum",
			Message: fmt.Sprintf(
				"Pattern confidence score %.6f is below the configured minimum %.6f.",
				score,
				config.MinimumUsableScore,
			),
		})
	}

	usable := supportSufficient &&
		similarityFloorSufficient &&
		dispersionAcceptable &&
		agreementAvailable &&
		agreementAcceptable &&
		scoreSufficient
	if !usable {
		return patternDecision{
			status:      StatusUnavailable,
			usable:      false,
			limitations: normalizeNotices(limitations),
		}
	}

	if neighborCount >= config.TargetNeighborCount &&
		selection.Status == projectionneighbors.StatusComplete {
		return patternDecision{
			status:      StatusComplete,
			usable:      true,
			limitations: normalizeNotices(limitations),
		}
	}

	limitations = append(limitations, Notice{
		Code:    "pattern_support_partial",
		Message: "Historical pattern is usable but does not satisfy complete target support.",
	})
	return patternDecision{
		status:      StatusLimited,
		usable:      true,
		limitations: normalizeNotices(limitations),
	}
}
