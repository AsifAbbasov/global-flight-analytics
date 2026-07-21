package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (repository *TrajectoryRepository) ListTrajectoriesByEndTime(
	ctx context.Context,
	observedFrom time.Time,
	observedTo time.Time,
	limit int,
) ([]trajectory.FlightTrajectory, error) {
	if err := requireRepositoryContext(ctx); err != nil {
		return nil, err
	}

	rows, err := repository.trajectoryReadExecutor().Query(
		ctx,
		trajectoriesByEndTimeQuery,
		observedFrom.UTC(),
		observedTo.UTC(),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query analytical trajectories by end time: %w",
			err,
		)
	}
	defer rows.Close()

	return scanFlightTrajectoryRows(
		rows,
		"query analytical trajectories by end time",
	)
}

func (repository *TrajectoryRepository) ListTrajectoriesByIDs(
	ctx context.Context,
	trajectoryIDs []string,
) ([]trajectory.FlightTrajectory, error) {
	if err := requireRepositoryContext(ctx); err != nil {
		return nil, err
	}

	rows, err := repository.trajectoryReadExecutor().Query(
		ctx,
		trajectoriesByIDsQuery,
		trajectoryIDs,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query analytical trajectories by ids: %w",
			err,
		)
	}
	defer rows.Close()

	return scanFlightTrajectoryRows(
		rows,
		"query analytical trajectories by ids",
	)
}
