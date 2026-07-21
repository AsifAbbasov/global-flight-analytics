package postgres

import (
	"context"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (
	repository *TrajectoryRepository,
) getLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
	normalizedICAO24 := normalizeICAO24Lookup(icao24)
	if normalizedICAO24 == "" {
		return trajectory.FlightTrajectory{}, trajectory.ErrNotFound
	}

	item, err := repository.queryFlightTrajectory(
		ctx,
		latestTrajectoryByICAO24Query,
		normalizedICAO24,
	)
	if err != nil {
		return trajectory.FlightTrajectory{}, err
	}
	if err := repository.loadTrajectoryChildren(ctx, &item); err != nil {
		return trajectory.FlightTrajectory{}, err
	}
	return item, nil
}

func (
	repository *TrajectoryRepository,
) getTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	trimmedTrajectoryID := strings.TrimSpace(trajectoryID)
	if trimmedTrajectoryID == "" {
		return trajectory.FlightTrajectory{}, trajectory.ErrNotFound
	}

	item, err := repository.queryFlightTrajectory(
		ctx,
		trajectoryByIDQuery,
		trimmedTrajectoryID,
	)
	if err != nil {
		return trajectory.FlightTrajectory{}, err
	}
	if err := repository.loadTrajectoryChildren(ctx, &item); err != nil {
		return trajectory.FlightTrajectory{}, err
	}
	return item, nil
}

func normalizeICAO24Lookup(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
