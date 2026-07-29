package projectionproduction

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func validateHistoricalCandidateRouteScope(
	route routecontract.Result,
	scope projectionneighbors.RouteScope,
	candidates []trajectory.FlightTrajectory,
) error {
	if route.Status != routecontract.RouteStatusComplete ||
		route.Origin == nil ||
		route.Destination == nil {
		return fmt.Errorf(
			"complete origin and destination route evidence is required",
		)
	}
	if err := scope.ValidateForCandidates(candidates); err != nil {
		return fmt.Errorf("validate supplied route scope: %w", err)
	}

	expected := projectionneighbors.RouteKey{
		OriginICAO: route.Origin.Airport.ICAOCode,
		DestinationICAO: route.Destination.
			Airport.ICAOCode,
	}
	if !scope.Route.Equal(expected) {
		return fmt.Errorf(
			"historical candidate route scope does not match the current route",
		)
	}
	return nil
}

func routeScopeFromState(
	state *compositionState,
) projectionneighbors.RouteScope {
	if state == nil || state.routeScope == nil {
		return projectionneighbors.RouteScope{}
	}
	return state.routeScope.Clone()
}
