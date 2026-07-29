package projectionpatternconfidence

import "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"

func buildComponents(config Config, evidence patternEvidence) []Component {
	return []Component{
		{
			Name:   ComponentSimilarityStrength,
			Score:  evidence.similarityStrengthScore,
			Weight: config.SimilarityStrengthWeight,
		},
		{
			Name:   ComponentSupport,
			Score:  evidence.supportScore,
			Weight: config.SupportWeight,
		},
		{
			Name:   ComponentSimilarityConsistency,
			Score:  evidence.similarityConsistencyScore,
			Weight: config.SimilarityConsistencyWeight,
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

func confidenceLevelForDecision(
	score float64,
	status Status,
	mediumMinimum float64,
	highMinimum float64,
) projectioncontract.ConfidenceLevel {
	if status == StatusUnavailable {
		return projectioncontract.ConfidenceLevelNone
	}

	level := projectioncontract.ConfidenceLevelLow
	switch {
	case score >= highMinimum:
		level = projectioncontract.ConfidenceLevelHigh
	case score >= mediumMinimum:
		level = projectioncontract.ConfidenceLevelMedium
	}
	if status == StatusLimited && level == projectioncontract.ConfidenceLevelHigh {
		return projectioncontract.ConfidenceLevelMedium
	}
	return level
}
