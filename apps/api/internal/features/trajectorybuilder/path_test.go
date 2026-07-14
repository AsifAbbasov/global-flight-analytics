package trajectorybuilder

import (
	"context"
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestCalculatePathEfficiencyForStraightPath(
	t *testing.T,
) {
	item := trajectory.FlightTrajectory{
		Points: []trajectory.TrackPoint4D{
			{Latitude: 0, Longitude: 0},
			{Latitude: 0, Longitude: 1},
			{Latitude: 0, Longitude: 2},
		},
	}

	metric, limitations :=
		calculatePathEfficiency(
			context.Background(),
			item,
		)

	if !metric.available ||
		!approximatelyEqual(metric.value, 1, 1e-12) {
		t.Fatalf(
			"path efficiency = %#v, want 1",
			metric,
		)
	}
	if len(limitations) != 0 {
		t.Fatalf(
			"unexpected limitations: %#v",
			limitations,
		)
	}
}

func TestCalculatePathEfficiencyForDetour(t *testing.T) {
	item := trajectory.FlightTrajectory{
		Points: []trajectory.TrackPoint4D{
			{Latitude: 0, Longitude: 0},
			{Latitude: 1, Longitude: 1},
			{Latitude: 0, Longitude: 2},
		},
	}

	metric, limitations :=
		calculatePathEfficiency(
			context.Background(),
			item,
		)

	if !metric.available ||
		metric.value <= 0 ||
		metric.value >= 1 {
		t.Fatalf(
			"detour efficiency = %#v",
			metric,
		)
	}
	if len(limitations) != 0 {
		t.Fatalf(
			"unexpected limitations: %#v",
			limitations,
		)
	}
}

func TestCalculatePathEfficiencyUsesShortestAntimeridianArc(
	t *testing.T,
) {
	item := trajectory.FlightTrajectory{
		Points: []trajectory.TrackPoint4D{
			{Latitude: 10, Longitude: 170},
			{Latitude: 10, Longitude: 179},
			{Latitude: 10, Longitude: -170},
		},
	}

	metric, limitations :=
		calculatePathEfficiency(
			context.Background(),
			item,
		)

	if !metric.available ||
		metric.value <= 0.99 ||
		metric.value > 1 {
		t.Fatalf(
			"antimeridian efficiency = %#v",
			metric,
		)
	}
	if len(limitations) != 0 {
		t.Fatalf(
			"unexpected limitations: %#v",
			limitations,
		)
	}
}

func TestCalculatePathEfficiencyFallsBackToSegments(
	t *testing.T,
) {
	item := trajectory.FlightTrajectory{
		Points: []trajectory.TrackPoint4D{
			{
				Latitude:  100,
				Longitude: 0,
			},
		},
		Segments: []trajectory.TrajectorySegment{
			{
				ID:             "second",
				SequenceNumber: 2,
				Status:         trajectory.SegmentStatusObserved,
				StartLatitude:  0,
				StartLongitude: 1,
				EndLatitude:    0,
				EndLongitude:   2,
			},
			{
				ID:             "invalid",
				SequenceNumber: 3,
				Status:         trajectory.SegmentStatusInvalid,
				StartLatitude:  0,
				StartLongitude: 2,
				EndLatitude:    0,
				EndLongitude:   20,
			},
			{
				ID:             "first",
				SequenceNumber: 1,
				Status:         trajectory.SegmentStatusObserved,
				StartLatitude:  0,
				StartLongitude: 0,
				EndLatitude:    0,
				EndLongitude:   1,
			},
		},
	}

	metric, limitations :=
		calculatePathEfficiency(
			context.Background(),
			item,
		)

	if !metric.available ||
		!approximatelyEqual(metric.value, 1, 1e-12) {
		t.Fatalf(
			"fallback efficiency = %#v",
			metric,
		)
	}
	if !hasLimitation(
		limitations,
		"trajectory_path_point_evidence_unusable",
	) || !hasLimitation(
		limitations,
		"trajectory_path_segment_endpoint_fallback",
	) {
		t.Fatalf(
			"missing fallback limitations: %#v",
			limitations,
		)
	}
}

func TestCalculatePathEfficiencyRejectsZeroPath(
	t *testing.T,
) {
	item := trajectory.FlightTrajectory{
		Points: []trajectory.TrackPoint4D{
			{Latitude: 40, Longitude: 49},
			{Latitude: 40, Longitude: 49},
		},
	}

	metric, limitations :=
		calculatePathEfficiency(
			context.Background(),
			item,
		)

	if metric.available {
		t.Fatalf(
			"zero path unexpectedly available: %#v",
			metric,
		)
	}
	if !hasLimitation(
		limitations,
		"trajectory_path_efficiency_zero_path",
	) {
		t.Fatalf(
			"missing zero-path limitation: %#v",
			limitations,
		)
	}
}

func TestNormalizeCoordinateAndHaversineDistance(
	t *testing.T,
) {
	value, ok := normalizeCoordinate(40, 180)
	if !ok ||
		value.latitude != 40 ||
		value.longitude != -180 {
		t.Fatalf(
			"normalizeCoordinate() = %#v, %v",
			value,
			ok,
		)
	}

	for _, values := range [][2]float64{
		{math.NaN(), 0},
		{0, math.Inf(1)},
		{-91, 0},
		{91, 0},
		{0, -181},
		{0, 181},
	} {
		if _, ok := normalizeCoordinate(
			values[0],
			values[1],
		); ok {
			t.Fatalf(
				"coordinate (%v, %v) unexpectedly valid",
				values[0],
				values[1],
			)
		}
	}

	distance := haversineDistanceKM(
		coordinate{
			latitude:  0,
			longitude: 179,
		},
		coordinate{
			latitude:  0,
			longitude: -179,
		},
	)
	if distance < 220 || distance > 225 {
		t.Fatalf(
			"distance = %v km, want approximately 222 km",
			distance,
		)
	}
}
