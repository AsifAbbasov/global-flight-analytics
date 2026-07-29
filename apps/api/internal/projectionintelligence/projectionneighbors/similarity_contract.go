package projectionneighbors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalsimilarity"
)

type similarityEvidence struct {
	Score            float64
	Level            historicalsimilarity.Level
	InputFingerprint string
}

func similarityEvidenceFromResult(
	result historicalsimilarity.Result,
	reference trajectory.FlightTrajectory,
	candidate trajectory.FlightTrajectory,
) (similarityEvidence, error) {
	if result.Version != historicalsimilarity.Version {
		return similarityEvidence{}, fmt.Errorf(
			"similarity version is invalid: %q",
			result.Version,
		)
	}
	if strings.TrimSpace(result.ReferenceTrajectoryID) !=
		strings.TrimSpace(reference.ID) ||
		strings.TrimSpace(result.CandidateTrajectoryID) !=
			strings.TrimSpace(candidate.ID) {
		return similarityEvidence{}, fmt.Errorf(
			"similarity trajectory identity does not match the compared inputs",
		)
	}
	if !unitInterval(result.Score) ||
		result.Level != historicalsimilarity.LevelForScore(result.Score) {
		return similarityEvidence{}, fmt.Errorf(
			"similarity score or level is invalid",
		)
	}
	if result.ReferencePointCount != len(reference.Points) ||
		result.CandidatePointCount != len(candidate.Points) ||
		result.SampleCount < 2 ||
		result.SampleCount > historicalsimilarity.MaximumSampleCount {
		return similarityEvidence{}, fmt.Errorf(
			"similarity point or sample counts are invalid",
		)
	}
	if !fingerprintPattern.MatchString(result.InputFingerprint) {
		return similarityEvidence{}, fmt.Errorf(
			"similarity input fingerprint is invalid",
		)
	}

	return similarityEvidence{
		Score:            result.Score,
		Level:            result.Level,
		InputFingerprint: result.InputFingerprint,
	}, nil
}

func candidateSimilarityFailure(err error) bool {
	return errors.Is(
		err,
		historicalsimilarity.ErrCandidateNotComparable,
	)
}
