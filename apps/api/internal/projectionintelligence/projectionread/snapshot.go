package projectionread

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionneighbors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type Snapshot struct {
	CurrentTrajectory trajectory.FlightTrajectory

	Route *routecontract.Result

	HistoricalCandidates          []trajectory.FlightTrajectory
	HistoricalCandidateRouteScope *projectionneighbors.RouteScope
	RouteHistory                  *projectionroutefrequency.HistorySummary
}

func (snapshot Snapshot) Clone() Snapshot {
	cloned := Snapshot{
		CurrentTrajectory: cloneTrajectory(
			snapshot.CurrentTrajectory,
		),
		HistoricalCandidates: make(
			[]trajectory.FlightTrajectory,
			0,
			len(snapshot.HistoricalCandidates),
		),
	}

	if snapshot.Route != nil {
		routeCopy := snapshot.Route.Clone()
		cloned.Route = &routeCopy
	}
	if snapshot.HistoricalCandidateRouteScope != nil {
		routeScopeCopy := snapshot.HistoricalCandidateRouteScope.Clone()
		cloned.HistoricalCandidateRouteScope = &routeScopeCopy
	}
	if snapshot.RouteHistory != nil {
		historyCopy := snapshot.RouteHistory.Clone()
		cloned.RouteHistory = &historyCopy
	}
	for _, candidate := range snapshot.HistoricalCandidates {
		cloned.HistoricalCandidates = append(
			cloned.HistoricalCandidates,
			cloneTrajectory(candidate),
		)
	}

	return cloned
}

func (snapshot Snapshot) CandidateRouteScope() projectionneighbors.RouteScope {
	if snapshot.HistoricalCandidateRouteScope == nil {
		return projectionneighbors.RouteScope{}
	}
	return snapshot.HistoricalCandidateRouteScope.Clone()
}

func cloneTrajectory(
	item trajectory.FlightTrajectory,
) trajectory.FlightTrajectory {
	cloned := item
	cloned.Points = append(
		[]trajectory.TrackPoint4D(nil),
		item.Points...,
	)
	cloned.Segments = append(
		[]trajectory.TrajectorySegment(nil),
		item.Segments...,
	)
	cloned.CoverageGaps = append(
		[]trajectory.CoverageGap(nil),
		item.CoverageGaps...,
	)
	return cloned
}
