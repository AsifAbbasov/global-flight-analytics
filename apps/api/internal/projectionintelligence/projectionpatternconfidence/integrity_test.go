package projectionpatternconfidence

import "testing"

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
		t.Fatal("Validate() accepted a component inconsistent with aggregate evidence")
	}
}

func TestResultValidateRejectsPolicyMutation(t *testing.T) {
	result := validEvaluatedResult(t)
	result.Policy.MinimumUsableScore = 0.99
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted a usable decision inconsistent with policy")
	}
}

func TestResultValidateRejectsComponentWeightMutation(t *testing.T) {
	result := validEvaluatedResult(t)
	result.Components[0].Weight -= 0.05
	result.Components[1].Weight += 0.05
	result.Score = weightedComponentScore(result.Components)
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted component weights inconsistent with policy")
	}
}

func TestResultValidateRejectsUnknownAgreementForUsableResult(t *testing.T) {
	result := validEvaluatedResult(t)
	result.ContinuationAgreementKnown = false
	result.ContinuationAgreementSampleCount = 0
	result.ContinuationAgreementPairCount = 0
	result.ContinuationComparisonCount = 0
	result.ContinuationHorizonSeconds = 0
	result.MeanContinuationSpreadM = 0
	result.MaximumContinuationSpreadM = 0
	result.MeanContinuationDivergenceMPS = 0
	result.MaximumContinuationDivergenceMPS = 0
	result.Components[4].Score = 0
	result.Score = weightedComponentScore(result.Components)
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted usable pattern without continuation agreement")
	}
}

func TestResultValidateRejectsInvalidPairCount(t *testing.T) {
	result := validEvaluatedResult(t)
	result.ContinuationAgreementPairCount++
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted invalid continuation pair count")
	}
}

func TestResultValidateRejectsDivergenceDecisionMismatch(t *testing.T) {
	result := validEvaluatedResult(t)
	result.MaximumContinuationDivergenceMPS =
		result.Policy.MaximumContinuationDivergenceMPS + 1
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted usable pattern above divergence maximum")
	}
}

func TestResultValidateRejectsInconsistentSpreadEvidence(t *testing.T) {
	result := validEvaluatedResult(t)
	result.MaximumContinuationSpreadM =
		result.MaximumContinuationDivergenceMPS*result.ContinuationHorizonSeconds + 1
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted inconsistent continuation spread evidence")
	}
}

func TestResultValidateRejectsMissingDecisionLimitation(t *testing.T) {
	evaluator := newConfidenceEvaluator(t)
	result, err := evaluator.Evaluate(confidenceSelection(3))
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	filtered := make([]Notice, 0, len(result.Limitations))
	for _, limitation := range result.Limitations {
		if limitation.Code != "pattern_continuation_agreement_unavailable" {
			filtered = append(filtered, limitation)
		}
	}
	result.Limitations = filtered
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() accepted a missing decision limitation")
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
	selection := confidenceSelection(3)
	result, err := newConfidenceEvaluator(t).EvaluateWithContinuations(
		selection,
		confidenceCandidates(selection),
	)
	if err != nil {
		t.Fatalf("EvaluateWithContinuations() error = %v", err)
	}
	return result
}
