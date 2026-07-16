package proximityscanner

import "math"

const earthRadiusKilometers = 6371.0088

func horizontalDistanceKilometers(
	leftLatitude float64,
	leftLongitude float64,
	rightLatitude float64,
	rightLongitude float64,
) float64 {
	leftLatitudeRadians := degreesToRadians(leftLatitude)
	rightLatitudeRadians := degreesToRadians(rightLatitude)
	latitudeDifference := rightLatitudeRadians - leftLatitudeRadians
	longitudeDifference := degreesToRadians(rightLongitude - leftLongitude)

	haversine := math.Sin(latitudeDifference/2)*math.Sin(latitudeDifference/2) +
		math.Cos(leftLatitudeRadians)*math.Cos(rightLatitudeRadians)*
			math.Sin(longitudeDifference/2)*math.Sin(longitudeDifference/2)
	centralAngle := 2 * math.Atan2(math.Sqrt(haversine), math.Sqrt(1-haversine))
	return earthRadiusKilometers * centralAngle
}

func closingRateMetersPerSecond(left, right aircraftVector) float64 {
	northMeters, eastMeters := localOffsetMeters(
		left.latitude,
		left.longitude,
		right.latitude,
		right.longitude,
	)
	distanceMeters := math.Hypot(northMeters, eastMeters)
	if distanceMeters <= 1e-9 {
		return 0
	}

	leftNorthVelocity, leftEastVelocity := velocityComponents(
		left.speedMetersPerSecond,
		left.headingDegrees,
	)
	rightNorthVelocity, rightEastVelocity := velocityComponents(
		right.speedMetersPerSecond,
		right.headingDegrees,
	)
	relativeNorthVelocity := rightNorthVelocity - leftNorthVelocity
	relativeEastVelocity := rightEastVelocity - leftEastVelocity
	rangeRate := (northMeters*relativeNorthVelocity + eastMeters*relativeEastVelocity) /
		distanceMeters
	return -rangeRate
}

func localOffsetMeters(
	leftLatitude float64,
	leftLongitude float64,
	rightLatitude float64,
	rightLongitude float64,
) (northMeters float64, eastMeters float64) {
	latitudeDifference := degreesToRadians(rightLatitude - leftLatitude)
	longitudeDifference := degreesToRadians(rightLongitude - leftLongitude)
	meanLatitude := degreesToRadians((leftLatitude + rightLatitude) / 2)
	northMeters = latitudeDifference * earthRadiusKilometers * 1000
	eastMeters = longitudeDifference * math.Cos(meanLatitude) * earthRadiusKilometers * 1000
	return northMeters, eastMeters
}

func velocityComponents(speedMetersPerSecond float64, headingDegrees float64) (
	northMetersPerSecond float64,
	eastMetersPerSecond float64,
) {
	headingRadians := degreesToRadians(headingDegrees)
	return speedMetersPerSecond * math.Cos(headingRadians),
		speedMetersPerSecond * math.Sin(headingRadians)
}

func headingDifferenceDegrees(left float64, right float64) float64 {
	difference := math.Abs(left - right)
	if difference > 180 {
		return 360 - difference
	}
	return difference
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

type aircraftVector struct {
	latitude             float64
	longitude            float64
	speedMetersPerSecond float64
	headingDegrees       float64
}
