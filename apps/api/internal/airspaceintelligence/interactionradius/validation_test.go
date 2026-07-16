package interactionradius

import "testing"

func TestDefaultPolicyIsValid(t *testing.T) {
	if err := DefaultPolicy().Validate(); err != nil {
		t.Fatalf("DefaultPolicy().Validate() error = %v", err)
	}
}

func TestPolicyValidationRejectsInvalidWeights(t *testing.T) {
	policy := DefaultPolicy()
	policy.Weights.VerticalEvidence = 0.50
	if err := policy.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want invalid weights")
	}
}

func TestValidateRejectsMissingScopeGuard(t *testing.T) {
	decision, err := Evaluate(validRequest(), DefaultPolicy())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	decision.ScopeGuard = ""
	report := Validate(decision)
	if report.Status != ValidationStatusInvalid {
		t.Fatalf("Status = %q, want %q", report.Status, ValidationStatusInvalid)
	}
}

func TestValidateRejectsPositiveBlockedRadius(t *testing.T) {
	request := validRequest()
	request.QualityScore = 0.10
	decision, err := Evaluate(request, DefaultPolicy())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	decision.HorizontalRadiusKilometers = 10
	report := Validate(decision)
	if report.Status != ValidationStatusInvalid {
		t.Fatalf("Status = %q, want %q", report.Status, ValidationStatusInvalid)
	}
}
