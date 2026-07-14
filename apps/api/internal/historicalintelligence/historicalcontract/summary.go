package historicalcontract

import "sort"

func Summarize(
	points []Point,
) Summary {
	values := make(
		[]float64,
		0,
		len(points),
	)

	for _, point := range points {
		if point.Status ==
			BucketStatusUnavailable {
			continue
		}

		values = append(values, point.Value)
	}

	if len(values) == 0 {
		return Summary{}
	}

	sortedValues := append(
		[]float64(nil),
		values...,
	)
	sort.Float64s(sortedValues)

	total := 0.0
	minimum := sortedValues[0]
	maximum := sortedValues[len(sortedValues)-1]
	for _, value := range sortedValues {
		total += value
	}

	median := sortedValues[len(sortedValues)/2]
	if len(sortedValues)%2 == 0 {
		middle := len(sortedValues) / 2
		median = (sortedValues[middle-1] +
			sortedValues[middle]) / 2
	}

	return Summary{
		PointCount: len(sortedValues),
		Total:      total,
		Minimum:    minimum,
		Maximum:    maximum,
		Average: total /
			float64(len(sortedValues)),
		Median: median,
	}
}
