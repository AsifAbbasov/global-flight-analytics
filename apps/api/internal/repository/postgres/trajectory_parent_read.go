package postgres

import (
	"context"
	"errors"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
	"strings"
)

func (
	repository *TrajectoryRepository,
) getLatestTrajectoryByICAO24(
	ctx context.Context,
	icao24 string,
) (trajectory.FlightTrajectory, error) {
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
) getTrajectoryByID(
	ctx context.Context,
	trajectoryID string,
) (trajectory.FlightTrajectory, error) {
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

	err := repository.trajectoryReadExecutor().QueryRow(
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
func normalizeICAO24Lookup(
	value string,
) string {
	return strings.ToUpper(
		strings.TrimSpace(
			value,
		),
	)
}
