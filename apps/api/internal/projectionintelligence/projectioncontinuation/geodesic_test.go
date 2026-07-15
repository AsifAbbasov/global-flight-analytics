package projectioncontinuation

import (
	"math"
	"testing"
)

func TestDestinationPointAndDistanceRoundTrip(
	t *testing.T,
) {
	latitude, longitude, valid :=
		destinationPoint(
			0,
			0,
			90,
			6000,
		)
	if !valid {
		t.Fatal(
			"destinationPoint() returned invalid",
		)
	}

	distance := greatCircleDistanceM(
		0,
		0,
		latitude,
		longitude,
	)
	if math.Abs(distance-6000) > 0.01 {
		t.Fatalf(
			"distance = %f, want 6000",
			distance,
		)
	}
}

func TestWeightedMeanGeoPointHandlesDateline(
	t *testing.T,
) {
	latitude, longitude, valid :=
		weightedMeanGeoPoint(
			[]weightedGeoPoint{
				{
					latitude:  10,
					longitude: 179.9,
					weight:    1,
				},
				{
					latitude:  10,
					longitude: -179.9,
					weight:    1,
				},
			},
		)
	if !valid {
		t.Fatal(
			"weightedMeanGeoPoint() returned invalid",
		)
	}
	if math.Abs(latitude-10) > 0.001 {
		t.Fatalf(
			"latitude = %f, want approximately 10",
			latitude,
		)
	}
	if math.Abs(
		math.Abs(longitude)-180,
	) > 0.001 {
		t.Fatalf(
			"longitude = %f, want approximately dateline",
			longitude,
		)
	}
}
