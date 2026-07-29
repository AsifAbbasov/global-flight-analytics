package projectionpatternconfidence

import (
	"sort"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

type neighborPatternEvidence struct {
	trajectoryID               string
	similarityScore            float64
	similarityInputFingerprint string
	candidateAge               time.Duration
	anchorDistanceKM           float64
}

type patternEvidence struct {
	neighbors []neighborPatternEvidence

	trajectoryIDs []string
	limitations   []Notice

	meanSimilarityScore     float64
	meanCandidateAgeSeconds float64
	meanAnchorDistanceKM    float64

	similarityScore      float64
	supportScore         float64
	freshnessScore       float64
	anchorProximityScore float64
}

func extractPatternEvidence(
	selection projectionneighbors.Result,
	config Config,
) patternEvidence {
	evidence := patternEvidence{
		neighbors:     make([]neighborPatternEvidence, 0, len(selection.Neighbors)),
		trajectoryIDs: make([]string, 0, len(selection.Neighbors)),
		limitations:   make([]Notice, 0, len(selection.Limitations)+3),
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
			candidateAge:               neighbor.CandidateAge,
			anchorDistanceKM:           neighbor.AnchorDistanceKM,
		}
		evidence.neighbors = append(evidence.neighbors, item)
		evidence.trajectoryIDs = append(evidence.trajectoryIDs, item.trajectoryID)
		evidence.meanSimilarityScore += item.similarityScore
		evidence.meanCandidateAgeSeconds += item.candidateAge.Seconds()
		evidence.meanAnchorDistanceKM += item.anchorDistanceKM
		evidence.freshnessScore += clampUnit(
			1 - item.candidateAge.Seconds()/config.MaximumCandidateAge.Seconds(),
		)
	}

	sort.SliceStable(evidence.neighbors, func(left int, right int) bool {
		return evidence.neighbors[left].trajectoryID < evidence.neighbors[right].trajectoryID
	})
	sort.Strings(evidence.trajectoryIDs)

	neighborCount := len(evidence.neighbors)
	if neighborCount > 0 {
		divisor := float64(neighborCount)
		evidence.meanSimilarityScore /= divisor
		evidence.meanCandidateAgeSeconds /= divisor
		evidence.meanAnchorDistanceKM /= divisor
		evidence.freshnessScore /= divisor
	}

	evidence.similarityScore = clampUnit(evidence.meanSimilarityScore)
	evidence.supportScore = clampUnit(
		float64(neighborCount) / float64(config.TargetNeighborCount),
	)
	if neighborCount > 0 {
		evidence.anchorProximityScore = clampUnit(
			1 - evidence.meanAnchorDistanceKM/config.MaximumMeanAnchorDistanceKM,
		)
	}
	evidence.limitations = normalizeNotices(evidence.limitations)
	return evidence
}
