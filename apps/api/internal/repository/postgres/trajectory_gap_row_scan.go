package postgres

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

func scanCoverageGap(
	scanner postgresRowScanner,
) (trajectory.CoverageGap, error) {
	var item trajectory.CoverageGap
	var reason string
	if err := scanner.Scan(
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
		return trajectory.CoverageGap{}, err
	}
	item.Reason = trajectory.CoverageGapReason(reason)
	return item, nil
}

func scanCoverageGapRows(
	rows pgx.Rows,
) ([]trajectory.CoverageGap, error) {
	items := make([]trajectory.CoverageGap, 0)
	for rows.Next() {
		item, err := scanCoverageGap(rows)
		if err != nil {
			return nil, fmt.Errorf("scan trajectory coverage gap: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trajectory coverage gaps: %w", err)
	}
	return items, nil
}
