package projectionpatternconfidence

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func TestResultValidateRejectsUnknownComponentName(t *testing.T) {
	result := validEvaluatedResult(t)
	result.Components[0].Name = ComponentName("unknown")
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted an unknown component name")
	}
}

func TestResultValidateRejectsComponentScoreMismatch(t *testing.T) {
	result := validEvaluatedResult(t)
	result.Components[0].Score -= 0.1
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted a score inconsistent with components")
	}
}

func TestResultValidateRejectsInvalidSimilarityDistribution(t *testing.T) {
	result := validEvaluatedResult(t)
	result.MinimumSimilarityScore = result.MeanSimilarityScore + 0.1
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted minimum similarity above mean similarity")
	}

	result = validEvaluatedResult(t)
	result.SimilarityStandardDeviation = 0.6
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted impossible similarity deviation")
	}
}

func TestResultValidateRejectsUsableNoneConfidence(t *testing.T) {
	result := validEvaluatedResult(t)
	for index := range result.Components {
		result.Components[index].Score = 0
	}
	result.Score = 0
	result.Level = projectioncontract.ConfidenceLevelNone
	result.Usable = true
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted usable result with none confidence")
	}
}

func TestResultValidateRejectsUnavailableHighConfidence(t *testing.T) {
	result := validEvaluatedResult(t)
	result.Status = StatusUnavailable
	result.Usable = false
	result.Level = projectioncontract.ConfidenceLevelHigh
	result.Limitations = []Notice{{Code: "blocked", Message: "Blocked."}}
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted unavailable high confidence")
	}
}

func TestResultValidateRejectsLimitedHighConfidence(t *testing.T) {
	result := validEvaluatedResult(t)
	result.Status = StatusLimited
	result.Level = projectioncontract.ConfidenceLevelHigh
	result.Limitations = []Notice{{Code: "limited", Message: "Limited."}}
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted limited high confidence")
	}
}

func TestResultValidateRejectsUnnormalizedTrajectoryID(t *testing.T) {
	result := validEvaluatedResult(t)
	result.SelectedTrajectoryIDs[0] = " " + result.SelectedTrajectoryIDs[0]
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted an unnormalized trajectory identifier")
	}
}

func TestResultValidateRejectsUnsortedLimitations(t *testing.T) {
	result := validEvaluatedResult(t)
	result.Limitations = []Notice{
		{Code: "z", Message: "Z."},
		{Code: "a", Message: "A."},
	}
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted unsorted limitations")
	}
}

func TestResultValidateRejectsDuplicateLimitations(t *testing.T) {
	result := validEvaluatedResult(t)
	result.Limitations = []Notice{
		{Code: "a", Message: "A."},
		{Code: "a", Message: "A."},
	}
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted duplicate limitations")
	}
}

func validEvaluatedResult(t *testing.T) Result {
	t.Helper()
	result, err := newConfidenceEvaluator(t).Evaluate(confidenceSelection(3))
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	return result
}
