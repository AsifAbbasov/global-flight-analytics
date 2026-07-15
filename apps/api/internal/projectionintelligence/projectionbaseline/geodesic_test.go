package projectionbaseline

import (
	"math"
	"testing"
)

func TestDestinationPointProjectsEastAtEquator(
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

	if math.Abs(latitude) > 1e-9 {
		t.Fatalf(
			"latitude = %.12f, want approximately zero",
			latitude,
		)
	}

	const expectedLongitude = 0.053959221
	if math.Abs(
		longitude-expectedLongitude,
	) > 1e-6 {
		t.Fatalf(
			"longitude = %.12f, want %.12f",
			longitude,
			expectedLongitude,
		)
	}
}

func TestDestinationPointNormalizesHeadingAndLongitude(
	t *testing.T,
) {
	firstLatitude, firstLongitude, firstValid :=
		destinationPoint(
			10,
			179.99,
			450,
			5000,
		)
	secondLatitude, secondLongitude, secondValid :=
		destinationPoint(
			10,
			179.99,
			90,
			5000,
		)

	if !firstValid || !secondValid {
		t.Fatal(
			"destinationPoint() returned invalid",
		)
	}
	if math.Abs(
		firstLatitude-secondLatitude,
	) > 1e-12 ||
		math.Abs(
			firstLongitude-secondLongitude,
		) > 1e-12 {
		t.Fatal(
			"normalized heading changed projection",
		)
	}
	if secondLongitude < -180 ||
		secondLongitude > 180 {
		t.Fatalf(
			"longitude = %f outside normalized range",
			secondLongitude,
		)
	}
}

func TestDestinationPointRejectsInvalidInput(
	t *testing.T,
) {
	_, _, valid := destinationPoint(
		91,
		0,
		0,
		100,
	)
	if valid {
		t.Fatal(
			"invalid latitude was accepted",
		)
	}
}
