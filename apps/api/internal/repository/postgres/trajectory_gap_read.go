package postgres

import (
	"context"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"strings"
)

func (
	repository *TrajectoryRepository,
) ListCoverageGaps(
	ctx context.Context,
	trajectoryID string,
) ([]trajectory.CoverageGap, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		SELECT
			id::text,
			trajectory_id::text,
			COALESCE(previous_segment_id::text, ''),
			COALESCE(next_segment_id::text, ''),
			icao24,
			gap_start_time,
			gap_end_time,
			duration_seconds::bigint,
			distance_km::float8,
			reason,
			COALESCE(filled_by, ''),
			created_at
		FROM coverage_gaps
		WHERE trajectory_id = $1
		ORDER BY gap_start_time ASC;
	`

	rows, err := repository.trajectoryReadExecutor().Query(
		ctx,
		query,
		strings.TrimSpace(
			trajectoryID,
		),
	)
	if err != nil {
		return nil,
			err
	}
	defer rows.Close()

	items := make(
		[]trajectory.CoverageGap,
		0,
	)

	for rows.Next() {
		var item trajectory.CoverageGap
		var reason string

		if err := rows.Scan(
			&item.ID,
			&item.TrajectoryID,
			&item.PreviousSegmentID,
			&item.NextSegmentID,
			&item.ICAO24,
			&item.StartTime,
			&item.EndTime,
			&item.DurationSeconds,
			&item.DistanceKm,
			&reason,
			&item.FilledBy,
			&item.CreatedAt,
		); err != nil {
			return nil,
				err
		}

		item.Reason = trajectory.CoverageGapReason(
			reason,
		)

		items = append(
			items,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil,
			err
	}

	return items,
		nil
}
