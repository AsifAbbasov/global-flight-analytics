package postgres

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrajectoryRepository struct {
	db *pgxpool.Pool
}

func NewTrajectoryRepository(db *pgxpool.Pool) *TrajectoryRepository {
	return &TrajectoryRepository{
		db: db,
	}
}

func (repository *TrajectoryRepository) SaveTrajectory(ctx context.Context, item trajectory.FlightTrajectory) error {
	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := repository.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	committed := false

	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	trajectoryID, err := repository.insertFlightTrajectory(ctx, tx, item)
	if err != nil {
		return fmt.Errorf("insert flight trajectory: %w", err)
	}

	segmentIDs, err := repository.insertTrajectorySegments(ctx, tx, trajectoryID, item.Segments)
	if err != nil {
		return fmt.Errorf("insert trajectory segments: %w", err)
	}

	if err := repository.insertCoverageGaps(ctx, tx, trajectoryID, segmentIDs, item.CoverageGaps); err != nil {
		return fmt.Errorf("insert coverage gaps: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	committed = true

	return nil
}

func (repository *TrajectoryRepository) insertFlightTrajectory(
	ctx context.Context,
	tx pgx.Tx,
	item trajectory.FlightTrajectory,
) (string, error) {
	const query = `
		INSERT INTO flight_trajectories (
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
			$7, $8, $9, $10, $11, $12
		)
		RETURNING id::text;
	`

	var trajectoryID string

	err := tx.QueryRow(
		ctx,
		query,
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
	).Scan(&trajectoryID)
	if err != nil {
		return "", err
	}

	return trajectoryID, nil
}

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

	segmentIDs := make([]string, 0, len(segments))

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
			sourceNameOrUnknown(segment.SourceName),
		).Scan(&segmentID)
		if err != nil {
			return nil, err
		}

		segmentIDs = append(segmentIDs, segmentID)
	}

	return segmentIDs, nil
}

func (repository *TrajectoryRepository) insertCoverageGaps(
	ctx context.Context,
	tx pgx.Tx,
	trajectoryID string,
	segmentIDs []string,
	coverageGaps []trajectory.CoverageGap,
) error {
	const query = `
		INSERT INTO coverage_gaps (
			trajectory_id,
			previous_segment_id,
			next_segment_id,
			icao24,
			gap_start_time,
			gap_end_time,
			duration_seconds,
			distance_km,
			reason,
			filled_by
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10
		);
	`

	for index, coverageGap := range coverageGaps {
		previousSegmentID := inferredPreviousSegmentID(index, segmentIDs, coverageGap.PreviousSegmentID)
		nextSegmentID := inferredNextSegmentID(index, segmentIDs, coverageGap.NextSegmentID)

		if _, err := tx.Exec(
			ctx,
			query,
			trajectoryID,
			nullableUUID(previousSegmentID),
			nullableUUID(nextSegmentID),
			coverageGap.ICAO24,
			coverageGap.StartTime,
			coverageGap.EndTime,
			coverageGap.DurationSeconds,
			coverageGap.DistanceKm,
			string(coverageGap.Reason),
			nullableText(coverageGap.FilledBy),
		); err != nil {
			return err
		}
	}

	return nil
}

func inferredPreviousSegmentID(index int, segmentIDs []string, explicitValue string) string {
	if explicitValue != "" {
		return explicitValue
	}

	if index >= 0 && index < len(segmentIDs) {
		return segmentIDs[index]
	}

	return ""
}

func inferredNextSegmentID(index int, segmentIDs []string, explicitValue string) string {
	if explicitValue != "" {
		return explicitValue
	}

	nextIndex := index + 1

	if nextIndex >= 0 && nextIndex < len(segmentIDs) {
		return segmentIDs[nextIndex]
	}

	return ""
}
