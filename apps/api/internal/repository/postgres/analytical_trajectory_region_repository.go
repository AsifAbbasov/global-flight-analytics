package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func (repository *TrajectoryRepository) ListTrajectoriesByEndTimeAndBounds(
	ctx context.Context,
	observedFrom time.Time,
	observedTo time.Time,
	bounds metricquery.Bounds,
	limit int,
) ([]trajectory.FlightTrajectory, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := bounds.Validate(); err != nil {
		return nil, err
	}

	const query = `
		SELECT
			flight_trajectory.id::text,
			COALESCE(flight_trajectory.identity_key, ''),
			COALESCE(flight_trajectory.identity_basis, ''),
			COALESCE(flight_trajectory.split_reason, ''),
			COALESCE(flight_trajectory.flight_id::text, ''),
			COALESCE(flight_trajectory.aircraft_id::text, ''),
			flight_trajectory.icao24,
			COALESCE(flight_trajectory.callsign, ''),
			flight_trajectory.start_time,
			flight_trajectory.end_time,
			flight_trajectory.duration_seconds::bigint,
			flight_trajectory.segment_count,
			flight_trajectory.point_count,
			flight_trajectory.coverage_gap_count,
			flight_trajectory.quality_score::float8,
			flight_trajectory.source_name,
			flight_trajectory.created_at,
			flight_trajectory.updated_at
		FROM flight_trajectories flight_trajectory
		JOIN LATERAL (
			SELECT
				segment.end_latitude::float8 AS latitude,
				segment.end_longitude::float8 AS longitude
			FROM trajectory_segments segment
			WHERE segment.trajectory_id = flight_trajectory.id
				AND segment.status <> 'invalid'
			ORDER BY
				segment.sequence_number DESC,
				segment.end_time DESC,
				segment.created_at DESC
			LIMIT 1
		) latest_position ON TRUE
		WHERE flight_trajectory.end_time >= $1
			AND flight_trajectory.end_time <= $2
			AND latest_position.latitude BETWEEN $3 AND $4
			AND latest_position.longitude BETWEEN $5 AND $6
		ORDER BY
			flight_trajectory.end_time DESC,
			flight_trajectory.start_time DESC,
			flight_trajectory.created_at DESC
		LIMIT $7;
	`

	rows, err := repository.db.Query(
		ctx,
		query,
		observedFrom.UTC(),
		observedTo.UTC(),
		bounds.MinLatitude,
		bounds.MaxLatitude,
		bounds.MinLongitude,
		bounds.MaxLongitude,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query analytical trajectories by end time and bounds: %w",
			err,
		)
	}
	defer rows.Close()

	return scanAnalyticalTrajectories(rows)
}
