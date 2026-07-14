package geographicalbuilder

import (
	"math"
	"testing"
)

func TestHaversineDistanceUsesShortestAntimeridianArc(
	t *testing.T,
) {
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

func TestCircularLongitudeBounds(t *testing.T) {
	tests := []struct {
		name        string
		coordinates []coordinate
		wantMinimum float64
		wantMaximum float64
		wantSpan    float64
	}{
		{
			name: "ordinary interval",
			coordinates: []coordinate{
				{longitude: 10},
				{longitude: 20},
				{longitude: 15},
			},
			wantMinimum: 10,
			wantMaximum: 20,
			wantSpan:    10,
		},
		{
			name: "antimeridian interval",
			coordinates: []coordinate{
				{longitude: 170},
				{longitude: -170},
				{longitude: 179},
			},
			wantMinimum: 170,
			wantMaximum: -170,
			wantSpan:    20,
		},
		{
			name: "single longitude",
			coordinates: []coordinate{
				{longitude: -180},
			},
			wantMinimum: -180,
			wantMaximum: -180,
			wantSpan:    0,
		},
		{
			name: "duplicate longitude",
			coordinates: []coordinate{
				{longitude: 42},
				{longitude: 42},
			},
			wantMinimum: 42,
			wantMaximum: 42,
			wantSpan:    0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			minimum, maximum, span :=
				circularLongitudeBounds(
					test.coordinates,
				)

			if !approximatelyEqual(
				minimum,
				test.wantMinimum,
				1e-12,
			) || !approximatelyEqual(
				maximum,
				test.wantMaximum,
				1e-12,
			) || !approximatelyEqual(
				span,
				test.wantSpan,
				1e-12,
			) {
				t.Fatalf(
					"bounds = (%v, %v, %v), want (%v, %v, %v)",
					minimum,
					maximum,
					span,
					test.wantMinimum,
					test.wantMaximum,
					test.wantSpan,
				)
			}
		})
	}
}

func TestObservedPathAndMaximumDisplacement(t *testing.T) {
	coordinates := []coordinate{
		{
			latitude:  0,
			longitude: 0,
		},
		{
			latitude:  0,
			longitude: 1,
		},
		{
			latitude:  0,
			longitude: 0,
		},
	}

	pathDistance := observedPathDistanceKM(coordinates)
	maximumDisplacement :=
		maximumDisplacementKM(coordinates)

	if pathDistance < 222 || pathDistance > 224 {
		t.Fatalf(
			"path distance = %v, want approximately 222.4",
			pathDistance,
		)
	}
	if maximumDisplacement < 111 ||
		maximumDisplacement > 112 {
		t.Fatalf(
			"maximum displacement = %v, want approximately 111.2",
			maximumDisplacement,
		)
	}
}

func TestPathCrossesAntimeridian(t *testing.T) {
	if !pathCrossesAntimeridian(
		[]coordinate{
			{longitude: 179},
			{longitude: -179},
		},
	) {
		t.Fatal("expected antimeridian crossing")
	}
	if pathCrossesAntimeridian(
		[]coordinate{
			{longitude: 10},
			{longitude: 20},
		},
	) {
		t.Fatal("unexpected antimeridian crossing")
	}
}

func TestNormalizeCoordinate(t *testing.T) {
	valid, ok := normalizeCoordinate(40, 180)
	if !ok ||
		valid.latitude != 40 ||
		valid.longitude != -180 {
		t.Fatalf(
			"normalizeCoordinate() = %#v, %v",
			valid,
			ok,
		)
	}

	invalidValues := [][2]float64{
		{math.NaN(), 0},
		{0, math.Inf(1)},
		{-91, 0},
		{91, 0},
		{0, -181},
		{0, 181},
	}
	for _, values := range invalidValues {
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
}

func TestUniqueGeographicCellCount(t *testing.T) {
	coordinates := []coordinate{
		{
			latitude:  40.001,
			longitude: 49.001,
		},
		{
			latitude:  40.004,
			longitude: 49.004,
		},
		{
			latitude:  40.006,
			longitude: 49.006,
		},
	}

	if count := uniqueGeographicCellCount(
		coordinates,
		2,
	); count != 2 {
		t.Fatalf(
			"cell count = %d, want 2",
			count,
		)
	}
}

func TestShortestLongitudeDelta(t *testing.T) {
	if delta := shortestLongitudeDelta(
		179,
		-179,
	); delta != 2 {
		t.Fatalf("delta = %v, want 2", delta)
	}
	if delta := shortestLongitudeDelta(
		-179,
		179,
	); delta != -2 {
		t.Fatalf("delta = %v, want -2", delta)
	}
}

func approximatelyEqual(
	left float64,
	right float64,
	tolerance float64,
) bool {
	return math.Abs(left-right) <= tolerance
}
