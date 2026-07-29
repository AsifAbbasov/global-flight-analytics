package projectionproduction

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

type continuationAwarePatternConfidenceEvaluator interface {
	EvaluateWithContinuations(
		projectionneighbors.Result,
		[]trajectory.FlightTrajectory,
	) (projectionpatternconfidence.Result, error)
}

func evaluatePatternConfidence(
	evaluator PatternConfidenceEvaluator,
	selection projectionneighbors.Result,
	candidates []trajectory.FlightTrajectory,
) (projectionpatternconfidence.Result, error) {
	continuationAware, ok := evaluator.(continuationAwarePatternConfidenceEvaluator)
	if ok {
		return continuationAware.EvaluateWithContinuations(selection, candidates)
	}
	return evaluator.Evaluate(selection)
}
