package projectionread

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestSnapshotCloneDoesNotShareMutableSlices(
	t *testing.T,
) {
	asOfTime := projectionReadTestAsOfTime()
	current := projectionReadTrajectory(
		"73aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime,
	)
	candidate := projectionReadTrajectory(
		"83aa02ab-7061-4e9e-a238-d32710371ee3",
		asOfTime.Add(-24*time.Hour),
	)
	route := projectionReadCompleteRoute(current, asOfTime)
	history := projectionReadHistory(asOfTime)
	snapshot := Snapshot{
		CurrentTrajectory: current,
		Route:             routePointer(route),
		HistoricalCandidates: []trajectory.FlightTrajectory{
			candidate,
		},
		RouteHistory: historyPointer(history),
	}

	cloned := snapshot.Clone()
	cloned.CurrentTrajectory.Points[0].Latitude = 1
	cloned.HistoricalCandidates[0].Points[0].Latitude = 2
	cloned.Route.Status = "mutated"
	cloned.RouteHistory.ObservationCount = 999

	if snapshot.CurrentTrajectory.Points[0].Latitude == 1 ||
		snapshot.HistoricalCandidates[0].Points[0].Latitude == 2 ||
		snapshot.Route.Status == "mutated" ||
		snapshot.RouteHistory.ObservationCount == 999 {
		t.Fatal("snapshot clone shares mutable state")
	}
}
