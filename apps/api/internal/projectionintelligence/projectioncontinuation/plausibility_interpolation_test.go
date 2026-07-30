package projectioncontinuation

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestInterpolationPlausibilityRejectsUnsafeSegments(
	t *testing.T,
) {
	start := continuationTestAsOfTime().
		Add(-time.Hour)
	basePolicy := DefaultPlausibilityPolicy()

	tests := []struct {
		name       string
		rightTime  time.Time
		rightLat   float64
		rightLon   float64
		rightAltM  float64
		policy     PlausibilityPolicy
		targetTime time.Time
	}{
		{
			name:      "gap",
			rightTime: start.Add(6 * time.Minute),
			rightLat:  0,
			rightLon:  0.02,
			rightAltM: 1200,
			policy:    basePolicy,
			targetTime: start.Add(
				3 * time.Minute,
			),
		},
		{
			name:      "horizontal speed",
			rightTime: start.Add(time.Minute),
			rightLat:  0,
			rightLon:  1,
			rightAltM: 1200,
			policy:    basePolicy,
			targetTime: start.Add(
				30 * time.Second,
			),
		},
		{
			name:      "vertical speed",
			rightTime: start.Add(time.Minute),
			rightLat:  0,
			rightLon:  0.001,
			rightAltM: 10000,
			policy:    basePolicy,
			targetTime: start.Add(
				30 * time.Second,
			),
		},
		{
			name:      "exact endpoint still validates segment",
			rightTime: start.Add(time.Minute),
			rightLat:  0,
			rightLon:  1,
			rightAltM: 1200,
			policy:    basePolicy,
			targetTime: start.Add(
				time.Minute,
			),
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				points := plausibilityTestPoints(
					start,
					test.rightTime,
					test.rightLat,
					test.rightLon,
					test.rightAltM,
				)
				_, decision :=
					interpolateTrajectoryPointWithPolicy(
						points,
						test.targetTime,
						test.policy,
					)
				if decision !=
					interpolationRejectedByPlausibility {
					t.Fatalf(
						"decision = %v, want plausibility rejection",
						decision,
					)
				}
			},
		)
	}
}

func TestInterpolationPlausibilityAcceptsValidSegment(
	t *testing.T,
) {
	start := continuationTestAsOfTime().
		Add(-time.Hour)
	points := plausibilityTestPoints(
		start,
		start.Add(2*time.Minute),
		0,
		0.02,
		1200,
	)

	_, decision :=
		interpolateTrajectoryPointWithPolicy(
			points,
			start.Add(time.Minute),
			DefaultPlausibilityPolicy(),
		)
	if decision != interpolationAccepted {
		t.Fatalf(
			"decision = %v, want accepted",
			decision,
		)
	}
}

func plausibilityTestPoints(
	leftTime time.Time,
	rightTime time.Time,
	rightLatitude float64,
	rightLongitude float64,
	rightAltitudeM float64,
) []trajectory.TrackPoint4D {
	return []trajectory.TrackPoint4D{
		{
			Latitude:           0,
			Longitude:          0,
			GeometricAltitudeM: 1000,
			GeometricAltitudeStatus: flightstate.
				AltitudeStatusObserved,
			ObservedAt: leftTime,
		},
		{
			Latitude:           rightLatitude,
			Longitude:          rightLongitude,
			GeometricAltitudeM: rightAltitudeM,
			GeometricAltitudeStatus: flightstate.
				AltitudeStatusObserved,
			ObservedAt: rightTime,
		},
	}
}
