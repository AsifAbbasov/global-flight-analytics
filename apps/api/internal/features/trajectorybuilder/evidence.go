package trajectorybuilder

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func canonicalizeEvidence(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) (canonicalEvidence, error) {
	windowStart, windowEnd, windowAvailable, err := resolveTrajectoryWindow(item)
	if err != nil {
		return canonicalEvidence{}, err
	}

	result := canonicalEvidence{
		windowStart:     windowStart,
		windowEnd:       windowEnd,
		windowAvailable: windowAvailable,
		segments:        append([]trajectory.TrajectorySegment(nil), item.Segments...),
		gaps:            append([]trajectory.CoverageGap(nil), item.CoverageGaps...),
	}

	if !windowAvailable && (len(item.Points) > 0 || len(item.Segments) > 0 || len(item.CoverageGaps) > 0) {
		result.limitations = append(result.limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationTemporalWindowUnavailable,
			Message: "Trajectory start and end timestamps are unavailable, so temporal evidence cannot be accepted as canonical trajectory evidence.",
		})
	}

	points, pointLimitations, err := canonicalizePoints(
		ctx,
		item.Points,
		windowStart,
		windowEnd,
		windowAvailable,
	)
	if err != nil {
		return canonicalEvidence{}, err
	}
	result.points = points
	result.limitations = append(result.limitations, pointLimitations...)

	sort.SliceStable(result.segments, func(left int, right int) bool {
		if result.segments[left].SequenceNumber != result.segments[right].SequenceNumber {
			return result.segments[left].SequenceNumber < result.segments[right].SequenceNumber
		}
		if !result.segments[left].StartTime.Equal(result.segments[right].StartTime) {
			return result.segments[left].StartTime.Before(result.segments[right].StartTime)
		}
		return result.segments[left].ID < result.segments[right].ID
	})
	sort.SliceStable(result.gaps, func(left int, right int) bool {
		if !result.gaps[left].StartTime.Equal(result.gaps[right].StartTime) {
			return result.gaps[left].StartTime.Before(result.gaps[right].StartTime)
		}
		if !result.gaps[left].EndTime.Equal(result.gaps[right].EndTime) {
			return result.gaps[left].EndTime.Before(result.gaps[right].EndTime)
		}
		return result.gaps[left].ID < result.gaps[right].ID
	})
	if err := ctx.Err(); err != nil {
		return canonicalEvidence{}, err
	}

	loadedPointCount := len(item.Points)
	switch {
	case loadedPointCount > 0:
		result.pointCount = len(result.points)
		result.pointCountAvailable = true
		result.supportingPointCount = result.pointCount
		if item.PointCount != loadedPointCount {
			result.limitations = append(result.limitations, flightfeatures.FeatureLimitation{
				Code: flightfeatures.LimitationTrajectoryPointCountMetadataMismatch,
				Message: fmt.Sprintf(
					"Trajectory point-count metadata reports %d points while %d point records are materialized.",
					item.PointCount,
					loadedPointCount,
				),
			})
		}
	case item.PointCount > 0:
		result.pointCount = item.PointCount
		result.pointCountAvailable = true
		result.supportingPointCount = item.PointCount
		result.limitations = append(result.limitations, flightfeatures.FeatureLimitation{
			Code:    flightfeatures.TrajectoryLimitationPointRecordsUnmaterialized,
			Message: "Trajectory point records are not materialized, so point count uses persisted metadata while sampling and point-path metrics remain unavailable.",
		})
	}

	authoritativeEnvelope := windowAvailable && (result.pointCountAvailable ||
		len(item.Segments) > 0 ||
		len(item.CoverageGaps) > 0 ||
		item.SegmentCount > 0 ||
		item.CoverageGapCount > 0)

	if len(item.Segments) > 0 || item.SegmentCount > 0 || authoritativeEnvelope {
		result.segmentCount = len(item.Segments)
		result.segmentCountAvailable = true
	}
	if len(item.CoverageGaps) > 0 || item.CoverageGapCount > 0 || authoritativeEnvelope {
		result.gapCount = len(item.CoverageGaps)
		result.gapCountAvailable = true
	}

	if result.segmentCountAvailable && item.SegmentCount != len(item.Segments) {
		result.limitations = append(result.limitations, flightfeatures.FeatureLimitation{
			Code: "trajectory_segment_count_metadata_mismatch",
			Message: fmt.Sprintf(
				"Trajectory segment-count metadata reports %d segments while %d segment records are materialized.",
				item.SegmentCount,
				len(item.Segments),
			),
		})
	}
	if result.gapCountAvailable && item.CoverageGapCount != len(item.CoverageGaps) {
		result.limitations = append(result.limitations, flightfeatures.FeatureLimitation{
			Code: "trajectory_coverage_gap_count_metadata_mismatch",
			Message: fmt.Sprintf(
				"Trajectory coverage-gap metadata reports %d gaps while %d gap records are materialized.",
				item.CoverageGapCount,
				len(item.CoverageGaps),
			),
		})
	}

	return result, nil
}

