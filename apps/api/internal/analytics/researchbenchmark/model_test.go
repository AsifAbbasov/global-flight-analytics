package researchbenchmark

import "testing"

func TestDefaultPlansRemainBoundedAndOffline(t *testing.T) {
	plans := DefaultPlans()
	if len(plans) != 5 {
		t.Fatalf("plan count = %d, want 5", len(plans))
	}
	for _, plan := range plans {
		if err := Validate(plan); err != nil {
			t.Fatalf("validate %s: %v", plan.ID, err)
		}
		if plan.ProductionDependency {
			t.Fatalf("%s became a production dependency", plan.ID)
		}
	}
}
