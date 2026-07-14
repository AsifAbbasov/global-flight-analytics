package geographicalbuilder

import (
	"math"
	"sort"
)

func latitudeBounds(
	coordinates []coordinate,
) (float64, float64) {
	minimum := coordinates[0].latitude
	maximum := coordinates[0].latitude

	for _, value := range coordinates[1:] {
		if value.latitude < minimum {
			minimum = value.latitude
		}
		if value.latitude > maximum {
			maximum = value.latitude
		}
	}

	return minimum, maximum
}

func circularLongitudeBounds(
	coordinates []coordinate,
) (float64, float64, float64) {
	values := make([]float64, 0, len(coordinates))
	for _, value := range coordinates {
		longitude := value.longitude
		if longitude < 0 {
			longitude += 360
		}
		values = append(values, longitude)
	}
	sort.Float64s(values)

	if len(values) == 1 {
		longitude := normalizeLongitude(values[0])
		return longitude, longitude, 0
	}

	largestGap := -1.0
	largestGapIndex := 0

	for index, current := range values {
		nextIndex := (index + 1) % len(values)
		next := values[nextIndex]
		if nextIndex == 0 {
			next += 360
		}

		gap := next - current
		if gap > largestGap {
			largestGap = gap
			largestGapIndex = index
		}
	}

	startIndex := (largestGapIndex + 1) % len(values)
	intervalStart := values[startIndex]
	intervalEnd := values[largestGapIndex]
	span := 360 - largestGap

	if math.Abs(span) < 1e-12 {
		span = 0
	}

	return normalizeLongitude(intervalStart),
		normalizeLongitude(intervalEnd),
		span
}

func haversineDistanceKM(
	left coordinate,
	right coordinate,
) float64 {
	leftLatitude := degreesToRadians(left.latitude)
	rightLatitude := degreesToRadians(right.latitude)
	latitudeDifference := rightLatitude - leftLatitude
	longitudeDifference := degreesToRadians(
		shortestLongitudeDelta(
			left.longitude,
			right.longitude,
		),
	)

	sineLatitude := math.Sin(latitudeDifference / 2)
	sineLongitude := math.Sin(longitudeDifference / 2)
	a := sineLatitude*sineLatitude +
		math.Cos(leftLatitude)*
			math.Cos(rightLatitude)*
			sineLongitude*sineLongitude

	a = math.Min(1, math.Max(0, a))

	return earthMeanRadiusKM *
		2 *
		math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func observedPathDistanceKM(
	coordinates []coordinate,
) float64 {
	if len(coordinates) < 2 {
		return 0
	}

	total := 0.0
	for index := 1; index < len(coordinates); index++ {
		total += haversineDistanceKM(
			coordinates[index-1],
			coordinates[index],
		)
	}

	return total
}

func maximumDisplacementKM(
	coordinates []coordinate,
) float64 {
	if len(coordinates) < 2 {
		return 0
	}

	start := coordinates[0]
	maximum := 0.0

	for _, value := range coordinates[1:] {
		distance := haversineDistanceKM(start, value)
		if distance > maximum {
			maximum = distance
		}
	}

	return maximum
}

func pathCrossesAntimeridian(
	coordinates []coordinate,
) bool {
	for index := 1; index < len(coordinates); index++ {
		difference := math.Abs(
			coordinates[index].longitude -
				coordinates[index-1].longitude,
		)
		if difference > 180 {
			return true
		}
	}

	return false
}

func uniqueGeographicCellCount(
	coordinates []coordinate,
	precision int,
) int {
	cells := make(map[string]struct{}, len(coordinates))

	for _, value := range coordinates {
		cells[value.cellKey(precision)] = struct{}{}
	}

	return len(cells)
}

func normalizeLongitude(value float64) float64 {
	normalized := math.Mod(value+180, 360)
	if normalized < 0 {
		normalized += 360
	}
	normalized -= 180

	if normalized == 0 {
		return 0
	}

	return normalized
}

func shortestLongitudeDelta(
	left float64,
	right float64,
) float64 {
	return normalizeLongitude(right - left)
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}
