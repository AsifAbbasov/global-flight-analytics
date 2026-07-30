package projectionfreshness

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

type freshnessDecision struct {
	decision    Decision
	usable      bool
	limitations []Notice
}

func evaluateFreshnessPolicy(
	metrics freshnessMetrics,
	score float64,
	selectionStatus projectionneighbors.Status,
	patternStatus projectionpatternconfidence.Status,
	patternUsable bool,
	config Config,
) freshnessDecision {
	limitations := make([]Notice, 0, 10)
	if len(metrics.ages) == 0 {
		limitations = append(limitations, Notice{
			Code:    "historical_neighbors_unavailable",
			Message: "Historical continuation is blocked because no selected historical neighbors are available.",
		})
	}
	if !patternUsable {
		limitations = append(limitations, Notice{
			Code:    "pattern_confidence_unusable",
			Message: "Historical continuation is blocked because Pattern Confidence did not authorize the selected evidence.",
		})
	}
	if len(metrics.ages) > 0 {
		if metrics.newestAge > config.MaximumNewestNeighborAge {
			limitations = append(limitations, Notice{
				Code: "newest_historical_neighbor_too_old",
				Message: fmt.Sprintf(
					"Newest selected historical neighbor age %s exceeds the configured maximum %s.",
					metrics.newestAge,
					config.MaximumNewestNeighborAge,
				),
			})
		}
		if metrics.meanAge > config.MaximumMeanNeighborAge {
			limitations = append(limitations, Notice{
				Code: "mean_historical_neighbor_age_too_old",
				Message: fmt.Sprintf(
					"Mean selected historical neighbor age %s exceeds the configured maximum %s.",
					metrics.meanAge,
					config.MaximumMeanNeighborAge,
				),
			})
		}
		if metrics.oldestAge > config.MaximumOldestNeighborAge {
			limitations = append(limitations, Notice{
				Code: "oldest_historical_neighbor_too_old",
				Message: fmt.Sprintf(
					"Oldest selected historical neighbor age %s exceeds the configured maximum %s.",
					metrics.oldestAge,
					config.MaximumOldestNeighborAge,
				),
			})
		}
	}
	if metrics.recentCount < config.MinimumRecentNeighborCount {
		limitations = append(limitations, Notice{
			Code: "recent_historical_neighbor_support_insufficient",
			Message: fmt.Sprintf(
				"Recent historical neighbor count %d is below the configured minimum %d.",
				metrics.recentCount,
				config.MinimumRecentNeighborCount,
			),
		})
	}
	if scoreBelowThreshold(score, config.MinimumUsableScore) {
		limitations = append(limitations, Notice{
			Code: "pattern_freshness_score_below_minimum",
			Message: fmt.Sprintf(
				"Pattern freshness score %.6f is below the configured usable minimum %.6f.",
				score,
				config.MinimumUsableScore,
			),
		})
	}
	limitations = normalizeNotices(limitations)
	if len(limitations) > 0 {
		return freshnessDecision{
			decision:    DecisionBlocked,
			usable:      false,
			limitations: limitations,
		}
	}

	if scoreBelowThreshold(score, config.CompleteScoreMinimum) {
		limitations = append(limitations, Notice{
			Code: "pattern_freshness_limited",
			Message: fmt.Sprintf(
				"Pattern freshness score %.6f is below the configured complete threshold %.6f.",
				score,
				config.CompleteScoreMinimum,
			),
		})
	}
	if patternStatus != projectionpatternconfidence.StatusComplete {
		limitations = append(limitations, Notice{
			Code:    "pattern_confidence_not_complete",
			Message: "Pattern confidence remains usable but is not complete, so freshness approval is limited.",
		})
	}
	if selectionStatus != projectionneighbors.StatusComplete {
		limitations = append(limitations, Notice{
			Code:    "neighbor_selection_not_complete",
			Message: "Historical neighbor selection remains usable but did not fill the configured selection target.",
		})
	}
	limitations = normalizeNotices(limitations)
	if len(limitations) > 0 {
		return freshnessDecision{
			decision:    DecisionLimited,
			usable:      true,
			limitations: limitations,
		}
	}
	return freshnessDecision{decision: DecisionAllowed, usable: true}
}

func scoreBelowThreshold(score float64, threshold float64) bool {
	return score+scoreComparisonTolerance < threshold
}
