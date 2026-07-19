package researchbenchmark

import (
	"errors"
	"testing"
)

func TestPlanByIDReturnsIndependentClone(t *testing.T) {
	first, err := PlanByID(ProjectionFormulaEvaluationPlanID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := PlanByID(ProjectionFormulaEvaluationPlanID)
	if err != nil {
		t.Fatal(err)
	}

	first.Metrics[0] = "changed"
	if second.Metrics[0] == "changed" {
		t.Fatal("plan metrics share mutable storage")
	}
}

func TestPlanByIDRejectsUnknownPlan(t *testing.T) {
	_, err := PlanByID("unknown")
	if !errors.Is(err, ErrPlanInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrPlanInvalid)
	}
}
