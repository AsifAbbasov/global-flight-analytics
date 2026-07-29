package projectionproduction

import (
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
)

func TestComposeRejectsMissingHistoricalCandidateRouteScope(
	t *testing.T,
) {
	fixture := newProductionFixture()
	fixture.config.DependencyFailurePolicy = DependencyFailureReturnError
	fixture.request.HistoricalCandidateRouteScope = projectionneighbors.RouteScope{}

	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = composer.Compose(fixture.request)
	if !errors.Is(err, ErrNeighborSelectionFailed) {
		t.Fatalf(
			"Compose() error = %v, want %v",
			err,
			ErrNeighborSelectionFailed,
		)
	}
	if fixture.selector.calls != 0 {
		t.Fatalf("selector calls = %d, want 0", fixture.selector.calls)
	}
}

func TestComposeRejectsHistoricalCandidateRouteMismatch(
	t *testing.T,
) {
	fixture := newProductionFixture()
	fixture.config.DependencyFailurePolicy = DependencyFailureReturnError
	fixture.request.HistoricalCandidateRouteScope =
		projectionneighbors.UniformRouteScope(
			"UBBB",
			"UGTB",
			"production-test-route-scope-v1",
		)

	composer, err := New(fixture.config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = composer.Compose(fixture.request)
	if !errors.Is(err, ErrNeighborSelectionFailed) {
		t.Fatalf(
			"Compose() error = %v, want %v",
			err,
			ErrNeighborSelectionFailed,
		)
	}
	if fixture.selector.calls != 0 {
		t.Fatalf("selector calls = %d, want 0", fixture.selector.calls)
	}
}
