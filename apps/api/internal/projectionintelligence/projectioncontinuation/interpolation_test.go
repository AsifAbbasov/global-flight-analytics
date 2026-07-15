package projectioncontinuation

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestInterpolateTrajectoryPointInterpolatesPositionAndAltitude(
	t *testing.T,
) {
	start := continuationTestAsOfTime().
		Add(-time.Hour)
	points := []trajectory.TrackPoint4D{
		{
			Latitude:           0,
			Longitude:          0,
			GeometricAltitudeM: 1000,
			GeometricAltitudeStatus: flightstate.
				AltitudeStatusObserved,
			ObservedAt: start,
		},
		{
			Latitude:           0,
			Longitude:          0.02,
			GeometricAltitudeM: 1200,
			GeometricAltitudeStatus: flightstate.
				AltitudeStatusObserved,
			ObservedAt: start.Add(
				2 * time.Minute,
			),
		},
	}

	point, valid :=
		interpolateTrajectoryPoint(
			points,
			start.Add(time.Minute),
		)
	if !valid {
		t.Fatal(
			"interpolateTrajectoryPoint() returned invalid",
		)
	}
	if math.Abs(
		point.longitude-0.01,
	) > 1e-6 {
		t.Fatalf(
			"longitude = %f, want approximately 0.01",
			point.longitude,
		)
	}
	if point.altitudeM == nil ||
		math.Abs(
			*point.altitudeM-1100,
		) > 1e-9 {
		t.Fatalf(
			"altitude = %#v, want 1100",
			point.altitudeM,
		)
	}
}

func TestTrajectorySnapshotAtExcludesFutureAndSorts(
	t *testing.T,
) {
	asOfTime := continuationTestAsOfTime()
	item := trajectory.FlightTrajectory{
		ID: "trajectory",
		Points: []trajectory.TrackPoint4D{
			{
				ID:         "future",
				Latitude:   1,
				Longitude:  1,
				ObservedAt: asOfTime.Add(time.Minute),
			},
			{
				ID:         "second",
				Latitude:   1,
				Longitude:  1,
				ObservedAt: asOfTime.Add(-time.Minute),
			},
			{
				ID:         "first",
				Latitude:   1,
				Longitude:  1,
				ObservedAt: asOfTime.Add(-2 * time.Minute),
			},
		},
	}

	snapshot := trajectorySnapshotAt(
		item,
		asOfTime,
	)
	if len(snapshot.Points) != 2 ||
		snapshot.Points[0].ID != "first" ||
		snapshot.Points[1].ID != "second" {
		t.Fatalf(
			"unexpected snapshot: %#v",
			snapshot.Points,
		)
	}
}
