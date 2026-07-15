package projectionbaseline

import "math"

// meanEarthRadiusM is the International Union of Geodesy and Geophysics
// mean Earth radius. The baseline uses a spherical direct-geodesic step;
// advanced ellipsoidal and wind-aware motion remain later-stage work.
const meanEarthRadiusM = 6371008.8

func destinationPoint(
	latitudeDegrees float64,
	longitudeDegrees float64,
	headingDegrees float64,
	distanceM float64,
) (float64, float64, bool) {
	if !finiteLatitude(latitudeDegrees) ||
		!finiteLongitude(longitudeDegrees) ||
		!nonNegativeFinite(distanceM) ||
		!finite(headingDegrees) {
		return 0, 0, false
	}

	if distanceM == 0 {
		return latitudeDegrees,
			normalizeLongitude(
				longitudeDegrees,
			),
			true
	}

	latitudeRadians := degreesToRadians(
		latitudeDegrees,
	)
	longitudeRadians := degreesToRadians(
		longitudeDegrees,
	)
	bearingRadians := degreesToRadians(
		normalizeHeading(
			headingDegrees,
		),
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

	latitude := radiansToDegrees(
		projectedLatitude,
	)
	longitude := normalizeLongitude(
		radiansToDegrees(
			projectedLongitude,
		),
	)

	if !finiteLatitude(latitude) ||
		!finiteLongitude(longitude) {
		return 0, 0, false
	}

	return latitude, longitude, true
}

func normalizeHeading(
	value float64,
) float64 {
	normalized := math.Mod(
		value,
		360,
	)
	if normalized < 0 {
		normalized += 360
	}

	return normalized
}

func normalizeLongitude(
	value float64,
) float64 {
	normalized := math.Mod(
		value+180,
		360,
	)
	if normalized < 0 {
		normalized += 360
	}

	return normalized - 180
}

func degreesToRadians(
	value float64,
) float64 {
	return value * math.Pi / 180
}

func radiansToDegrees(
	value float64,
) float64 {
	return value * 180 / math.Pi
}

func finite(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func finiteLatitude(
	value float64,
) bool {
	return finite(value) &&
		value >= -90 &&
		value <= 90
}

func finiteLongitude(
	value float64,
) bool {
	return finite(value) &&
		value >= -180 &&
		value <= 180
}
