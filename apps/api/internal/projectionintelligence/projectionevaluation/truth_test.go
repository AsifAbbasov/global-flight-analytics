package projectionevaluation

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestTruthAtInterpolatesPositionAndAltitude(
	t *testing.T,
) {
	start := evaluationTestAsOfTime()
	points := []trajectory.TrackPoint4D{
		{
			ID:                 "left",
			Latitude:           0,
			Longitude:          0,
			GeometricAltitudeM: 1000,
			GeometricAltitudeStatus: flightstate.
				AltitudeStatusObserved,
			ObservedAt: start,
		},
		{
			ID:                 "right",
			Latitude:           0,
			Longitude:          0.02,
			GeometricAltitudeM: 1200,
			GeometricAltitudeStatus: flightstate.
				AltitudeStatusObserved,
			ObservedAt: start.Add(2 * time.Minute),
		},
	}

	actual, valid := truthAt(
		points,
		start.Add(time.Minute),
		3*time.Minute,
	)
	if !valid {
		t.Fatal(
			"truthAt() returned invalid",
		)
	}
	if actual.source !=
		ActualPointSourceInterpolated {
		t.Fatalf(
			"source = %q, want interpolated",
			actual.source,
		)
	}
	if math.Abs(
		actual.longitude-0.01,
	) > 1e-6 {
		t.Fatalf(
			"longitude = %f, want approximately 0.01",
			actual.longitude,
		)
	}
	if actual.altitudeM == nil ||
		math.Abs(
			*actual.altitudeM-1100,
		) > 1e-9 {
		t.Fatalf(
			"altitude = %#v, want 1100",
			actual.altitudeM,
		)
	}
}

func TestTruthAtRejectsInterpolationAcrossLargeGap(
	t *testing.T,
) {
	start := evaluationTestAsOfTime()
	points := []trajectory.TrackPoint4D{
		{
			Latitude:   0,
			Longitude:  0,
			ObservedAt: start,
		},
		{
			Latitude:   0,
			Longitude:  0.05,
			ObservedAt: start.Add(5 * time.Minute),
		},
	}

	_, valid := truthAt(
		points,
		start.Add(time.Minute),
		2*time.Minute,
	)
	if valid {
		t.Fatal(
			"truthAt() interpolated across an excessive gap",
		)
	}
}

func TestNormalizeTruthPointsAppliesReplayCutoff(
	t *testing.T,
) {
	asOfTime := evaluationTestAsOfTime()
	evaluatedAt :=
		asOfTime.Add(2 * time.Minute)
	item := trajectory.FlightTrajectory{
		ID: "trajectory",
		Points: []trajectory.TrackPoint4D{
			{
				ID:         "future",
				Latitude:   0,
				Longitude:  3,
				ObservedAt: evaluatedAt.Add(time.Minute),
			},
			{
				ID:         "second",
				Latitude:   0,
				Longitude:  2,
				ObservedAt: asOfTime.Add(2 * time.Minute),
			},
			{
				ID:         "first",
				Latitude:   0,
				Longitude:  1,
				ObservedAt: asOfTime.Add(time.Minute),
			},
			{
				ID:         "past",
				Latitude:   0,
				Longitude:  0,
				ObservedAt: asOfTime.Add(-time.Minute),
			},
		},
	}

	points, excluded :=
		normalizeTruthPoints(
			item,
			asOfTime,
			evaluatedAt,
		)
	if len(points) != 2 ||
		points[0].ID != "first" ||
		points[1].ID != "second" ||
		excluded != 1 {
		t.Fatalf(
			"unexpected truth normalization: points=%#v excluded=%d",
			points,
			excluded,
		)
	}
}

func TestGreatCircleDistanceAcrossDateline(
	t *testing.T,
) {
	distanceM := greatCircleDistanceM(
		0,
		179.9,
		0,
		-179.9,
	)
	if math.Abs(distanceM-22239) > 100 {
		t.Fatalf(
			"distance = %f, want approximately 22239",
			distanceM,
		)
	}
}
