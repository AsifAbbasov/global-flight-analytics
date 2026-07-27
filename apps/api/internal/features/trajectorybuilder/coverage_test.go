package trajectorybuilder

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestCoverageMergesAndClipsGaps(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	end := start.Add(100 * time.Second)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: end, PointCount: 1, SegmentCount: 1, CoverageGapCount: 4,
		Segments: []trajectory.TrajectorySegment{{ID: "segment", Status: trajectory.SegmentStatusObserved, StartTime: start, EndTime: end}},
		CoverageGaps: []trajectory.CoverageGap{
			{ID: "left", StartTime: start.Add(-10 * time.Second), EndTime: start.Add(20 * time.Second), DurationSeconds: 30},
			{ID: "overlap", StartTime: start.Add(10 * time.Second), EndTime: start.Add(40 * time.Second), DurationSeconds: 30},
			{ID: "right", StartTime: start.Add(80 * time.Second), EndTime: end.Add(10 * time.Second), DurationSeconds: 30},
			{ID: "outside", StartTime: end.Add(10 * time.Second), EndTime: end.Add(20 * time.Second), DurationSeconds: 10},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculateCoverageRatio(context.Background(), evidence, summarizeCoverageSegments(evidence))
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || !approximatelyEqual(metric.value, 0.4, 1e-12) {
		t.Fatalf("coverage=%#v", metric)
	}
	if !hasLimitation(limitations, flightfeatures.TrajectoryLimitationCoverageGapOutsideWindow) {
		t.Fatalf("limitations=%#v", limitations)
	}
}

func TestCoverageDetectsZeroDurationMetadataMismatch(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(time.Minute), PointCount: 1, SegmentCount: 1, CoverageGapCount: 1,
		Segments:     []trajectory.TrajectorySegment{{ID: "segment", Status: trajectory.SegmentStatusObserved, StartTime: start, EndTime: start.Add(time.Minute)}},
		CoverageGaps: []trajectory.CoverageGap{{ID: "gap", StartTime: start.Add(10 * time.Second), EndTime: start.Add(20 * time.Second), DurationSeconds: 0}},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculateCoverageRatio(context.Background(), evidence, summarizeCoverageSegments(evidence))
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || !hasLimitation(limitations, flightfeatures.TrajectoryLimitationCoverageGapDurationMismatch) {
		t.Fatalf("metric=%#v limitations=%#v", metric, limitations)
	}
}

func TestCoverageRejectsInvalidGapWindow(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(time.Minute), PointCount: 1, SegmentCount: 1, CoverageGapCount: 1,
		Segments:     []trajectory.TrajectorySegment{{ID: "segment", Status: trajectory.SegmentStatusObserved, StartTime: start, EndTime: start.Add(time.Minute)}},
		CoverageGaps: []trajectory.CoverageGap{{ID: "reversed", StartTime: start.Add(20 * time.Second), EndTime: start.Add(10 * time.Second)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculateCoverageRatio(context.Background(), evidence, summarizeCoverageSegments(evidence))
	if err != nil {
		t.Fatal(err)
	}
	if metric.available || !hasLimitation(limitations, flightfeatures.TrajectoryLimitationCoverageGapWindowInvalid) {
		t.Fatalf("metric=%#v limitations=%#v", metric, limitations)
	}
}

func TestCoverageZeroDurationIgnoresOutsideGap(t *testing.T) {
	instant := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: instant, EndTime: instant, PointCount: 1, CoverageGapCount: 1,
		Points:       []trajectory.TrackPoint4D{{ID: "point", ObservedAt: instant, Latitude: 0, Longitude: 0}},
		CoverageGaps: []trajectory.CoverageGap{{ID: "outside", StartTime: instant.Add(time.Second), EndTime: instant.Add(2 * time.Second), DurationSeconds: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculateCoverageRatio(context.Background(), evidence, summarizeCoverageSegments(evidence))
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || metric.value != 1 || !hasLimitation(limitations, flightfeatures.TrajectoryLimitationCoverageGapOutsideWindow) {
		t.Fatalf("metric=%#v limitations=%#v", metric, limitations)
	}
}

func TestCoverageUsesSubsecondDurationsWithoutRoundingLoss(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(2 * time.Second), PointCount: 1, SegmentCount: 1, CoverageGapCount: 1,
		Segments:     []trajectory.TrajectorySegment{{ID: "segment", Status: trajectory.SegmentStatusObserved, StartTime: start, EndTime: start.Add(2 * time.Second)}},
		CoverageGaps: []trajectory.CoverageGap{{ID: "subsecond", StartTime: start.Add(250 * time.Millisecond), EndTime: start.Add(750 * time.Millisecond), DurationSeconds: 0}},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculateCoverageRatio(context.Background(), evidence, summarizeCoverageSegments(evidence))
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || metric.value != 0.75 || hasLimitation(limitations, flightfeatures.TrajectoryLimitationCoverageGapDurationMismatch) {
		t.Fatalf("metric=%#v limitations=%#v", metric, limitations)
	}
}

func TestCoverageRequiresObservationEvidence(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{StartTime: start, EndTime: start.Add(time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculateCoverageRatio(context.Background(), evidence, summarizeCoverageSegments(evidence))
	if err != nil {
		t.Fatal(err)
	}
	if metric.available || !hasLimitation(limitations, flightfeatures.TrajectoryLimitationCoverageObservationEvidenceUnavailable) {
		t.Fatalf("metric=%#v limitations=%#v", metric, limitations)
	}
}

func TestUnionDurationDoesNotDoubleCountOverlap(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	duration, err := unionDurationContext(context.Background(), []timeInterval{
		{start: start.Add(10 * time.Second), end: start.Add(30 * time.Second)},
		{start: start.Add(20 * time.Second), end: start.Add(40 * time.Second)},
		{start: start.Add(50 * time.Second), end: start.Add(60 * time.Second)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if duration != 40*time.Second {
		t.Fatalf("duration=%v", duration)
	}
}

func summarizeCoverageSegments(evidence canonicalEvidence) segmentStatusSummary {
	summary := segmentStatusSummary{available: evidence.segmentCountAvailable}
	for _, segment := range evidence.segments {
		switch segment.Status {
		case trajectory.SegmentStatusObserved:
			summary.observedCount++
		case trajectory.SegmentStatusInterpolated:
			summary.interpolatedCount++
		case trajectory.SegmentStatusEstimated:
			summary.estimatedCount++
		case trajectory.SegmentStatusInvalid:
			summary.invalidCount++
		}
	}
	return summary
}
