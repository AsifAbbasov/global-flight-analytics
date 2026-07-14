package routeresolver

import (
	"math"
	"testing"
)

func TestGreatCircleDistanceKnownRoute(
	t *testing.T,
) {
	distanceKM := greatCircleDistanceKM(
		40.4675,
		50.0467,
		41.6692,
		44.9547,
	)

	if math.Abs(distanceKM-448.8) > 2 {
		t.Fatalf(
			"distance = %.6f kilometres",
			distanceKM,
		)
	}
}

func TestGreatCircleDistanceAcrossAntimeridian(
	t *testing.T,
) {
	distanceKM := greatCircleDistanceKM(
		10,
		179.9,
		10,
		-179.9,
	)

	if distanceKM <= 0 || distanceKM > 25 {
		t.Fatalf(
			"antimeridian distance = %.6f kilometres",
			distanceKM,
		)
	}
}

func TestGreatCircleDistanceIsSymmetric(
	t *testing.T,
) {
	left := greatCircleDistanceKM(
		40,
		50,
		41,
		45,
	)
	right := greatCircleDistanceKM(
		41,
		45,
		40,
		50,
	)

	if math.Abs(left-right) > 1e-9 {
		t.Fatalf(
			"distance is not symmetric: %v %v",
			left,
			right,
		)
	}
}
