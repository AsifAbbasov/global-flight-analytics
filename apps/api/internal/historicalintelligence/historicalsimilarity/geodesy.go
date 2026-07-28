package historicalsimilarity

import (
	"math"
	"sort"
	"time"
)

type resampleResult struct {
	points                []geoPoint
	usedIndexFallback     bool
	usedAntipodalFallback bool
}

func resample(
	points []geoPoint,
	sampleCount int,
	pathLengthKM float64,
) resampleResult {
	if pathLengthKM <= coordinateEpsilon {
		samples, fallback :=
			resampleByIndex(
				points,
				sampleCount,
			)
		return resampleResult{
			points:                samples,
			usedIndexFallback:     true,
			usedAntipodalFallback: fallback,
		}
	}

	cumulative := make(
		[]float64,
		len(points),
	)
	for index := 1; index < len(points); index++ {
		cumulative[index] =
			cumulative[index-1] +
				haversineKM(
					points[index-1],
					points[index],
				)
	}

	result := make(
		[]geoPoint,
		0,
		sampleCount,
	)
	usedAntipodalFallback := false
	for sampleIndex := 0; sampleIndex < sampleCount; sampleIndex++ {
		target := pathLengthKM *
			float64(sampleIndex) /
			float64(sampleCount-1)
		segment := sort.Search(
			len(cumulative),
			func(index int) bool {
				return cumulative[index] >= target
			},
		)

		switch {
		case segment <= 0:
			result = append(
				result,
				points[0],
			)
		case segment >= len(points):
			result = append(
				result,
				points[len(points)-1],
			)
		default:
			startDistance :=
				cumulative[segment-1]
			endDistance :=
				cumulative[segment]
			fraction := 0.0
			if endDistance > startDistance {
				fraction =
					(target - startDistance) /
						(endDistance -
							startDistance)
			}

			point, fallback :=
				interpolateGreatCircle(
					points[segment-1],
					points[segment],
					fraction,
				)
			usedAntipodalFallback =
				usedAntipodalFallback ||
					fallback
			result = append(
				result,
				point,
			)
		}
	}

	return resampleResult{
		points:                result,
		usedAntipodalFallback: usedAntipodalFallback,
	}
}

func resampleByIndex(
	points []geoPoint,
	sampleCount int,
) ([]geoPoint, bool) {
	result := make(
		[]geoPoint,
		0,
		sampleCount,
	)
	maxIndex := float64(len(points) - 1)
	usedAntipodalFallback := false

	for sampleIndex := 0; sampleIndex < sampleCount; sampleIndex++ {
		position := maxIndex *
			float64(sampleIndex) /
			float64(sampleCount-1)
		left := int(math.Floor(position))
		right := int(math.Ceil(position))
		if right >= len(points) {
			right = len(points) - 1
		}

		point, fallback :=
			interpolateGreatCircle(
				points[left],
				points[right],
				position-float64(left),
			)
		usedAntipodalFallback =
			usedAntipodalFallback ||
				fallback
		result = append(result, point)
	}

	return result, usedAntipodalFallback
}

func interpolateGreatCircle(
	left geoPoint,
	right geoPoint,
	fraction float64,
) (geoPoint, bool) {
	fraction = clampRatio(fraction)
	if fraction == 0 {
		return left, false
	}
	if fraction == 1 {
		return right, false
	}

	leftVector := vectorFromPoint(left)
	rightVector := vectorFromPoint(right)
	dot := clamp(
		leftVector.x*rightVector.x+
			leftVector.y*rightVector.y+
			leftVector.z*rightVector.z,
		-1,
		1,
	)
	angle := math.Acos(dot)
	sineAngle := math.Sin(angle)

	if math.Abs(sineAngle) <= coordinateEpsilon {
		return interpolateCoordinateFallback(
			left,
			right,
			fraction,
		), angle > coordinateEpsilon
	}

	leftWeight := math.Sin(
		(1-fraction)*angle,
	) / sineAngle
	rightWeight := math.Sin(
		fraction*angle,
	) / sineAngle

	vector := cartesianVector{
		x: leftWeight*leftVector.x +
			rightWeight*rightVector.x,
		y: leftWeight*leftVector.y +
			rightWeight*rightVector.y,
		z: leftWeight*leftVector.z +
			rightWeight*rightVector.z,
	}
	norm := math.Sqrt(
		vector.x*vector.x +
			vector.y*vector.y +
			vector.z*vector.z,
	)
	if norm <= coordinateEpsilon {
		return interpolateCoordinateFallback(
			left,
			right,
			fraction,
		), true
	}
	vector.x /= norm
	vector.y /= norm
	vector.z /= norm

	return geoPoint{
		latitude: degrees(
			math.Atan2(
				vector.z,
				math.Sqrt(
					vector.x*vector.x+
						vector.y*vector.y,
				),
			),
		),
		longitude: normalizeLongitude(
			degrees(
				math.Atan2(
					vector.y,
					vector.x,
				),
			),
		),
		observedAt: interpolateTime(
			left.observedAt,
			right.observedAt,
			fraction,
		),
	}, false
}