func resolveTrajectoryWindow(
	item trajectory.FlightTrajectory,
) (time.Time, time.Time, bool, error) {
	if item.StartTime.IsZero() && item.EndTime.IsZero() {
		return time.Time{}, time.Time{}, false, nil
	}
	if item.StartTime.IsZero() {
		return time.Time{}, time.Time{}, false, ErrTrajectoryStartTimeRequired
	}
	if item.EndTime.IsZero() {
		return time.Time{}, time.Time{}, false, ErrTrajectoryEndTimeRequired
	}
	if item.EndTime.Before(item.StartTime) {
		return time.Time{}, time.Time{}, false, ErrInvalidTrajectoryWindow
	}
	return item.StartTime.UTC(), item.EndTime.UTC(), true, nil
}

func canonicalizePoints(
	ctx context.Context,
	points []trajectory.TrackPoint4D,
	windowStart time.Time,
	windowEnd time.Time,
	windowAvailable bool,
) ([]canonicalPoint, []flightfeatures.FeatureLimitation, error) {
	if !windowAvailable {
		return nil, nil, nil
	}
	observed := make([]canonicalPoint, 0, len(points))
	missingTimestampCount := 0
	outsideWindowCount := 0
	nonMonotonicCount := 0
	invalidCoordinateCount := 0
	var previousInputTimestamp time.Time

	for index, point := range points {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
		}
		if point.ObservedAt.IsZero() {
			missingTimestampCount++
			continue
		}
		observedAt := point.ObservedAt.UTC()
		if !previousInputTimestamp.IsZero() && observedAt.Before(previousInputTimestamp) {
			nonMonotonicCount++
		}
		previousInputTimestamp = observedAt
		if windowAvailable && (observedAt.Before(windowStart) || observedAt.After(windowEnd)) {
			outsideWindowCount++
			continue
		}
		value, valid := normalizeCoordinate(point.Latitude, point.Longitude)
		if !valid {
			invalidCoordinateCount++
		}
		observed = append(observed, canonicalPoint{
			point:               point,
			observedAt:          observedAt,
			coordinate:          value,
			coordinateAvailable: valid,
			inputIndex:          index,
		})
	}

	sort.SliceStable(observed, func(left int, right int) bool {
		if !observed[left].observedAt.Equal(observed[right].observedAt) {
			return observed[left].observedAt.Before(observed[right].observedAt)
		}
		if observed[left].point.ID != observed[right].point.ID {
			return observed[left].point.ID < observed[right].point.ID
		}
		if observed[left].point.FlightStateID != observed[right].point.FlightStateID {
			return observed[left].point.FlightStateID < observed[right].point.FlightStateID
		}
		if canonicalPointTieBreakLess(observed[left].point, observed[right].point) {
			return true
		}
		if canonicalPointTieBreakLess(observed[right].point, observed[left].point) {
			return false
		}
		// The original index is consulted only for semantically identical records;
		// choosing either record therefore cannot change any derived metric.
		return observed[left].inputIndex < observed[right].inputIndex
	})

	unique := make([]canonicalPoint, 0, len(observed))
	duplicateTimestampCount := 0
	for _, point := range observed {
		if len(unique) > 0 && point.observedAt.Equal(unique[len(unique)-1].observedAt) {
			duplicateTimestampCount++
			continue
		}
		unique = append(unique, point)
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	limitations := make([]flightfeatures.FeatureLimitation, 0, 5)
	if missingTimestampCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPointTimestampMissing,
			Message: fmt.Sprintf(
				"%d trajectory point records have no observation timestamp and were excluded from canonical evidence.",
				missingTimestampCount,
			),
		})
	}
	if outsideWindowCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPointOutsideWindow,
			Message: fmt.Sprintf(
				"%d trajectory point records lie outside the authoritative trajectory window and were excluded.",
				outsideWindowCount,
			),
		})
	}
	if nonMonotonicCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationPointOrderCanonicalized,
			Message: fmt.Sprintf(
				"%d point-order decreases were detected; all trajectory metrics use one deterministic chronological point order.",
				nonMonotonicCount,
			),
		})
	}
	if duplicateTimestampCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationDuplicateTimestampsCollapsed,
			Message: fmt.Sprintf(
				"%d duplicate point timestamps were collapsed by selecting the first deterministic point identity for each instant.",
				duplicateTimestampCount,
			),
		})
	}
	if invalidCoordinateCount > 0 {
		limitations = append(limitations, flightfeatures.FeatureLimitation{
			Code: flightfeatures.TrajectoryLimitationInvalidPointCoordinates,
			Message: fmt.Sprintf(
				"%d canonical point observations have invalid coordinates and break point-path continuity while remaining eligible for timestamp sampling.",
				invalidCoordinateCount,
			),
		})
	}
	return unique, limitations, nil
}

func canonicalPointTieBreakLess(
	left trajectory.TrackPoint4D,
	right trajectory.TrackPoint4D,
) bool {
	// TrackPoint4D is a value-only evidence object with no maps or slices, so its
	// Go-syntax representation provides a stable complete tie-break key for
	// records sharing one observation instant.
	return fmt.Sprintf("%#v", left) < fmt.Sprintf("%#v", right)
}
