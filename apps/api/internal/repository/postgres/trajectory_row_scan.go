package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

type postgresRowScanner interface {
	Scan(destinations ...any) error
}

func scanFlightTrajectory(
	scanner postgresRowScanner,
) (trajectory.FlightTrajectory, error) {
	var item trajectory.FlightTrajectory
	if err := scanner.Scan(
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
		return trajectory.FlightTrajectory{}, err
	}
	return item, nil
}

func (
	repository *TrajectoryRepository,
) queryFlightTrajectory(
	ctx context.Context,
	query string,
	arguments ...any,
) (trajectory.FlightTrajectory, error) {
	item, err := scanFlightTrajectory(
		repository.trajectoryReadExecutor().QueryRow(
			ctx,
			query,
			arguments...,
		),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return trajectory.FlightTrajectory{}, trajectory.ErrNotFound
		}
		return trajectory.FlightTrajectory{}, err
	}
	return item, nil
}

func scanFlightTrajectoryRows(
	rows pgx.Rows,
	operation string,
) ([]trajectory.FlightTrajectory, error) {
	items := make([]trajectory.FlightTrajectory, 0)
	for rows.Next() {
		item, err := scanFlightTrajectory(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: scan row: %w", operation, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: iterate rows: %w", operation, err)
	}
	return items, nil
}
