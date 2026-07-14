package airportresolver

import (
	"math"
	"testing"
)

func TestHaversineDistanceKnownRoute(
	t *testing.T,
) {
	distanceKM := haversineDistanceKM(
		Point{
			Latitude:  40.4675,
			Longitude: 50.0467,
		},
		Point{
			Latitude:  41.6692,
			Longitude: 44.9547,
		},
	)

	if math.Abs(distanceKM-448.8) > 2 {
		t.Fatalf(
			"distance = %.6f kilometres",
			distanceKM,
		)
	}
}

func TestHaversineDistanceAcrossAntimeridian(
	t *testing.T,
) {
	distanceKM := haversineDistanceKM(
		Point{
			Latitude:  10,
			Longitude: 179.9,
		},
		Point{
			Latitude:  10,
			Longitude: -179.9,
		},
	)

	if distanceKM <= 0 || distanceKM > 25 {
		t.Fatalf(
			"antimeridian distance = %.6f kilometres",
			distanceKM,
		)
	}
}

func TestHaversineDistanceIsSymmetric(
	t *testing.T,
) {
	first := Point{
		Latitude:  40,
		Longitude: 50,
	}
	second := Point{
		Latitude:  41,
		Longitude: 45,
	}

	left := haversineDistanceKM(first, second)
	right := haversineDistanceKM(second, first)

	if math.Abs(left-right) > 1e-9 {
		t.Fatalf(
			"distance is not symmetric: %v %v",
			left,
			right,
		)
	}
}
