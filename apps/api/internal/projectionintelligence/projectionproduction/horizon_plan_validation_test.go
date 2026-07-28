package projectionproduction

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionhorizon"
)

type invalidProductionHorizonPlanner struct{}

func (invalidProductionHorizonPlanner) Build(
	projectionhorizon.Request,
) (projectionhorizon.Plan, error) {
	return projectionhorizon.Plan{}, nil
}

func TestComposeRejectsInvalidHorizonPlannerResult(t *testing.T) {
	fixture := newProductionFixture()
	fixture.config.HorizonPlanner = invalidProductionHorizonPlanner{}

	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = composer.Compose(fixture.request)
	if !errors.Is(err, ErrHorizonPlanInvalid) {
		t.Fatalf("Compose() error = %v, want %v", err, ErrHorizonPlanInvalid)
	}
}
