package projectionproduction

import (
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (request Request) Clone() Request {
	cloned := request
	cloned.CurrentTrajectory = request.CurrentTrajectory.Clone()
	cloned.HistoricalCandidates = cloneTrajectories(request.HistoricalCandidates)
	cloned.HistoricalCandidateRouteScope = request.HistoricalCandidateRouteScope.Clone()
	cloned.Route = request.Route.Clone()
	if request.RouteHistory != nil {
		history := request.RouteHistory.Clone()
		cloned.RouteHistory = &history
	}
	return cloned
}

func cloneTrajectories(items []trajectory.FlightTrajectory) []trajectory.FlightTrajectory {
	result := make([]trajectory.FlightTrajectory, 0, len(items))
	for _, item := range items {
		result = append(result, item.Clone())
	}
	return result
}

func candidateIDs(items []trajectory.FlightTrajectory) map[string]struct{} {
	result := make(map[string]struct{}, len(items))
	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		if id != "" {
			result[id] = struct{}{}
		}
	}
	return result
}
