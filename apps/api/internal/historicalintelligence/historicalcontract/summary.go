package historicalcontract

import "sort"

// Summarize aggregates represented partial and complete buckets. Summary
// statistics are temporal bucket-level statistics: Average is the unweighted
// mean of represented bucket values and Median is the median of represented
// bucket values. They do not claim observation-weighted or raw-observation
// semantics. Contract validation owns the invariant that unavailable buckets
// carry a zero payload; unavailable buckets never contribute to statistics.
func Summarize(points []Point) Summary {
	values := make([]float64, 0, len(points))
	for _, point := range points {
		if point.Status == BucketStatusUnavailable {
			continue
		}
		values = append(values, point.Value)
	}
	if len(values) == 0 {
		return Summary{}
	}

	sortedValues := append([]float64(nil), values...)
	sort.Float64s(sortedValues)
	total := 0.0
	compensation := 0.0
	for _, value := range sortedValues {
		corrected := value - compensation
		next := total + corrected
		compensation = (next - total) - corrected
		total = next
	}

	minimum := sortedValues[0]
	maximum := sortedValues[len(sortedValues)-1]
	median := sortedValues[len(sortedValues)/2]
	if len(sortedValues)%2 == 0 {
		middle := len(sortedValues) / 2
		left := sortedValues[middle-1]
		right := sortedValues[middle]
		median = left + (right-left)/2
	}

	return Summary{
		PointCount: len(sortedValues),
		Total:      total,
		Minimum:    minimum,
		Maximum:    maximum,
		Average:    total / float64(len(sortedValues)),
		Median:     median,
	}
}
