package interactiongraph

const (
	MinimumCompleteNodeCount = 2
	MinimumCompleteEdgeCount = 1

	MediumConfidenceMinimumScore = 0.50
	HighConfidenceMinimumScore   = 0.80
)

func statusForCounts(nodeCount int, edgeCount int) ResultStatus {
	switch {
	case nodeCount == 0:
		return ResultStatusUnavailable
	case nodeCount < MinimumCompleteNodeCount ||
		edgeCount < MinimumCompleteEdgeCount:
		return ResultStatusLimited
	default:
		return ResultStatusComplete
	}
}

func confidenceLevelForScore(score float64) ConfidenceLevel {
	switch {
	case score <= 0:
		return ConfidenceLevelNone
	case score < MediumConfidenceMinimumScore:
		return ConfidenceLevelLow
	case score < HighConfidenceMinimumScore:
		return ConfidenceLevelMedium
	default:
		return ConfidenceLevelHigh
	}
}
