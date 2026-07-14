package routeresolver

import "math"

const meanEarthRadiusKM = 6371.0088

func greatCircleDistanceKM(
	firstLatitude float64,
	firstLongitude float64,
	secondLatitude float64,
	secondLongitude float64,
) float64 {
	firstLatitudeRadians := degreesToRadians(
		firstLatitude,
	)
	secondLatitudeRadians := degreesToRadians(
		secondLatitude,
	)
	latitudeDelta := degreesToRadians(
		secondLatitude - firstLatitude,
	)
	longitudeDelta := degreesToRadians(
		secondLongitude - firstLongitude,
	)

	value := math.Sin(latitudeDelta/2)*
		math.Sin(latitudeDelta/2) +
		math.Cos(firstLatitudeRadians)*
			math.Cos(secondLatitudeRadians)*
			math.Sin(longitudeDelta/2)*
			math.Sin(longitudeDelta/2)
	value = clampUnit(value)

	return meanEarthRadiusKM * 2 * math.Atan2(
		math.Sqrt(value),
		math.Sqrt(1-value),
	)
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func clampUnit(value float64) float64 {
	switch {
	case math.IsNaN(value),
		math.IsInf(value, 0),
		value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

func finiteRatio(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0 &&
		value <= 1
}
