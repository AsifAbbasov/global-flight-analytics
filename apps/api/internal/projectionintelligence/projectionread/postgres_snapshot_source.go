package projectionread

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func (
	source *PostgresDataSource,
) LoadSnapshot(
	ctx context.Context,
	request SnapshotRequest,
) (Snapshot, error) {
	if source == nil ||
		source.snapshotExecutor == nil {
		return Snapshot{}, ErrServiceUnavailable
	}
	ctx = nonNilContext(ctx)
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}

	return source.snapshotExecutor.Execute(
		ctx,
		func(
			client postgresClient,
			repository trajectoryRepository,
		) (Snapshot, error) {
			session := &PostgresDataSource{
				client:               client,
				trajectoryRepository: repository,
				policy:               source.policy,
			}
			return session.loadSnapshotWithinSession(
				ctx,
				request,
			)
		},
	)
}

func (
	source *PostgresDataSource,
) loadSnapshotWithinSession(
	ctx context.Context,
	request SnapshotRequest,
) (Snapshot, error) {
	current, err := source.LoadCurrentTrajectory(
		ctx,
		request.TrajectoryID,
		request.AsOfTime,
	)
	if err != nil {
		return Snapshot{}, err
	}

	route, err := source.LoadRoute(
		ctx,
		request.TrajectoryID,
		request.AsOfTime,
	)
	if errors.Is(err, ErrRouteNotFound) {
		return Snapshot{
			CurrentTrajectory: current,
		}.Clone(), nil
	}
	if err != nil {
		return Snapshot{}, fmt.Errorf(
			"load Projection Intelligence route inside snapshot: %w",
			err,
		)
	}

	snapshot := Snapshot{
		CurrentTrajectory: current,
		Route:             routePointer(route),
		HistoricalCandidates: []trajectory.
			FlightTrajectory{},
	}
	if route.Status != routecontract.RouteStatusComplete {
		return snapshot.Clone(), nil
	}

	candidates, err := source.LoadHistoricalCandidates(
		ctx,
		current,
		route,
		request.AsOfTime,
	)
	if err != nil {
		return Snapshot{}, fmt.Errorf(
			"load Projection Intelligence historical candidates inside snapshot: %w",
			err,
		)
	}
	snapshot.HistoricalCandidates = candidates

	history, err := source.LoadRouteHistory(
		ctx,
		route,
		request.AsOfTime,
	)
	switch {
	case err == nil:
		snapshot.RouteHistory = historyPointer(history)
	case errors.Is(err, ErrRouteHistoryNotFound):
		snapshot.RouteHistory = nil
	default:
		return Snapshot{}, fmt.Errorf(
			"load Projection Intelligence route history inside snapshot: %w",
			err,
		)
	}

	return snapshot.Clone(), nil
}

func routePointer(
	value routecontract.Result,
) *routecontract.Result {
	cloned := value.Clone()
	return &cloned
}

func historyPointer(
	value projectionroutefrequency.HistorySummary,
) *projectionroutefrequency.HistorySummary {
	cloned := value.Clone()
	return &cloned
}
