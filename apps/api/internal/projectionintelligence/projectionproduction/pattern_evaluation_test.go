package projectionproduction

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

type continuationEvaluatorProbe struct {
	calls          int
	candidateCount int
}

func (probe *continuationEvaluatorProbe) EvaluateWithContinuations(
	_ projectionneighbors.Result,
	candidates []trajectory.FlightTrajectory,
) (projectionpatternconfidence.Result, error) {
	probe.calls++
	probe.candidateCount = len(candidates)
	return projectionpatternconfidence.Result{}, nil
}

var _ PatternConfidenceEvaluator = (*continuationEvaluatorProbe)(nil)
var _ PatternConfidenceEvaluator = (*projectionpatternconfidence.Evaluator)(nil)

func TestEvaluatePatternConfidenceRequiresContinuationEvidence(t *testing.T) {
	probe := &continuationEvaluatorProbe{}
	_, err := evaluatePatternConfidence(
		probe,
		projectionneighbors.Result{},
		[]trajectory.FlightTrajectory{{ID: "candidate"}},
	)
	if err != nil {
		t.Fatalf("evaluatePatternConfidence() error = %v", err)
	}
	if probe.calls != 1 || probe.candidateCount != 1 {
		t.Fatalf("unexpected evaluator calls: %#v", probe)
	}
}
