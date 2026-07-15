package projectionread

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type Request struct {
	TrajectoryID      string
	AsOfTime          time.Time
	RequestedDuration time.Duration
}

type DataSource interface {
	LoadCurrentTrajectory(
		context.Context,
		string,
		time.Time,
	) (trajectory.FlightTrajectory, error)

	LoadRoute(
		context.Context,
		string,
		time.Time,
	) (routecontract.Result, error)

	LoadHistoricalCandidates(
		context.Context,
		trajectory.FlightTrajectory,
		routecontract.Result,
		time.Time,
	) ([]trajectory.FlightTrajectory, error)

	LoadRouteHistory(
		context.Context,
		routecontract.Result,
		time.Time,
	) (projectionroutefrequency.HistorySummary, error)
}

type Composer interface {
	Compose(
		projectionproduction.Request,
	) (projectionproduction.Result, error)
}
