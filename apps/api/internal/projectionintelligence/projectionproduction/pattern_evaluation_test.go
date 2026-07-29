package projectionproduction

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionpatternconfidence"
)

type agreementAwareEvaluatorProbe struct {
	legacyCalls       int
	continuationCalls int
	candidateCount    int
}

func (probe *agreementAwareEvaluatorProbe) Evaluate(
	projectionneighbors.Result,
) (projectionpatternconfidence.Result, error) {
	probe.legacyCalls++
	return projectionpatternconfidence.Result{}, nil
}

func (probe *agreementAwareEvaluatorProbe) EvaluateWithContinuations(
	_ projectionneighbors.Result,
	candidates []trajectory.FlightTrajectory,
) (projectionpatternconfidence.Result, error) {
	probe.continuationCalls++
	probe.candidateCount = len(candidates)
	return projectionpatternconfidence.Result{}, nil
}

func TestEvaluatePatternConfidenceUsesContinuationEvidence(t *testing.T) {
	probe := &agreementAwareEvaluatorProbe{}
	_, err := evaluatePatternConfidence(
		probe,
		projectionneighbors.Result{},
		[]trajectory.FlightTrajectory{{ID: "candidate"}},
	)
	if err != nil {
		t.Fatalf("evaluatePatternConfidence() error = %v", err)
	}
	if probe.legacyCalls != 0 ||
		probe.continuationCalls != 1 ||
		probe.candidateCount != 1 {
		t.Fatalf("unexpected evaluator calls: %#v", probe)
	}
}
