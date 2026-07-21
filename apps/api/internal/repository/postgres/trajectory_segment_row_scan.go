package postgres

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

func scanTrajectorySegment(
	scanner postgresRowScanner,
) (trajectory.TrajectorySegment, error) {
	var item trajectory.TrajectorySegment
	var status string
	if err := scanner.Scan(
		&item.ID,
		&item.TrajectoryID,
		&item.FlightID,
		&item.AircraftID,
		&item.ICAO24,
		&item.Callsign,
		&item.SequenceNumber,
		&status,
		&item.QualityScore,
		&item.StartTime,
		&item.EndTime,
		&item.DurationSeconds,
		&item.StartLatitude,
		&item.StartLongitude,
		&item.EndLatitude,
		&item.EndLongitude,
		&item.PointCount,
		&item.SourceName,
		&item.CreatedAt,
	); err != nil {
		return trajectory.TrajectorySegment{}, err
	}
	item.Status = trajectory.SegmentStatus(status)
	return item, nil
}

func scanTrajectorySegmentRows(
	rows pgx.Rows,
) ([]trajectory.TrajectorySegment, error) {
	items := make([]trajectory.TrajectorySegment, 0)
	for rows.Next() {
		item, err := scanTrajectorySegment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan trajectory segment: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trajectory segments: %w", err)
	}
	return items, nil
}
