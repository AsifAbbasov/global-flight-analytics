package projectionneighbors

import (
	"fmt"
	"sort"
)

type rankedNeighborSelection struct {
	neighbors                 []Neighbor
	qualifiedCandidateCount   int
	qualifiedSelectionLimited bool
}

func (selector *Selector) assembleSelectionResult(
	context selectionContext,
	evaluation candidateEvaluation,
) (Result, error) {
	ranked := rankAndLimitNeighbors(
		evaluation.qualified,
		selector.config.SelectionLimit,
	)
	candidateEvaluationTruncated :=
		context.candidatePool.CandidateEvaluationTruncated ||
			context.candidatePool.Truncated

	limitations := selectionLimitations(
		context,
		evaluation,
		ranked,
		candidateEvaluationTruncated,
	)
	status := selectionStatus(
		len(ranked.neighbors),
		selector.config.SelectionLimit,
		candidateEvaluationTruncated,
	)
	if status == StatusUnavailable {
		limitations = append(
			limitations,
			Notice{
				Code:    "historical_neighbor_unavailable",
				Message: "No historical trajectory satisfied the configured selection policy.",
			},
		)
	}

	result := Result{
		Version: Version,
		Status:  status,

		CurrentTrajectoryID:          context.currentID,
		AsOfTime:                     context.asOfTime,
		RequiredContinuationDuration: context.requiredContinuationDuration,

		InputCandidateCount: len(context.inputCandidates),
		CheckedCandidateCount: ranked.qualifiedCandidateCount +
			len(evaluation.rejected),
		QualifiedCandidateCount: ranked.qualifiedCandidateCount,
		RejectedCandidateCount:  len(evaluation.rejected),

		SelectionLimit: selector.config.SelectionLimit,

		CandidateEvaluationTruncated: candidateEvaluationTruncated,
		QualifiedSelectionLimited:    ranked.qualifiedSelectionLimited,
		Truncated:                    candidateEvaluationTruncated,

		Neighbors: append(
			[]Neighbor(nil),
			ranked.neighbors...,
		),
		Rejections: append(
			[]Rejection(nil),
			evaluation.rejected...,
		),
		Limitations: normalizeNotices(limitations),

		InputFingerprint: selectionFingerprint(
			context.current,
			context.inputCandidates,
			context.asOfTime,
			context.requiredContinuationDuration,
			selector.config,
			context.routeScope.scope,
		),
	}

	if err := result.Validate(); err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrSelectionResultInvalid,
			err,
		)
	}

	return result.Clone(), nil
}

func rankAndLimitNeighbors(
	input []Neighbor,
	selectionLimit int,
) rankedNeighborSelection {
	neighbors := append([]Neighbor(nil), input...)
	sort.SliceStable(
		neighbors,
		func(left int, right int) bool {
			if neighbors[left].SimilarityScore !=
				neighbors[right].SimilarityScore {
				return neighbors[left].SimilarityScore >
					neighbors[right].SimilarityScore
			}
			if neighbors[left].AnchorDistanceKM !=
				neighbors[right].AnchorDistanceKM {
				return neighbors[left].AnchorDistanceKM <
					neighbors[right].AnchorDistanceKM
			}
			return neighbors[left].TrajectoryID <
				neighbors[right].TrajectoryID
		},
	)

	qualifiedCandidateCount := len(neighbors)
	qualifiedSelectionLimited := qualifiedCandidateCount > selectionLimit
	if qualifiedSelectionLimited {
		neighbors = append(
			[]Neighbor(nil),
			neighbors[:selectionLimit]...,
		)
	}

	return rankedNeighborSelection{
		neighbors:                 neighbors,
		qualifiedCandidateCount:   qualifiedCandidateCount,
		qualifiedSelectionLimited: qualifiedSelectionLimited,
	}
}

func selectionLimitations(
	context selectionContext,
	evaluation candidateEvaluation,
	ranked rankedNeighborSelection,
	candidateEvaluationTruncated bool,
) []Notice {
	limitations := make([]Notice, 0, 5)
	if context.excludedCurrentPointCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "future_current_points_excluded",
				Message: fmt.Sprintf(
					"%d current-trajectory points after the as-of time were excluded.",
					context.excludedCurrentPointCount,
				),
			},
		)
	}
	if context.candidatePool.ExcludedFuturePointCount > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "future_candidate_points_excluded",
				Message: fmt.Sprintf(
					"%d historical-candidate points after the as-of time were excluded.",
					context.candidatePool.ExcludedFuturePointCount,
				),
			},
		)
	}
	if ranked.qualifiedSelectionLimited {
		limitations = append(
			limitations,
			Notice{
				Code: "qualified_neighbors_limited",
				Message: fmt.Sprintf(
					"%d qualified historical neighbors were reduced to the configured selection limit of %d.",
					ranked.qualifiedCandidateCount,
					len(ranked.neighbors),
				),
			},
		)
	}
	if candidateEvaluationTruncated {
		limitations = append(
			limitations,
			Notice{
				Code:    "candidate_evaluation_truncated",
				Message: "Historical candidate evaluation was truncated at the configured maximum candidate count.",
			},
		)
	}
	if len(evaluation.rejected) > 0 {
		limitations = append(
			limitations,
			Notice{
				Code: "historical_candidates_rejected",
				Message: fmt.Sprintf(
					"%d historical candidates were rejected by deterministic selection guards.",
					len(evaluation.rejected),
				),
			},
		)
	}
	return limitations
}

func selectionStatus(
	selectedNeighborCount int,
	selectionLimit int,
	candidateEvaluationTruncated bool,
) Status {
	switch {
	case selectedNeighborCount == 0:
		return StatusUnavailable
	case selectedNeighborCount == selectionLimit &&
		!candidateEvaluationTruncated:
		return StatusComplete
	default:
		return StatusPartial
	}
}
