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
	usable := neighborCount >= config.MinimumNeighborCount &&
		score >= config.MinimumUsableScore
	limitations := append([]Notice(nil), evidence.limitations...)

	if !usable {
		if neighborCount < config.MinimumNeighborCount {
			limitations = append(limitations, Notice{
				Code: "insufficient_historical_neighbor_support",
				Message: fmt.Sprintf(
					"Pattern requires at least %d neighbors, but %d were selected.",
					config.MinimumNeighborCount,
					neighborCount,
				),
			})
		}
		if score < config.MinimumUsableScore {
			limitations = append(limitations, Notice{
				Code: "pattern_confidence_below_minimum",
				Message: fmt.Sprintf(
					"Pattern confidence score %.6f is below the configured minimum %.6f.",
					score,
					config.MinimumUsableScore,
				),
			})
		}
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
