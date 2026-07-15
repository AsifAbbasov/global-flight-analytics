package projectionneighbors

import "math"

const earthRadiusKM = 6371.0088

type geoPoint struct {
	latitude  float64
	longitude float64
}

func haversineKM(
	left geoPoint,
	right geoPoint,
) float64 {
	leftLatitude := degreesToRadians(
		left.latitude,
	)
	rightLatitude := degreesToRadians(
		right.latitude,
	)
	latitudeDelta := rightLatitude -
		leftLatitude
	longitudeDelta := degreesToRadians(
		normalizeLongitude(
			right.longitude - left.longitude,
		),
	)

	sineLatitude := math.Sin(
		latitudeDelta / 2,
	)
	sineLongitude := math.Sin(
		longitudeDelta / 2,
	)
	value := sineLatitude*sineLatitude +
		math.Cos(leftLatitude)*
			math.Cos(rightLatitude)*
			sineLongitude*sineLongitude
	value = math.Min(
		1,
		math.Max(0, value),
	)

	return earthRadiusKM *
		2 *
		math.Atan2(
			math.Sqrt(value),
			math.Sqrt(1-value),
		)
}

func degreesToRadians(
	value float64,
) float64 {
	return value * math.Pi / 180
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

func validLatitude(
	value float64,
) bool {
	return finite(value) &&
		value >= -90 &&
		value <= 90
}

func validLongitude(
	value float64,
) bool {
	return finite(value) &&
		value >= -180 &&
		value <= 180
}
