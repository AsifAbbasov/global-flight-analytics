package projectionbaseline

import (
	"math"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/trajectoryquality"
)

type cutoffSnapshot struct {
	Trajectory trajectory.FlightTrajectory

	ExcludedPointCount   int
	ExcludedSegmentCount int
	ExcludedGapCount     int

	QualityEvidenceAvailable bool
}

func buildCutoffSnapshot(
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
) cutoffSnapshot {
	asOfTime = asOfTime.UTC()

	points, excludedPointCount := cutoffPointsAt(
		item.Points,
		asOfTime,
	)
	segments, excludedSegmentCount := completedSegmentsAt(
		item.Segments,
		asOfTime,
	)
	gaps, excludedGapCount := completedCoverageGapsAt(
		item.CoverageGaps,
		asOfTime,
	)

	snapshot := item
	snapshot.Points = points
	snapshot.PointCount = len(points)
	snapshot.Segments = segments
	snapshot.SegmentCount = len(segments)
	snapshot.CoverageGaps = gaps
	snapshot.CoverageGapCount = len(gaps)

	qualityAvailable := false
	if len(points) == 0 {
		snapshot.StartTime = time.Time{}
		snapshot.EndTime = time.Time{}
		snapshot.DurationSeconds = 0
		snapshot.QualityScore = 0
		qualityAvailable = true
	} else {
		snapshot.StartTime = points[0].ObservedAt.UTC()
		snapshot.EndTime = points[len(points)-1].ObservedAt.UTC()
		snapshot.DurationSeconds = int64(
			snapshot.EndTime.Sub(snapshot.StartTime) / time.Second,
		)

		if len(segments) > 0 {
			snapshot.QualityScore = trajectoryquality.TrajectoryScore(
				segments,
			)
			qualityAvailable = cutoffQualityScoreValid(snapshot.QualityScore) &&
				qualityEvidenceCoversLatestPoint(
					segments,
					points[len(points)-1],
				)
		} else {
			snapshot.QualityScore = 0
			qualityAvailable = false
		}
	}

	if snapshot.UpdatedAt.After(asOfTime) {
		snapshot.UpdatedAt = asOfTime
	}

	return cutoffSnapshot{
		Trajectory:               snapshot,
		ExcludedPointCount:       excludedPointCount,
		ExcludedSegmentCount:     excludedSegmentCount,
		ExcludedGapCount:         excludedGapCount,
		QualityEvidenceAvailable: qualityAvailable,
	}
}

func (snapshot cutoffSnapshot) excludedEvidenceCount() int {
	return snapshot.ExcludedPointCount +
		snapshot.ExcludedSegmentCount +
		snapshot.ExcludedGapCount
}

func cutoffPointsAt(
	items []trajectory.TrackPoint4D,
	asOfTime time.Time,
) ([]trajectory.TrackPoint4D, int) {
	result := make([]trajectory.TrackPoint4D, 0, len(items))
	excluded := 0

	for _, item := range items {
		if item.ObservedAt.IsZero() ||
			item.ObservedAt.UTC().After(asOfTime) {
			excluded++
			continue
		}
		result = append(result, item)
	}

	sort.SliceStable(result, func(left int, right int) bool {
		leftTime := result[left].ObservedAt.UTC()
		rightTime := result[right].ObservedAt.UTC()
		if !leftTime.Equal(rightTime) {
			return leftTime.Before(rightTime)
		}
		if result[left].SourceName != result[right].SourceName {
			return result[left].SourceName < result[right].SourceName
		}
		if result[left].FlightStateID != result[right].FlightStateID {
			return result[left].FlightStateID < result[right].FlightStateID
		}
		return result[left].ID < result[right].ID
	})

	return result, excluded
}

func completedSegmentsAt(
	items []trajectory.TrajectorySegment,
	asOfTime time.Time,
) ([]trajectory.TrajectorySegment, int) {
	result := make([]trajectory.TrajectorySegment, 0, len(items))
	excluded := 0

	for _, item := range items {
		if item.StartTime.IsZero() ||
			item.EndTime.IsZero() ||
			item.StartTime.UTC().After(asOfTime) ||
			item.EndTime.UTC().After(asOfTime) ||
			item.EndTime.Before(item.StartTime) {
			excluded++
			continue
		}
		result = append(result, item)
	}

	sort.SliceStable(result, func(left int, right int) bool {
		if result[left].SequenceNumber != result[right].SequenceNumber {
			return result[left].SequenceNumber < result[right].SequenceNumber
		}
		leftStart := result[left].StartTime.UTC()
		rightStart := result[right].StartTime.UTC()
		if !leftStart.Equal(rightStart) {
			return leftStart.Before(rightStart)
		}
		return result[left].ID < result[right].ID
	})

	return result, excluded
}

func completedCoverageGapsAt(
	items []trajectory.CoverageGap,
	asOfTime time.Time,
) ([]trajectory.CoverageGap, int) {
	result := make([]trajectory.CoverageGap, 0, len(items))
	excluded := 0

	for _, item := range items {
		if item.StartTime.IsZero() ||
			item.EndTime.IsZero() ||
			item.StartTime.UTC().After(asOfTime) ||
			item.EndTime.UTC().After(asOfTime) ||
			item.EndTime.Before(item.StartTime) {
			excluded++
			continue
		}
		result = append(result, item)
	}

	sort.SliceStable(result, func(left int, right int) bool {
		leftStart := result[left].StartTime.UTC()
		rightStart := result[right].StartTime.UTC()
		if !leftStart.Equal(rightStart) {
			return leftStart.Before(rightStart)
		}
		leftEnd := result[left].EndTime.UTC()
		rightEnd := result[right].EndTime.UTC()
		if !leftEnd.Equal(rightEnd) {
			return leftEnd.Before(rightEnd)
		}
		return result[left].ID < result[right].ID
	})

	return result, excluded
}

func qualityEvidenceCoversLatestPoint(
	segments []trajectory.TrajectorySegment,
	latestPoint trajectory.TrackPoint4D,
) bool {
	if len(segments) == 0 || latestPoint.ObservedAt.IsZero() {
		return false
	}

	latestSegmentEnd := segments[0].EndTime.UTC()
	for _, segment := range segments[1:] {
		endTime := segment.EndTime.UTC()
		if endTime.After(latestSegmentEnd) {
			latestSegmentEnd = endTime
		}
	}

	return !latestSegmentEnd.Before(latestPoint.ObservedAt.UTC())
}

func cutoffQualityScoreValid(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0 &&
		value <= 1
}
