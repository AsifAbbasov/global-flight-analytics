package projectionneighbors

import "fmt"

type candidateEvaluation struct {
	qualified []Neighbor
	rejected  []Rejection
}

type candidateAssessment struct {
	neighbor  Neighbor
	rejection Rejection
	qualified bool
}

func (selector *Selector) evaluateCandidatePool(
	context selectionContext,
) (candidateEvaluation, error) {
	evaluation := candidateEvaluation{
		qualified: make(
			[]Neighbor,
			0,
			selector.config.SelectionLimit,
		),
		rejected: append(
			[]Rejection(nil),
			context.candidatePool.Rejections...,
		),
	}

	for _, prepared := range context.candidatePool.Candidates {
		assessment, err := selector.evaluateCandidate(
			context,
			prepared,
		)
		if err != nil {
			return candidateEvaluation{}, err
		}
		if assessment.qualified {
			evaluation.qualified = append(
				evaluation.qualified,
				assessment.neighbor,
			)
			continue
		}
		evaluation.rejected = append(
			evaluation.rejected,
			assessment.rejection,
		)
	}

	return evaluation, nil
}

func (selector *Selector) evaluateCandidate(
	context selectionContext,
	prepared preparedCandidate,
) (candidateAssessment, error) {
	candidateID := prepared.ID
	candidate := prepared.Trajectory

	anchorSearch := findAnchor(
		context.latestCurrentPoint,
		candidate.Points,
		selector.config.MinimumCurrentPointCount,
		context.requiredContinuationDuration,
		selector.config.effectiveMaximumContinuationGap(),
	)
	if !anchorSearch.Found() {
		code := RejectionContinuationUnavailable
		message := "Historical candidate does not provide enough observed continuation after a comparable prefix."
		if anchorSearch.Failure == anchorSearchFailureDiscontinuous {
			code = RejectionContinuationDiscontinuous
			message = "Historical candidate continuation crosses an observation gap larger than the configured maximum."
		}
		return rejectedCandidateAssessment(
			candidateID,
			code,
			message,
		), nil
	}

	anchor := anchorSearch.Evidence
	if anchor.DistanceKM > selector.config.MaximumAnchorDistanceKM {
		return rejectedCandidateAssessment(
			candidateID,
			RejectionAnchorTooDistant,
			"Historical candidate anchor exceeds the configured maximum distance from the current endpoint.",
		), nil
	}

	prefix := candidatePrefix(candidate, anchor.PointIndex)
	similarityResult, err := selector.config.SimilarityEngine.Compare(
		context.current,
		prefix,
	)
	if err != nil {
		if candidateSimilarityFailure(err) {
			return rejectedCandidateAssessment(
				candidateID,
				RejectionSimilarityUnavailable,
				"Historical candidate prefix could not be compared with the current trajectory.",
			), nil
		}
		return candidateAssessment{}, fmt.Errorf(
			"%w: candidate=%s: %v",
			ErrSimilarityEngineFailed,
			candidateID,
			err,
		)
	}

	similarity, err := similarityEvidenceFromResult(
		similarityResult,
		context.current,
		prefix,
	)
	if err != nil {
		return candidateAssessment{}, fmt.Errorf(
			"%w: candidate=%s: %v",
			ErrSimilarityEvidenceInvalid,
			candidateID,
			err,
		)
	}
	if similarity.Score < selector.config.MinimumSimilarityScore {
		return rejectedCandidateAssessment(
			candidateID,
			RejectionSimilarityBelowMinimum,
			"Historical candidate similarity is below the configured minimum.",
		), nil
	}

	return candidateAssessment{
		qualified: true,
		neighbor: Neighbor{
			TrajectoryID: candidateID,

			SimilarityScore:            similarity.Score,
			SimilarityLevel:            similarity.Level,
			SimilarityInputFingerprint: similarity.InputFingerprint,

			AnchorPointIndex: anchor.PointIndex,
			AnchorObservedAt: anchor.ObservedAt,
			AnchorDistanceKM: anchor.DistanceKM,

			CandidateStartTime: prepared.StartTime,
			CandidateEndTime:   prepared.EndTime,
			CandidateAge:       prepared.Age,

			PrefixPointCount:       anchor.PointIndex + 1,
			ContinuationPointCount: anchor.ContinuationPointCount,
			ContinuationEndTime:    anchor.ContinuationEndTime,
		},
	}, nil
}

func rejectedCandidateAssessment(
	trajectoryID string,
	code RejectionCode,
	message string,
) candidateAssessment {
	return candidateAssessment{
		rejection: rejection(
			trajectoryID,
			code,
			message,
		),
	}
}
