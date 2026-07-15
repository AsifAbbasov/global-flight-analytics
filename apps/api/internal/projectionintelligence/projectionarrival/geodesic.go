package projectionarrival

import "math"

const meanEarthRadiusM = 6371008.8

func greatCircleDistanceM(
	leftLatitude float64,
	leftLongitude float64,
	rightLatitude float64,
	rightLongitude float64,
) float64 {
	if !validLatitude(leftLatitude) ||
		!validLongitude(leftLongitude) ||
		!validLatitude(rightLatitude) ||
		!validLongitude(rightLongitude) {
		return math.NaN()
	}

	leftLatitudeRadians :=
		degreesToRadians(leftLatitude)
	rightLatitudeRadians :=
		degreesToRadians(rightLatitude)
	latitudeDelta :=
		rightLatitudeRadians -
			leftLatitudeRadians
	longitudeDelta := degreesToRadians(
		normalizeLongitude(
			rightLongitude -
				leftLongitude,
		),
	)

	sineLatitude := math.Sin(
		latitudeDelta / 2,
	)
	sineLongitude := math.Sin(
		longitudeDelta / 2,
	)
	value := sineLatitude*sineLatitude +
		math.Cos(leftLatitudeRadians)*
			math.Cos(rightLatitudeRadians)*
			sineLongitude*sineLongitude
	value = math.Min(
		1,
		math.Max(0, value),
	)

	return meanEarthRadiusM *
		2 *
		math.Atan2(
			math.Sqrt(value),
			math.Sqrt(1-value),
		)
}

func normalizeLongitude(value float64) float64 {
	normalized := math.Mod(
		value+180,
		360,
	)
	if normalized < 0 {
		normalized += 360
	}

	return normalized - 180
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func validLatitude(value float64) bool {
	return finite(value) &&
		value >= -90 &&
		value <= 90
}

func validLongitude(value float64) bool {
	return finite(value) &&
		value >= -180 &&
		value <= 180
}
