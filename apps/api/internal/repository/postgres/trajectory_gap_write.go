package postgres

import (
	"context"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

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
		previousSegmentID := inferredPreviousSegmentID(
			index,
			segmentIDs,
			coverageGap.PreviousSegmentID,
		)
		nextSegmentID := inferredNextSegmentID(
			index,
			segmentIDs,
			coverageGap.NextSegmentID,
		)

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
func inferredPreviousSegmentID(
	index int,
	segmentIDs []string,
	explicitValue string,
) string {
	if explicitValue != "" {
		return explicitValue
	}

	if index >= 0 && index < len(segmentIDs) {
		return segmentIDs[index]
	}

	return ""
}
func inferredNextSegmentID(
	index int,
	segmentIDs []string,
	explicitValue string,
) string {
	if explicitValue != "" {
		return explicitValue
	}

	nextIndex := index + 1

	if nextIndex >= 0 && nextIndex < len(segmentIDs) {
		return segmentIDs[nextIndex]
	}

	return ""
}
