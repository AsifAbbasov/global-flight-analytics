package projectionread

import "math"

const historicalCandidateScanMultiplier = 4

func historicalCandidateScanLimit(
	maximumCandidateCount int,
) int {
	if maximumCandidateCount < 1 {
		return 0
	}
	if maximumCandidateCount >
		math.MaxInt/historicalCandidateScanMultiplier {
		return math.MaxInt
	}
	return maximumCandidateCount *
		historicalCandidateScanMultiplier
}