func interpolateCoordinateFallback(
	left geoPoint,
	right geoPoint,
	fraction float64,
) geoPoint {
	longitudeDelta := right.longitude -
		left.longitude
	if longitudeDelta > 180 {
		longitudeDelta -= 360
	} else if longitudeDelta < -180 {
		longitudeDelta += 360
	}

	return geoPoint{
		latitude: left.latitude +
			(right.latitude-left.latitude)*
				fraction,
		longitude: normalizeLongitude(
			left.longitude +
				longitudeDelta*fraction,
		),
		observedAt: interpolateTime(
			left.observedAt,
			right.observedAt,
			fraction,
		),
	}
}

func interpolateTime(
	left time.Time,
	right time.Time,
	fraction float64,
) time.Time {
	duration := right.Sub(left)
	return left.Add(
		time.Duration(
			float64(duration) * fraction,
		),
	)
}

type cartesianVector struct {
	x float64
	y float64
	z float64
}

func vectorFromPoint(
	point geoPoint,
) cartesianVector {
	latitude := radians(point.latitude)
	longitude := radians(point.longitude)
	cosineLatitude := math.Cos(latitude)

	return cartesianVector{
		x: cosineLatitude *
			math.Cos(longitude),
		y: cosineLatitude *
			math.Sin(longitude),
		z: math.Sin(latitude),
	}
}

func sampleDistances(
	left []geoPoint,
	right []geoPoint,
) (float64, float64) {
	total := 0.0
	compensation := 0.0
	maximum := 0.0

	for index := range left {
		distance := haversineKM(
			left[index],
			right[index],
		)
		corrected := distance - compensation
		next := total + corrected
		compensation = (next - total) -
			corrected
		total = next

		if distance > maximum {
			maximum = distance
		}
	}

	return total / float64(len(left)),
		maximum
}

func trajectoryLengthKM(
	points []geoPoint,
) float64 {
	total := 0.0
	compensation := 0.0
	for index := 1; index < len(points); index++ {
		distance := haversineKM(
			points[index-1],
			points[index],
		)
		corrected := distance - compensation
		next := total + corrected
		compensation = (next - total) -
			corrected
		total = next
	}
	return total
}

func haversineKM(
	left geoPoint,
	right geoPoint,
) float64 {
	leftLatitude := radians(left.latitude)
	rightLatitude := radians(right.latitude)
	latitudeDelta := rightLatitude -
		leftLatitude
	longitudeDelta := radians(
		right.longitude - left.longitude,
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
	value = clamp(value, 0, 1)

	return earthRadiusKM *
		2 *
		math.Atan2(
			math.Sqrt(value),
			math.Sqrt(1-value),
		)
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

func normalizeLongitude(
	value float64,
) float64 {
	for value > 180 {
		value -= 360
	}
	for value < -180 {
		value += 360
	}
	return value
}

func radians(value float64) float64 {
	return value * math.Pi / 180
}

func degrees(value float64) float64 {
	return value * 180 / math.Pi
}

func clamp(
	value float64,
	minimum float64,
	maximum float64,
) float64 {
	switch {
	case value < minimum:
		return minimum
	case value > maximum:
		return maximum
	default:
		return value
	}
}

func clampRatio(value float64) float64 {
	return clamp(value, 0, 1)
}
