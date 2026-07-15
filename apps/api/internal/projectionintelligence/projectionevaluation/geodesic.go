package projectionevaluation

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

func initialBearingDegrees(
	leftLatitude float64,
	leftLongitude float64,
	rightLatitude float64,
	rightLongitude float64,
) float64 {
	leftLatitudeRadians :=
		degreesToRadians(leftLatitude)
	rightLatitudeRadians :=
		degreesToRadians(rightLatitude)
	longitudeDelta := degreesToRadians(
		normalizeLongitude(
			rightLongitude -
				leftLongitude,
		),
	)

	y := math.Sin(longitudeDelta) *
		math.Cos(rightLatitudeRadians)
	x := math.Cos(leftLatitudeRadians)*
		math.Sin(rightLatitudeRadians) -
		math.Sin(leftLatitudeRadians)*
			math.Cos(rightLatitudeRadians)*
			math.Cos(longitudeDelta)

	return normalizeHeading(
		radiansToDegrees(
			math.Atan2(y, x),
		),
	)
}

func destinationPoint(
	latitudeDegrees float64,
	longitudeDegrees float64,
	bearingDegrees float64,
	distanceM float64,
) (float64, float64, bool) {
	if !validLatitude(latitudeDegrees) ||
		!validLongitude(longitudeDegrees) ||
		!finite(bearingDegrees) ||
		!nonNegativeFinite(distanceM) {
		return 0, 0, false
	}
	if distanceM == 0 {
		return latitudeDegrees,
			normalizeLongitude(longitudeDegrees),
			true
	}

	latitudeRadians :=
		degreesToRadians(latitudeDegrees)
	longitudeRadians :=
		degreesToRadians(longitudeDegrees)
	bearingRadians := degreesToRadians(
		normalizeHeading(bearingDegrees),
	)
	angularDistance :=
		distanceM / meanEarthRadiusM

	projectedLatitude := math.Asin(
		math.Sin(latitudeRadians)*
			math.Cos(angularDistance) +
			math.Cos(latitudeRadians)*
				math.Sin(angularDistance)*
				math.Cos(bearingRadians),
	)
	projectedLongitude := longitudeRadians +
		math.Atan2(
			math.Sin(bearingRadians)*
				math.Sin(angularDistance)*
				math.Cos(latitudeRadians),
			math.Cos(angularDistance)-
				math.Sin(latitudeRadians)*
					math.Sin(projectedLatitude),
		)

	latitude :=
		radiansToDegrees(projectedLatitude)
	longitude := normalizeLongitude(
		radiansToDegrees(projectedLongitude),
	)
	if !validLatitude(latitude) ||
		!validLongitude(longitude) {
		return 0, 0, false
	}

	return latitude, longitude, true
}

func normalizeHeading(value float64) float64 {
	normalized := math.Mod(value, 360)
	if normalized < 0 {
		normalized += 360
	}

	return normalized
}

func normalizeLongitude(value float64) float64 {
	normalized := math.Mod(value+180, 360)
	if normalized < 0 {
		normalized += 360
	}

	return normalized - 180
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func radiansToDegrees(value float64) float64 {
	return value * 180 / math.Pi
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
