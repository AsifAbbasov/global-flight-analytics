package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

var ErrTrajectoryRelationalIntegrity = errors.New(
	"trajectory relational integrity violation",
)

func validateTrajectoryRelationalIntegrity(
	item trajectory.FlightTrajectory,
) error {
	if item.SegmentCount != len(item.Segments) {
		return fmt.Errorf(
			"%w: segment_count=%d does not match segments=%d",
			ErrTrajectoryRelationalIntegrity,
			item.SegmentCount,
			len(item.Segments),
		)
	}
	if item.CoverageGapCount != len(item.CoverageGaps) {
		return fmt.Errorf(
			"%w: coverage_gap_count=%d does not match coverage_gaps=%d",
			ErrTrajectoryRelationalIntegrity,
			item.CoverageGapCount,
			len(item.CoverageGaps),
		)
	}
	if item.SegmentCount < 0 || item.PointCount < 0 || item.CoverageGapCount < 0 {
		return fmt.Errorf(
			"%w: stored counts must be non-negative",
			ErrTrajectoryRelationalIntegrity,
		)
	}

	normalizedICAO24 := strings.ToUpper(strings.TrimSpace(item.ICAO24))
	segmentPointCount := 0

	for index, segment := range item.Segments {
		expectedSequenceNumber := index + 1
		if segment.SequenceNumber != expectedSequenceNumber {
			return fmt.Errorf(
				"%w: segment index %d has sequence_number=%d, expected %d",
				ErrTrajectoryRelationalIntegrity,
				index,
				segment.SequenceNumber,
				expectedSequenceNumber,
			)
		}
		if segment.PointCount < 0 {
			return fmt.Errorf(
				"%w: segment sequence_number=%d has negative point_count",
				ErrTrajectoryRelationalIntegrity,
				segment.SequenceNumber,
			)
		}
		if strings.ToUpper(strings.TrimSpace(segment.ICAO24)) != normalizedICAO24 {
			return fmt.Errorf(
				"%w: segment sequence_number=%d belongs to a different icao24",
				ErrTrajectoryRelationalIntegrity,
				segment.SequenceNumber,
			)
		}

		segmentPointCount += segment.PointCount
	}

	if item.PointCount != segmentPointCount {
		return fmt.Errorf(
			"%w: point_count=%d does not match segment point total=%d",
			ErrTrajectoryRelationalIntegrity,
			item.PointCount,
			segmentPointCount,
		)
	}
	if len(item.Points) > 0 && item.PointCount != len(item.Points) {
		return fmt.Errorf(
			"%w: point_count=%d does not match in-memory points=%d",
			ErrTrajectoryRelationalIntegrity,
			item.PointCount,
			len(item.Points),
		)
	}

	for index, coverageGap := range item.CoverageGaps {
		if strings.ToUpper(strings.TrimSpace(coverageGap.ICAO24)) != normalizedICAO24 {
			return fmt.Errorf(
				"%w: coverage gap index %d belongs to a different icao24",
				ErrTrajectoryRelationalIntegrity,
				index,
			)
		}
	}

	return nil
}
