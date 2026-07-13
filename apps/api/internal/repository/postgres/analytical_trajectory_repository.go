package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

func (repository *TrajectoryRepository) ListTrajectoriesByEndTime(
	ctx context.Context,
	observedFrom time.Time,
	observedTo time.Time,
	limit int,
) ([]trajectory.FlightTrajectory, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		SELECT
			id::text,
			COALESCE(identity_key, ''),
			COALESCE(identity_basis, ''),
			COALESCE(split_reason, ''),
			COALESCE(flight_id::text, ''),
			COALESCE(aircraft_id::text, ''),
			icao24,
			COALESCE(callsign, ''),
			start_time,
			end_time,
			duration_seconds::bigint,
			segment_count,
			point_count,
			coverage_gap_count,
			quality_score::float8,
			source_name,
			created_at,
			updated_at
		FROM flight_trajectories
		WHERE end_time >= $1
			AND end_time <= $2
		ORDER BY end_time DESC, start_time DESC, created_at DESC
		LIMIT $3;
	`

	rows, err := repository.db.Query(
		ctx,
		query,
		observedFrom.UTC(),
		observedTo.UTC(),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query analytical trajectories by end time: %w",
			err,
		)
	}
	defer rows.Close()

	return scanAnalyticalTrajectories(rows)
}

func (repository *TrajectoryRepository) ListTrajectoriesByIDs(
	ctx context.Context,
	trajectoryIDs []string,
) ([]trajectory.FlightTrajectory, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		SELECT
			id::text,
			COALESCE(identity_key, ''),
			COALESCE(identity_basis, ''),
			COALESCE(split_reason, ''),
			COALESCE(flight_id::text, ''),
			COALESCE(aircraft_id::text, ''),
			icao24,
			COALESCE(callsign, ''),
			start_time,
			end_time,
			duration_seconds::bigint,
			segment_count,
			point_count,
			coverage_gap_count,
			quality_score::float8,
			source_name,
			created_at,
			updated_at
		FROM flight_trajectories
		WHERE id::text = ANY($1::text[])
		ORDER BY array_position($1::text[], id::text);
	`

	rows, err := repository.db.Query(ctx, query, trajectoryIDs)
	if err != nil {
		return nil, fmt.Errorf(
			"query analytical trajectories by ids: %w",
			err,
		)
	}
	defer rows.Close()

	return scanAnalyticalTrajectories(rows)
}

func scanAnalyticalTrajectories(
	rows pgx.Rows,
) ([]trajectory.FlightTrajectory, error) {
	items := make([]trajectory.FlightTrajectory, 0)

	for rows.Next() {
		var item trajectory.FlightTrajectory

		if err := rows.Scan(
			&item.ID,
			&item.IdentityKey,
			(*string)(&item.IdentityBasis),
			(*string)(&item.SplitReason),
			&item.FlightID,
			&item.AircraftID,
			&item.ICAO24,
			&item.Callsign,
			&item.StartTime,
			&item.EndTime,
			&item.DurationSeconds,
			&item.SegmentCount,
			&item.PointCount,
			&item.CoverageGapCount,
			&item.QualityScore,
			&item.SourceName,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan analytical trajectory: %w",
				err,
			)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate analytical trajectories: %w",
			err,
		)
	}

	return items, nil
}
