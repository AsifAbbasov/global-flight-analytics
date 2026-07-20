package postgres

import (
	"context"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (
	repository *TrajectoryRepository,
) loadTrajectoryChildren(
	ctx context.Context,
	item *trajectory.FlightTrajectory,
) error {
	segments, err := repository.ListTrajectorySegments(
		ctx,
		item.ID,
	)
	if err != nil {
		return err
	}

	coverageGaps, err := repository.ListCoverageGaps(
		ctx,
		item.ID,
	)
	if err != nil {
		return err
	}

	item.Segments = segments
	item.CoverageGaps = coverageGaps

	return nil
}
