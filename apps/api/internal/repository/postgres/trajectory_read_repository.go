package postgres

import (
	"context"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (
	repository *TrajectoryRepository,
) GetLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	return repository.withTrajectoryReadSnapshot(
		ctx,
		func(
			snapshotRepository *TrajectoryRepository,
		) (trajectory.FlightTrajectory, error) {
			return snapshotRepository.getLatestTrajectoryByICAO24(
				ctx,
				icao24,
			)
		},
	)
}
func (
	repository *TrajectoryRepository,
) GetTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	return repository.withTrajectoryReadSnapshot(
		ctx,
		func(
			snapshotRepository *TrajectoryRepository,
		) (trajectory.FlightTrajectory, error) {
			return snapshotRepository.getTrajectoryByID(
				ctx,
				trajectoryID,
			)
		},
	)
}
