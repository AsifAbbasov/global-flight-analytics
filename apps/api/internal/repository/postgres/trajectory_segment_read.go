package postgres

import (
	"context"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (
	repository *TrajectoryRepository,
) ListTrajectorySegments(
	ctx context.Context,
	trajectoryID string,
) ([]trajectory.TrajectorySegment, error) {
	if err := requireRepositoryContext(ctx); err != nil {
		return nil, err
	}

	rows, err := repository.trajectoryReadExecutor().Query(
		ctx,
		trajectorySegmentsByTrajectoryIDQuery,
		strings.TrimSpace(trajectoryID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTrajectorySegmentRows(rows)
}
