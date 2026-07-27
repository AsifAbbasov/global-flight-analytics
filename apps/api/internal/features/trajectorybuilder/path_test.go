package trajectorybuilder

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestPathEfficiencyForStraightCanonicalPath(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(2 * time.Second), PointCount: 3,
		Points: []trajectory.TrackPoint4D{
			{ID: "c", Latitude: 0, Longitude: 2, ObservedAt: start.Add(2 * time.Second)},
			{ID: "a", Latitude: 0, Longitude: 0, ObservedAt: start},
			{ID: "b", Latitude: 0, Longitude: 1, ObservedAt: start.Add(time.Second)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, _, err := calculatePathEfficiency(context.Background(), evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || !approximatelyEqual(metric.value, 1, 1e-12) {
		t.Fatalf("metric=%#v", metric)
	}
}

func TestPathEfficiencyForDetour(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(2 * time.Second), PointCount: 3,
		Points: []trajectory.TrackPoint4D{
			{ID: "a", Latitude: 0, Longitude: 0, ObservedAt: start},
			{ID: "b", Latitude: 1, Longitude: 1, ObservedAt: start.Add(time.Second)},
			{ID: "c", Latitude: 0, Longitude: 2, ObservedAt: start.Add(2 * time.Second)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, _, err := calculatePathEfficiency(context.Background(), evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || metric.value <= 0 || metric.value >= 1 {
		t.Fatalf("metric=%#v", metric)
	}
}

func TestPathEfficiencyUsesShortestAntimeridianArc(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(2 * time.Second), PointCount: 3,
		Points: []trajectory.TrackPoint4D{
			{ID: "a", Latitude: 10, Longitude: 170, ObservedAt: start},
			{ID: "b", Latitude: 10, Longitude: 179, ObservedAt: start.Add(time.Second)},
			{ID: "c", Latitude: 10, Longitude: -170, ObservedAt: start.Add(2 * time.Second)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, _, err := calculatePathEfficiency(context.Background(), evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || metric.value <= 0.99 || metric.value > 1 {
		t.Fatalf("metric=%#v", metric)
	}
}

func TestPathFallbackUsesIndependentSegments(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(3 * time.Minute), PointCount: 1, SegmentCount: 2,
		Points: []trajectory.TrackPoint4D{{ID: "only", Latitude: 0, Longitude: 0, ObservedAt: start}},
		Segments: []trajectory.TrajectorySegment{
			{ID: "one", SequenceNumber: 1, Status: trajectory.SegmentStatusObserved, StartTime: start, EndTime: start.Add(time.Minute), StartLatitude: 0, StartLongitude: 0, EndLatitude: 0, EndLongitude: 1},
			{ID: "two", SequenceNumber: 2, Status: trajectory.SegmentStatusObserved, StartTime: start.Add(2 * time.Minute), EndTime: start.Add(3 * time.Minute), StartLatitude: 10, StartLongitude: 0, EndLatitude: 10, EndLongitude: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculatePathEfficiency(context.Background(), evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || !approximatelyEqual(metric.value, 1, 1e-12) {
		t.Fatalf("metric=%#v", metric)
	}
	if !hasLimitation(limitations, flightfeatures.TrajectoryLimitationPathSegmentFallback) {
		t.Fatalf("limitations=%#v", limitations)
	}
}

func TestPathFallbackJoinsOnlyContiguousSegments(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(3 * time.Minute), SegmentCount: 3,
		Segments: []trajectory.TrajectorySegment{
			{ID: "one", SequenceNumber: 1, Status: trajectory.SegmentStatusObserved, StartTime: start, EndTime: start.Add(time.Minute), StartLatitude: 0, StartLongitude: 0, EndLatitude: 1, EndLongitude: 1},
			{ID: "two", SequenceNumber: 2, Status: trajectory.SegmentStatusObserved, StartTime: start.Add(time.Minute), EndTime: start.Add(2 * time.Minute), StartLatitude: 1, StartLongitude: 1, EndLatitude: 0, EndLongitude: 2},
			{ID: "three", SequenceNumber: 3, Status: trajectory.SegmentStatusObserved, StartTime: start.Add(2 * time.Minute), EndTime: start.Add(3 * time.Minute), StartLatitude: 10, StartLongitude: 0, EndLatitude: 10, EndLongitude: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculatePathEfficiency(context.Background(), evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || metric.value <= 0 || metric.value >= 1 {
		t.Fatalf("metric=%#v", metric)
	}
	if !hasLimitation(limitations, flightfeatures.TrajectoryLimitationPathDiscontinuityExcluded) ||
		!hasLimitation(limitations, flightfeatures.TrajectoryLimitationPathSegmentFallback) {
		t.Fatalf("limitations=%#v", limitations)
	}
}

func TestCoverageGapSplitsPointPath(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(3 * time.Second), PointCount: 4, CoverageGapCount: 1,
		Points: []trajectory.TrackPoint4D{
			{ID: "a", Latitude: 0, Longitude: 0, ObservedAt: start},
			{ID: "b", Latitude: 0, Longitude: 1, ObservedAt: start.Add(time.Second)},
			{ID: "c", Latitude: 10, Longitude: 0, ObservedAt: start.Add(2 * time.Second)},
			{ID: "d", Latitude: 10, Longitude: 1, ObservedAt: start.Add(3 * time.Second)},
		},
		CoverageGaps: []trajectory.CoverageGap{{ID: "gap", StartTime: start.Add(time.Second), EndTime: start.Add(2 * time.Second), DurationSeconds: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculatePathEfficiency(context.Background(), evidence)
	if err != nil {
		t.Fatal(err)
	}
	if !metric.available || !approximatelyEqual(metric.value, 1, 1e-12) {
		t.Fatalf("metric=%#v", metric)
	}
	if !hasLimitation(limitations, flightfeatures.TrajectoryLimitationPathDiscontinuityExcluded) {
		t.Fatalf("limitations=%#v", limitations)
	}
}

func TestPathEfficiencyRejectsZeroPath(t *testing.T) {
	start := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: start, EndTime: start.Add(time.Second), PointCount: 2,
		Points: []trajectory.TrackPoint4D{
			{ID: "a", Latitude: 40, Longitude: 49, ObservedAt: start},
			{ID: "b", Latitude: 40, Longitude: 49, ObservedAt: start.Add(time.Second)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	metric, limitations, err := calculatePathEfficiency(context.Background(), evidence)
	if err != nil {
		t.Fatal(err)
	}
	if metric.available || !hasLimitation(limitations, flightfeatures.TrajectoryLimitationPathZeroDistance) {
		t.Fatalf("metric=%#v limitations=%#v", metric, limitations)
	}
}

func TestNormalizeCoordinateAndHaversineDistance(t *testing.T) {
	value, ok := normalizeCoordinate(40, 180)
	if !ok || value.latitude != 40 || value.longitude != -180 {
		t.Fatalf("value=%#v ok=%v", value, ok)
	}
	for _, values := range [][2]float64{{math.NaN(), 0}, {0, math.Inf(1)}, {-91, 0}, {91, 0}, {0, -181}, {0, 181}} {
		if _, ok := normalizeCoordinate(values[0], values[1]); ok {
			t.Fatalf("coordinate (%v,%v) valid", values[0], values[1])
		}
	}
	distance := haversineDistanceKM(coordinate{latitude: 0, longitude: 179}, coordinate{latitude: 0, longitude: -179})
	if distance < 220 || distance > 225 {
		t.Fatalf("distance=%v", distance)
	}
}
