package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (repository *TrajectoryRepository) ListTrajectoriesByEndTimeAndBounds(
	ctx context.Context,
	observedFrom time.Time,
	observedTo time.Time,
	bounds metricquery.Bounds,
	limit int,
) ([]trajectory.FlightTrajectory, error) {
	if err := requireRepositoryContext(ctx); err != nil {
		return nil, err
	}
	if err := bounds.Validate(); err != nil {
		return nil, err
	}

	rows, err := repository.trajectoryReadExecutor().Query(
		ctx,
		trajectoriesByEndTimeAndBoundsQuery,
		observedFrom.UTC(),
		observedTo.UTC(),
		bounds.MinLatitude,
		bounds.MaxLatitude,
		bounds.MinLongitude,
		bounds.MaxLongitude,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query analytical trajectories by end time and bounds: %w",
			err,
		)
	}
	defer rows.Close()

	return scanFlightTrajectoryRows(
		rows,
		"query analytical trajectories by end time and bounds",
	)
}
