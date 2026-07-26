package extractor

import (
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func validateSnapshotEvidence(
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
) error {
	cutoff := asOfTime.UTC()

	for index, point := range item.Points {
		if !point.ObservedAt.IsZero() && point.ObservedAt.After(cutoff) {
			return fmt.Errorf(
				"%w: point[%d] observed_at=%s as_of=%s",
				ErrTrajectoryPointAfterAsOf,
				index,
				point.ObservedAt.UTC().Format(time.RFC3339Nano),
				cutoff.Format(time.RFC3339Nano),
			)
		}
	}

	for index, segment := range item.Segments {
		if (!segment.StartTime.IsZero() && segment.StartTime.After(cutoff)) ||
			(!segment.EndTime.IsZero() && segment.EndTime.After(cutoff)) {
			return fmt.Errorf(
				"%w: segment[%d] start=%s end=%s as_of=%s",
				ErrTrajectorySegmentAfterAsOf,
				index,
				formatOptionalTime(segment.StartTime),
				formatOptionalTime(segment.EndTime),
				cutoff.Format(time.RFC3339Nano),
			)
		}
	}

	for index, gap := range item.CoverageGaps {
		if (!gap.StartTime.IsZero() && gap.StartTime.After(cutoff)) ||
			(!gap.EndTime.IsZero() && gap.EndTime.After(cutoff)) {
			return fmt.Errorf(
				"%w: coverage_gap[%d] start=%s end=%s as_of=%s",
				ErrCoverageGapAfterAsOf,
				index,
				formatOptionalTime(gap.StartTime),
				formatOptionalTime(gap.EndTime),
				cutoff.Format(time.RFC3339Nano),
			)
		}
	}

	return nil
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return "unknown"
	}

	return value.UTC().Format(time.RFC3339Nano)
}
