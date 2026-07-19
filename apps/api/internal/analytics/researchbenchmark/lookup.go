package researchbenchmark

import (
	"fmt"
	"strings"
)

const ProjectionFormulaEvaluationPlanID = "projection-formula-evaluation-v1"

func PlanByID(id string) (Plan, error) {
	normalized := strings.TrimSpace(id)
	for _, plan := range DefaultPlans() {
		if plan.ID != normalized {
			continue
		}

		cloned := plan
		cloned.Metrics = append([]string(nil), plan.Metrics...)
		cloned.Limitations = append(
			[]string(nil),
			plan.Limitations...,
		)
		return cloned, nil
	}

	return Plan{}, fmt.Errorf(
		"%w: unknown plan %q",
		ErrPlanInvalid,
		normalized,
	)
}
