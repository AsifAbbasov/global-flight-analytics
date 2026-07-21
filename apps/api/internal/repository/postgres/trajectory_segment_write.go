package postgres

import (
	"context"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

func (repository *TrajectoryRepository) insertTrajectorySegments(
	ctx context.Context,
	tx pgx.Tx,
	trajectoryID string,
	segments []trajectory.TrajectorySegment,
) ([]string, error) {
	const query = `
		INSERT INTO trajectory_segments (
			trajectory_id,
			flight_id,
			aircraft_id,
			icao24,
			callsign,
			sequence_number,
			status,
			quality_score,
			start_time,
			end_time,
			duration_seconds,
			start_latitude,
			start_longitude,
			end_latitude,
			end_longitude,
			point_count,
			source_name
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17
		)
		RETURNING id::text;
	`

	segmentIDs := make(
		[]string,
		0,
		len(segments),
	)

	for _, segment := range segments {
		var segmentID string

		err := tx.QueryRow(
			ctx,
			query,
			trajectoryID,
			nullableUUID(segment.FlightID),
			nullableUUID(segment.AircraftID),
			segment.ICAO24,
			nullableText(segment.Callsign),
			segment.SequenceNumber,
			string(segment.Status),
			segment.QualityScore,
			segment.StartTime,
			segment.EndTime,
			segment.DurationSeconds,
			segment.StartLatitude,
			segment.StartLongitude,
			segment.EndLatitude,
			segment.EndLongitude,
			segment.PointCount,
			requiredSourceNameValue(segment.SourceName),
		).Scan(
			&segmentID,
		)
		if err != nil {
			return nil, err
		}

		segmentIDs = append(
			segmentIDs,
			segmentID,
		)
	}

	return segmentIDs, nil
}
