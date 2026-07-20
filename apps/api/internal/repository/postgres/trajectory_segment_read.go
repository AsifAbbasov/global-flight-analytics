package postgres

import (
	"context"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"strings"
)

func (
	repository *TrajectoryRepository,
) ListTrajectorySegments(
	ctx context.Context,
	trajectoryID string,
) ([]trajectory.TrajectorySegment, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		SELECT
			id::text,
			trajectory_id::text,
			COALESCE(flight_id::text, ''),
			COALESCE(aircraft_id::text, ''),
			icao24,
			COALESCE(callsign, ''),
			sequence_number,
			status,
			quality_score::float8,
			start_time,
			end_time,
			duration_seconds::bigint,
			start_latitude::float8,
			start_longitude::float8,
			end_latitude::float8,
			end_longitude::float8,
			point_count,
			source_name,
			created_at
		FROM trajectory_segments
		WHERE trajectory_id = $1
		ORDER BY sequence_number ASC;
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
		[]trajectory.TrajectorySegment,
		0,
	)

	for rows.Next() {
		var item trajectory.TrajectorySegment
		var status string

		if err := rows.Scan(
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
			return nil,
				err
		}

		item.Status = trajectory.SegmentStatus(
			status,
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
