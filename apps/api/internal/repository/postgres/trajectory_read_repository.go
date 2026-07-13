package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

func (
	repository *TrajectoryRepository,
) GetLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	normalizedICAO24 := normalizeICAO24Lookup(
		icao24,
	)
	if normalizedICAO24 == "" {
		return trajectory.FlightTrajectory{},
			trajectory.ErrNotFound
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
		WHERE icao24 = $1
		ORDER BY end_time DESC, start_time DESC, created_at DESC
		LIMIT 1;
	`

	item, err := repository.queryTrajectory(
		ctx,
		query,
		normalizedICAO24,
	)
	if err != nil {
		return trajectory.FlightTrajectory{},
			err
	}

	if err := repository.loadTrajectoryChildren(
		ctx,
		&item,
	); err != nil {
		return trajectory.FlightTrajectory{},
			err
	}

	return item,
		nil
}

func (
	repository *TrajectoryRepository,
) GetTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	trimmedTrajectoryID := strings.TrimSpace(
		trajectoryID,
	)
	if trimmedTrajectoryID == "" {
		return trajectory.FlightTrajectory{},
			trajectory.ErrNotFound
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
		WHERE id = $1
		LIMIT 1;
	`

	item, err := repository.queryTrajectory(
		ctx,
		query,
		trimmedTrajectoryID,
	)
	if err != nil {
		return trajectory.FlightTrajectory{},
			err
	}

	if err := repository.loadTrajectoryChildren(
		ctx,
		&item,
	); err != nil {
		return trajectory.FlightTrajectory{},
			err
	}

	return item,
		nil
}

func (
	repository *TrajectoryRepository,
) queryTrajectory(
	ctx context.Context,
	query string,
	argument string,
) (trajectory.FlightTrajectory, error) {
	var item trajectory.FlightTrajectory

	err := repository.db.QueryRow(
		ctx,
		query,
		argument,
	).Scan(
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
	)
	if err != nil {
		if errors.Is(
			err,
			pgx.ErrNoRows,
		) {
			return trajectory.FlightTrajectory{},
				trajectory.ErrNotFound
		}

		return trajectory.FlightTrajectory{},
			err
	}

	return item,
		nil
}

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

	rows, err := repository.db.Query(
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

	rows, err := repository.db.Query(
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

func normalizeICAO24Lookup(
	value string,
) string {
	return strings.ToUpper(
		strings.TrimSpace(
			value,
		),
	)
}
