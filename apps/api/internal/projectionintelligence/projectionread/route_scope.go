package projectionread

import (
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const routeScopedCandidateEvidenceVersion = "projection-read-route-scoped-candidates-v1"

func routeScopedHistoricalCandidateEvidence(
	route routecontract.Result,
	candidates []trajectory.FlightTrajectory,
	sourceName string,
) (projectionneighbors.RouteScope, error) {
	if route.Status != routecontract.RouteStatusComplete ||
		route.Origin == nil ||
		route.Destination == nil {
		return projectionneighbors.RouteScope{}, fmt.Errorf(
			"complete origin and destination route evidence is required",
		)
	}

	source := strings.TrimSpace(sourceName)
	if source == "" {
		return projectionneighbors.RouteScope{}, fmt.Errorf(
			"projection read source name is required",
		)
	}
	source += ":" + routeScopedCandidateEvidenceVersion

	scope := projectionneighbors.UniformRouteScope(
		route.Origin.Airport.ICAOCode,
		route.Destination.Airport.ICAOCode,
		source,
	)
	if err := scope.ValidateForCandidates(candidates); err != nil {
		return projectionneighbors.RouteScope{}, fmt.Errorf(
			"validate route-scoped historical candidate evidence: %w",
			err,
		)
	}
	return scope, nil
}

func routeScopePointer(
	value projectionneighbors.RouteScope,
) *projectionneighbors.RouteScope {
	cloned := value.Clone()
	return &cloned
}
