package projectioncontinuation

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

func evaluatePatternConfidence(
	evaluator PatternConfidenceEvaluator,
	selection projectionneighbors.Result,
	candidates []trajectory.FlightTrajectory,
) (projectionpatternconfidence.Result, error) {
	return evaluator.EvaluateWithContinuations(
		selection,
		candidates,
	)
}
