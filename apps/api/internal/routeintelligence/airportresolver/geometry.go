package airportresolver

import "math"

const meanEarthRadiusKM = 6371.0088

func haversineDistanceKM(
	first Point,
	second Point,
) float64 {
	firstLatitudeRadians := degreesToRadians(
		first.Latitude,
	)
	secondLatitudeRadians := degreesToRadians(
		second.Latitude,
	)
	latitudeDelta := degreesToRadians(
		second.Latitude - first.Latitude,
	)
	longitudeDelta := degreesToRadians(
		second.Longitude - first.Longitude,
	)

	value := math.Sin(latitudeDelta/2)*
		math.Sin(latitudeDelta/2) +
		math.Cos(firstLatitudeRadians)*
			math.Cos(secondLatitudeRadians)*
			math.Sin(longitudeDelta/2)*
			math.Sin(longitudeDelta/2)
	value = clamp01(value)

	return meanEarthRadiusKM * 2 * math.Atan2(
		math.Sqrt(value),
		math.Sqrt(1-value),
	)
}

func degreesToRadians(
	value float64,
) float64 {
	return value * math.Pi / 180
}

func clamp01(
	value float64,
) float64 {
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

func validLatitude(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= -90 &&
		value <= 90
}

func validLongitude(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= -180 &&
		value <= 180
}
