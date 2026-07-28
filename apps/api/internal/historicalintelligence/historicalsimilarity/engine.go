package historicalsimilarity

import (
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (engine *Engine) Compare(
	reference trajectory.FlightTrajectory,
	candidate trajectory.FlightTrajectory,
) (Result, error) {
	if engine == nil {
		return Result{}, ErrEngineRequired
	}

	referenceID := strings.TrimSpace(reference.ID)
	candidateID := strings.TrimSpace(candidate.ID)
	if referenceID != "" &&
		referenceID == candidateID {
		return Result{}, ErrSameTrajectory
	}

	left, err := engine.prepare(reference)
	if err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %w",
				ErrReferenceNotComparable,
				err,
			)
	}
	right, err := engine.prepare(candidate)
	if err != nil {
		return Result{},
			fmt.Errorf(
				"%w: %w",
				ErrCandidateNotComparable,
				err,
			)
	}

	result := buildResult(
		left,
		right,
		engine.config,
	)
	if err := result.Validate(); err != nil {
		return Result{}, err
	}
	return result.Clone(), nil
}
