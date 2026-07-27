package geographicalbuilder

import (
	"context"
	"math"
	"sort"
)

type coordinateEdge struct {
	start coordinate
	end   coordinate
}

func consecutiveEdges(coordinates []coordinate) []coordinateEdge {
	if len(coordinates) < 2 {
		return nil
	}
	edges := make([]coordinateEdge, 0, len(coordinates)-1)
	for index := 1; index < len(coordinates); index++ {
		edges = append(edges, coordinateEdge{
			start: coordinates[index-1],
			end:   coordinates[index],
		})
	}
	return edges
}

func latitudeBounds(
	coordinates []coordinate,
) (float64, float64) {
	minimum, maximum, _ := latitudeBoundsContext(
		context.Background(),
		coordinates,
	)
	return minimum, maximum
}

func latitudeBoundsContext(
	ctx context.Context,
	coordinates []coordinate,
) (float64, float64, error) {
	minimum := coordinates[0].latitude
	maximum := coordinates[0].latitude

	for index, value := range coordinates[1:] {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, 0, err
			}
		}
		if value.latitude < minimum {
			minimum = value.latitude
		}
		if value.latitude > maximum {
			maximum = value.latitude
		}
	}

	return minimum, maximum, nil
}

func circularLongitudeBounds(
	coordinates []coordinate,
) (float64, float64, float64) {
	minimum, maximum, span, _ := circularLongitudeBoundsContext(
		context.Background(),
		coordinates,
	)
	return minimum, maximum, span
}

func circularLongitudeBoundsContext(
	ctx context.Context,
	coordinates []coordinate,
) (float64, float64, float64, error) {
	values := make([]float64, 0, len(coordinates))
	for index, value := range coordinates {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, 0, 0, err
			}
		}
		longitude := value.longitude
		if longitude < 0 {
			longitude += 360
		}
		values = append(values, longitude)
	}
	sort.Float64s(values)
	if err := ctx.Err(); err != nil {
		return 0, 0, 0, err
	}

	if len(values) == 1 {
		longitude := normalizeLongitude(values[0])
		return longitude, longitude, 0, nil
	}

	largestGap := -1.0
	largestGapIndex := 0
	for index, current := range values {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, 0, 0, err
			}
		}
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
		span,
		nil
}

func haversineDistanceKM(
	left coordinate,
	right coordinate,
) float64 {
	leftLatitude := degreesToRadians(left.latitude)
	rightLatitude := degreesToRadians(right.latitude)
	latitudeDifference := rightLatitude - leftLatitude
	longitudeDifference := degreesToRadians(
		shortestLongitudeDelta(left.longitude, right.longitude),
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

func observedPathDistanceKM(coordinates []coordinate) float64 {
	distance, _ := observedEdgeDistanceKMContext(
		context.Background(),
		consecutiveEdges(coordinates),
	)
	return distance
}

func observedEdgeDistanceKMContext(
	ctx context.Context,
	edges []coordinateEdge,
) (float64, error) {
	// Kahan compensated summation bounds accumulation error without rounding
	// the analytical binary64 result or changing the Haversine distance model.
	total := 0.0
	compensation := 0.0
	for index, edge := range edges {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
		}
		distance := haversineDistanceKM(edge.start, edge.end)
		adjusted := distance - compensation
		next := total + adjusted
		compensation = (next - total) - adjusted
		total = next
	}
	return total, nil
}

func maximumDisplacementKM(
	coordinates []coordinate,
) float64 {
	maximum, _ := maximumDisplacementKMContext(
		context.Background(),
		coordinates,
	)
	return maximum
}

func maximumDisplacementKMContext(
	ctx context.Context,
	coordinates []coordinate,
) (float64, error) {
	if len(coordinates) < 2 {
		return 0, nil
	}

	start := coordinates[0]
	maximum := 0.0
	for index, value := range coordinates[1:] {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
		}
		distance := haversineDistanceKM(start, value)
		if distance > maximum {
			maximum = distance
		}
	}
	return maximum, nil
}

func pathCrossesAntimeridian(coordinates []coordinate) bool {
	crosses, _ := edgeSetCrossesAntimeridianContext(
		context.Background(),
		consecutiveEdges(coordinates),
	)
	return crosses
}

func edgeSetCrossesAntimeridianContext(
	ctx context.Context,
	edges []coordinateEdge,
) (bool, error) {
	for index, edge := range edges {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return false, err
			}
		}
		difference := math.Abs(edge.end.longitude - edge.start.longitude)
		if difference > 180 {
			return true, nil
		}
	}
	return false, nil
}

func uniqueGeographicCellCount(
	coordinates []coordinate,
	precision int,
) int {
	count, _ := uniqueGeographicCellCountContext(
		context.Background(),
		coordinates,
		precision,
	)
	return count
}

func uniqueGeographicCellCountContext(
	ctx context.Context,
	coordinates []coordinate,
	precision int,
) (int, error) {
	cells := make(map[string]struct{}, len(coordinates))
	for index, value := range coordinates {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
		}
		cells[value.cellKey(precision)] = struct{}{}
	}
	return len(cells), nil
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

func shortestLongitudeDelta(left float64, right float64) float64 {
	return normalizeLongitude(right - left)
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}
