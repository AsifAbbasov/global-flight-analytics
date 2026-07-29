package projectionpatternconfidence

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"

func buildComponents(config Config, evidence patternEvidence) []Component {
	return []Component{
		{
			Name:   ComponentSimilarity,
			Score:  evidence.similarityScore,
			Weight: config.SimilarityWeight,
		},
		{
			Name:   ComponentSupport,
			Score:  evidence.supportScore,
			Weight: config.SupportWeight,
		},
		{
			Name:   ComponentFreshness,
			Score:  evidence.freshnessScore,
			Weight: config.FreshnessWeight,
		},
		{
			Name:   ComponentAnchorProximity,
			Score:  evidence.anchorProximityScore,
			Weight: config.AnchorProximityWeight,
		},
	}
}

func weightedComponentScore(components []Component) float64 {
	score := 0.0
	for _, component := range components {
		score += component.Score * component.Weight
	}
	return clampUnit(score)
}

func confidenceLevelForScore(
	score float64,
	mediumMinimum float64,
	highMinimum float64,
) projectioncontract.ConfidenceLevel {
	switch {
	case score >= highMinimum:
		return projectioncontract.ConfidenceLevelHigh
	case score >= mediumMinimum:
		return projectioncontract.ConfidenceLevelMedium
	case score > 0:
		return projectioncontract.ConfidenceLevelLow
	default:
		return projectioncontract.ConfidenceLevelNone
	}
}
