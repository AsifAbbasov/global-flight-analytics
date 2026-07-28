package projectionbaseline

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

type invalidBaselineHorizonPlanner struct{}

func (invalidBaselineHorizonPlanner) Build(
	projectionhorizon.Request,
) (projectionhorizon.Plan, error) {
	return projectionhorizon.Plan{}, nil
}

func TestProjectRejectsInvalidHorizonPlannerResult(t *testing.T) {
	config := validBaselineConfig()
	config.HorizonPlanner = invalidBaselineHorizonPlanner{}
	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = baseline.Project(baselineTestRequest())
	if !errors.Is(err, ErrHorizonPlanInvalid) {
		t.Fatalf("Project() error = %v, want %v", err, ErrHorizonPlanInvalid)
	}
}
