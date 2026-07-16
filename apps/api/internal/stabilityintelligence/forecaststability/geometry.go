package forecaststability

import "math"

const earthRadiusKilometers = 6371.0088

func haversineKilometers(
	leftLatitude float64,
	leftLongitude float64,
	rightLatitude float64,
	rightLongitude float64,
) float64 {
	leftLatitudeRadians := degreesToRadians(leftLatitude)
	rightLatitudeRadians := degreesToRadians(rightLatitude)
	latitudeDifference := degreesToRadians(rightLatitude - leftLatitude)
	longitudeDifference := degreesToRadians(rightLongitude - leftLongitude)

	a := math.Sin(latitudeDifference/2)*math.Sin(latitudeDifference/2) +
		math.Cos(leftLatitudeRadians)*math.Cos(rightLatitudeRadians)*
			math.Sin(longitudeDifference/2)*math.Sin(longitudeDifference/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKilometers * c
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func clampUnit(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 1:
		return 1
	default:
		return value
	}
}

func safeRelativeChange(previous float64, current float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 1
	}
	return math.Abs(current-previous) / math.Abs(previous)
}

func unitInterval(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}

func positiveFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value > 0
}

func nonNegativeFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}
