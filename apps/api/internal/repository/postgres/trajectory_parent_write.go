package postgres

import (
	"context"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

func (repository *TrajectoryRepository) insertFlightTrajectory(
	ctx context.Context,
	tx pgx.Tx,
	item trajectory.FlightTrajectory,
) (string, error) {
	const query = `
		INSERT INTO flight_trajectories (
			identity_key,
			identity_basis,
			split_reason,
			flight_id,
			aircraft_id,
			icao24,
			callsign,
			start_time,
			end_time,
			duration_seconds,
			segment_count,
			point_count,
			coverage_gap_count,
			quality_score,
			source_name
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15
		)
		RETURNING id::text;
	`

	var trajectoryID string

	err := tx.QueryRow(
		ctx,
		query,
		item.IdentityKey,
		string(item.IdentityBasis),
		string(item.SplitReason),
		nullableUUID(item.FlightID),
		nullableUUID(item.AircraftID),
		item.ICAO24,
		nullableText(item.Callsign),
		item.StartTime,
		item.EndTime,
		item.DurationSeconds,
		item.SegmentCount,
		item.PointCount,
		item.CoverageGapCount,
		item.QualityScore,
		sourceNameOrUnknown(item.SourceName),
	).Scan(
		&trajectoryID,
	)
	if err != nil {
		return "", err
	}

	return trajectoryID, nil
}
func (repository *TrajectoryRepository) insertReconciledFlightTrajectory(
	ctx context.Context,
	tx pgx.Tx,
	reconciliationTaskID string,
	item trajectory.FlightTrajectory,
) (string, error) {
	const query = `
		INSERT INTO flight_trajectories (
			identity_key,
			identity_basis,
			split_reason,
			flight_id,
			aircraft_id,
			icao24,
			callsign,
			start_time,
			end_time,
			duration_seconds,
			segment_count,
			point_count,
			coverage_gap_count,
			quality_score,
			source_name,
			reconciliation_task_id
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16
		)
		RETURNING id::text;
	`

	var trajectoryID string

	err := tx.QueryRow(
		ctx,
		query,
		item.IdentityKey,
		string(item.IdentityBasis),
		string(item.SplitReason),
		nullableUUID(item.FlightID),
		nullableUUID(item.AircraftID),
		item.ICAO24,
		nullableText(item.Callsign),
		item.StartTime,
		item.EndTime,
		item.DurationSeconds,
		item.SegmentCount,
		item.PointCount,
		item.CoverageGapCount,
		item.QualityScore,
		sourceNameOrUnknown(item.SourceName),
		nullableUUID(reconciliationTaskID),
	).Scan(
		&trajectoryID,
	)
	if err != nil {
		return "", err
	}

	return trajectoryID, nil
}
