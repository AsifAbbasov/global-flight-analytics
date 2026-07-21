package postgres

import (
	"context"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (
	repository *TrajectoryRepository,
) ListCoverageGaps(
	ctx context.Context,
	trajectoryID string,
) ([]trajectory.CoverageGap, error) {
	if err := requireRepositoryContext(ctx); err != nil {
		return nil, err
	}

	rows, err := repository.trajectoryReadExecutor().Query(
		ctx,
		coverageGapsByTrajectoryIDQuery,
		strings.TrimSpace(trajectoryID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCoverageGapRows(rows)
}
