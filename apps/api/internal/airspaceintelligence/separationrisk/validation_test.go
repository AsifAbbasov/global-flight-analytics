package separationrisk

import "testing"

func TestDefaultPolicyIsValid(t *testing.T) {
	if err := DefaultPolicy().Validate(); err != nil {
		t.Fatalf("DefaultPolicy().Validate() error = %v", err)
	}
}

func TestPolicyRejectsInvalidRiskThresholdOrder(t *testing.T) {
	policy := DefaultPolicy()
	policy.HighRiskMinimumScore = policy.ElevatedRiskMinimumScore
	if err := policy.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestValidateRejectsOperationalScopeDrift(t *testing.T) {
	result, err := Evaluate(riskRequest(t), DefaultPolicy())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	result.ScopeGuard = ""
	report := Validate(result, DefaultPolicy())
	if report.Status != ValidationStatusInvalid {
		t.Fatalf("Validate() = %+v", report)
	}
}

func TestCloneIsDeep(t *testing.T) {
	result, err := Evaluate(riskRequest(t), DefaultPolicy())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	cloned := result.Clone()
	*cloned.Assessments[0].RiskScore = 0
	cloned.Limitations[0].Code = "changed"
	if *result.Assessments[0].RiskScore == 0 || result.Limitations[0].Code == "changed" {
		t.Fatal("Clone() did not deep-copy nested values")
	}
}
