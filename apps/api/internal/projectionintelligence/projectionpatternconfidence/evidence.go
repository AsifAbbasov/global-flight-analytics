package projectionpatternconfidence

import (
	"math"
	"sort"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

const maximumUnitIntervalStandardDeviation = 0.5

type neighborPatternEvidence struct {
	trajectoryID               string
	similarityScore            float64
	similarityInputFingerprint string
	anchorDistanceKM           float64
}

type patternEvidence struct {
	neighbors []neighborPatternEvidence

	trajectoryIDs []string
	limitations   []Notice

	meanSimilarityScore         float64
	minimumSimilarityScore      float64
	similarityStandardDeviation float64
	meanAnchorDistanceKM        float64

	similarityStrengthScore    float64
	supportScore               float64
	similarityConsistencyScore float64
	anchorProximityScore       float64
}

func extractPatternEvidence(
	selection projectionneighbors.Result,
	config Config,
) patternEvidence {
	evidence := patternEvidence{
		neighbors:     make([]neighborPatternEvidence, 0, len(selection.Neighbors)),
		trajectoryIDs: make([]string, 0, len(selection.Neighbors)),
		limitations:   make([]Notice, 0, len(selection.Limitations)+4),
	}

	for _, limitation := range selection.Limitations {
		evidence.limitations = append(
			evidence.limitations,
			Notice{
				Code:    "neighbor_selection_" + strings.TrimSpace(limitation.Code),
				Message: strings.TrimSpace(limitation.Message),
			},
		)
	}

	for _, neighbor := range selection.Neighbors {
		item := neighborPatternEvidence{
			trajectoryID:               strings.TrimSpace(neighbor.TrajectoryID),
			similarityScore:            neighbor.SimilarityScore,
			similarityInputFingerprint: neighbor.SimilarityInputFingerprint,
			anchorDistanceKM:           neighbor.AnchorDistanceKM,
		}
		evidence.neighbors = append(evidence.neighbors, item)
		evidence.trajectoryIDs = append(evidence.trajectoryIDs, item.trajectoryID)
		evidence.meanSimilarityScore += item.similarityScore
		evidence.meanAnchorDistanceKM += item.anchorDistanceKM
	}

	sort.SliceStable(evidence.neighbors, func(left int, right int) bool {
		return evidence.neighbors[left].trajectoryID < evidence.neighbors[right].trajectoryID
	})
	sort.Strings(evidence.trajectoryIDs)

	neighborCount := len(evidence.neighbors)
	if neighborCount > 0 {
		divisor := float64(neighborCount)
		evidence.meanSimilarityScore /= divisor
		evidence.meanAnchorDistanceKM /= divisor
		evidence.minimumSimilarityScore = evidence.neighbors[0].similarityScore

		variance := 0.0
		for _, neighbor := range evidence.neighbors {
			if neighbor.similarityScore < evidence.minimumSimilarityScore {
				evidence.minimumSimilarityScore = neighbor.similarityScore
			}
			delta := neighbor.similarityScore - evidence.meanSimilarityScore
			variance += delta * delta
		}
		evidence.similarityStandardDeviation = math.Sqrt(variance / divisor)
	}

	evidence.similarityStrengthScore = clampUnit(evidence.meanSimilarityScore)
	evidence.supportScore = clampUnit(
		float64(neighborCount) / float64(config.TargetNeighborCount),
	)
	evidence.similarityConsistencyScore = clampUnit(
		1 - evidence.similarityStandardDeviation/maximumUnitIntervalStandardDeviation,
	)
	if neighborCount > 0 {
		evidence.anchorProximityScore = clampUnit(
			1 - evidence.meanAnchorDistanceKM/config.AnchorDistanceNormalizationKM,
		)
	}
	evidence.limitations = normalizeNotices(evidence.limitations)
	return evidence
}
