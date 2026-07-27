package trajectorybuilder

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestBuilderFiltersWindowAndCollapsesDuplicateTimestamps(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	item := trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(30 * time.Second), PointCount: 6, QualityScore: 0.8,
		Points: []trajectory.TrackPoint4D{
			{ID: "before", ObservedAt: start.Add(-time.Second), Latitude: 0, Longitude: -1},
			{ID: "late", ObservedAt: start.Add(30 * time.Second), Latitude: 0, Longitude: 2},
			{ID: "duplicate-b", ObservedAt: start.Add(10 * time.Second), Latitude: 0, Longitude: 9},
			{ID: "early", ObservedAt: start, Latitude: 0, Longitude: 0},
			{ID: "duplicate-a", ObservedAt: start.Add(10 * time.Second), Latitude: 0, Longitude: 1},
			{ID: "after", ObservedAt: start.Add(31 * time.Second), Latitude: 0, Longitude: 3},
		},
	}
	features, err := New().Build(context.Background(), item)
	if err != nil {
		t.Fatal(err)
	}
	if features.PointCount != 3 || features.Evidence.SupportingPointCount != 3 {
		t.Fatalf("features=%#v", features)
	}
	if features.MeanSamplingIntervalSeconds != 15 || features.MaximumSamplingGapSeconds != 20 {
		t.Fatalf("sampling=%#v", features)
	}
	for _, code := range []string{
		flightfeatures.TrajectoryLimitationPointOutsideWindow,
		flightfeatures.TrajectoryLimitationDuplicateTimestampsCollapsed,
	} {
		if !hasLimitation(features.Evidence.Limitations, code) {
			t.Fatalf("missing %q in %#v", code, features.Evidence.Limitations)
		}
	}
}

func TestCanonicalMetricsArePermutationInvariant(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	points := []trajectory.TrackPoint4D{
		{ID: "a", ObservedAt: start, Latitude: 0, Longitude: 0},
		{ID: "b", ObservedAt: start.Add(time.Second), Latitude: 1, Longitude: 1},
		{ID: "c", ObservedAt: start.Add(2 * time.Second), Latitude: 0, Longitude: 2},
	}
	firstItem := trajectory.FlightTrajectory{ID: "trajectory", StartTime: start, EndTime: start.Add(2 * time.Second), PointCount: 3, QualityScore: 1, Points: points}
	secondItem := firstItem
	secondItem.Points = []trajectory.TrackPoint4D{points[2], points[0], points[1]}
	first, err := New().Build(context.Background(), firstItem)
	if err != nil {
		t.Fatal(err)
	}
	second, err := New().Build(context.Background(), secondItem)
	if err != nil {
		t.Fatal(err)
	}
	if first.PointCount != second.PointCount ||
		first.MeanSamplingIntervalSeconds != second.MeanSamplingIntervalSeconds ||
		first.MaximumSamplingGapSeconds != second.MaximumSamplingGapSeconds ||
		first.PathEfficiencyRatio != second.PathEfficiencyRatio {
		t.Fatalf("canonical metrics differ: first=%#v second=%#v", first, second)
	}
}

func TestBuilderUsesPersistedPointCountWhenRecordsAreUnmaterialized(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	features, err := New().Build(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(time.Minute), PointCount: 10, SegmentCount: 1, QualityScore: 0.7,
		Segments: []trajectory.TrajectorySegment{{ID: "segment", Status: trajectory.SegmentStatusObserved, StartTime: start, EndTime: start.Add(time.Minute), StartLatitude: 0, StartLongitude: 0, EndLatitude: 0, EndLongitude: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if features.PointCount != 10 || features.Evidence.SupportingPointCount != 10 {
		t.Fatalf("features=%#v", features)
	}
	if !hasLimitation(features.Evidence.Limitations, flightfeatures.TrajectoryLimitationPointRecordsUnmaterialized) {
		t.Fatalf("limitations=%#v", features.Evidence.Limitations)
	}
}

func TestBuilderObservesCancellationDuringLargeCanonicalization(t *testing.T) {
	start := time.Date(2026, time.July, 27, 8, 0, 0, 0, time.UTC)
	points := make([]trajectory.TrackPoint4D, 200000)
	for index := range points {
		points[index] = trajectory.TrackPoint4D{
			ID:         "point",
			ObservedAt: start.Add(time.Duration(index) * time.Millisecond),
			Latitude:   40,
			Longitude:  49,
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := New().Build(ctx, trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(time.Duration(len(points)) * time.Millisecond), PointCount: len(points), QualityScore: 1, Points: points,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
