package postgres

import (
	"context"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FeatureTrajectoryReader is the explicit production read path for feature
// materialization. It preserves ordinary trajectory-read payload size while
// hydrating point telemetry inside the same repeatable-read snapshot as parent
// metadata, segments, and coverage gaps.
type FeatureTrajectoryReader struct {
	repository *TrajectoryRepository
}

func NewFeatureTrajectoryReader(
	pool *pgxpool.Pool,
) *FeatureTrajectoryReader {
	return &FeatureTrajectoryReader{
		repository: NewTrajectoryRepository(pool),
	}
}

func (
	reader *FeatureTrajectoryReader,
) GetTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	return reader.repository.withTrajectoryReadSnapshot(
		ctx,
		func(
			snapshot *TrajectoryRepository,
		) (trajectory.FlightTrajectory, error) {
			item, err := snapshot.getTrajectoryByID(ctx, trajectoryID)
			if err != nil {
				return trajectory.FlightTrajectory{}, err
			}
			return hydrateFeatureTrajectoryPoints(ctx, snapshot, item)
		},
	)
}

func (
	reader *FeatureTrajectoryReader,
) GetLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
	return reader.repository.withTrajectoryReadSnapshot(
		ctx,
		func(
			snapshot *TrajectoryRepository,
		) (trajectory.FlightTrajectory, error) {
			item, err := snapshot.getLatestTrajectoryByICAO24(ctx, icao24)
			if err != nil {
				return trajectory.FlightTrajectory{}, err
			}
			return hydrateFeatureTrajectoryPoints(ctx, snapshot, item)
		},
	)
}

func hydrateFeatureTrajectoryPoints(
	ctx context.Context,
	repository *TrajectoryRepository,
	item trajectory.FlightTrajectory,
) (trajectory.FlightTrajectory, error) {
	points, err := repository.listTrajectoryPoints(ctx, item)
	if err != nil {
		return trajectory.FlightTrajectory{}, err
	}
	item.Points = points
	return item, nil
}
