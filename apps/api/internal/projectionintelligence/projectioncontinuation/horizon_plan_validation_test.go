package projectioncontinuation

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

type invalidContinuationHorizonPlanner struct{}

func (invalidContinuationHorizonPlanner) Build(
	projectionhorizon.Request,
) (projectionhorizon.Plan, error) {
	return projectionhorizon.Plan{}, nil
}

func TestProjectRejectsInvalidHorizonPlannerResult(t *testing.T) {
	config := validContinuationConfig(t)
	config.HorizonPlanner = invalidContinuationHorizonPlanner{}
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = baseline.Project(continuationTestRequest())
	if !errors.Is(err, ErrHorizonPlanInvalid) {
		t.Fatalf("Project() error = %v, want %v", err, ErrHorizonPlanInvalid)
	}
}
