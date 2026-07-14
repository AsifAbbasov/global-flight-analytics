package datasetprofiler

import (
	"math"
	"sort"
)

type numericAccumulator struct {
	values       []float64
	invalidCount int
}

func (accumulator *numericAccumulator) add(
	value float64,
	valid bool,
) {
	if !valid ||
		math.IsNaN(value) ||
		math.IsInf(value, 0) {
		accumulator.invalidCount++
		return
	}

	accumulator.values = append(
		accumulator.values,
		value,
	)
}

func (accumulator numericAccumulator) profile() NumericProfile {
	if len(accumulator.values) == 0 {
		return NumericProfile{
			InvalidCount: accumulator.invalidCount,
		}
	}

	values := append(
		[]float64(nil),
		accumulator.values...,
	)
	sort.Float64s(values)

	total := 0.0
	for _, value := range values {
		total += value
	}

	return NumericProfile{
		Count:        len(values),
		InvalidCount: accumulator.invalidCount,
		Minimum:      values[0],
		Maximum:      values[len(values)-1],
		Mean:         total / float64(len(values)),
		Median:       percentile(values, 0.50),
		Percentile95: percentile(values, 0.95),
	}
}

func percentile(
	sortedValues []float64,
	fraction float64,
) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	if len(sortedValues) == 1 {
		return sortedValues[0]
	}
	if fraction <= 0 {
		return sortedValues[0]
	}
	if fraction >= 1 {
		return sortedValues[len(sortedValues)-1]
	}

	position := fraction * float64(len(sortedValues)-1)
	lowerIndex := int(math.Floor(position))
	upperIndex := int(math.Ceil(position))
	if lowerIndex == upperIndex {
		return sortedValues[lowerIndex]
	}

	weight := position - float64(lowerIndex)

	return sortedValues[lowerIndex]*(1-weight) +
		sortedValues[upperIndex]*weight
}

func validRatio(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0) &&
		value >= 0 &&
		value <= 1
}

func validNonNegativeInteger(value int) bool {
	return value >= 0
}
