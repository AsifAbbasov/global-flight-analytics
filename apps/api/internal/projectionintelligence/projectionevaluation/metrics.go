package projectionevaluation

import (
	"math"
	"sort"
)

func buildPositionMetrics(
	forecastPointCount int,
	points []PointEvaluation,
) PositionMetrics {
	horizontalErrors := make(
		[]float64,
		0,
		len(points),
	)
	horizontalRatios := make(
		[]float64,
		0,
		len(points),
	)
	altitudeErrors := make(
		[]float64,
		0,
		len(points),
	)
	altitudeRatios := make(
		[]float64,
		0,
		len(points),
	)

	horizontalCovered := 0
	verticalCovered := 0
	verticalEvaluated := 0

	for _, point := range points {
		horizontalErrors = append(
			horizontalErrors,
			point.HorizontalErrorM,
		)
		horizontalRatios = append(
			horizontalRatios,
			point.HorizontalErrorRatio,
		)
		if point.WithinHorizontalUncertainty {
			horizontalCovered++
		}

		if point.AltitudeAbsoluteErrorM != nil {
			altitudeErrors = append(
				altitudeErrors,
				*point.AltitudeAbsoluteErrorM,
			)
			verticalEvaluated++
		}
		if point.AltitudeErrorRatio != nil {
			altitudeRatios = append(
				altitudeRatios,
				*point.AltitudeErrorRatio,
			)
		}
		if point.WithinVerticalUncertainty != nil &&
			*point.WithinVerticalUncertainty {
			verticalCovered++
		}
	}

	metrics := PositionMetrics{
		ForecastPointCount:      forecastPointCount,
		EvaluatedPointCount:     len(points),
		MissingActualPointCount: forecastPointCount - len(points),

		AltitudeEvaluatedPointCount: verticalEvaluated,
	}
	if forecastPointCount > 0 {
		metrics.CoverageRatio =
			float64(len(points)) /
				float64(forecastPointCount)
	}
	if len(points) > 0 {
		metrics.MeanHorizontalErrorM =
			mean(horizontalErrors)
		metrics.MedianHorizontalErrorM =
			percentileNearestRank(
				horizontalErrors,
				0.50,
			)
		metrics.P95HorizontalErrorM =
			percentileNearestRank(
				horizontalErrors,
				0.95,
			)
		metrics.MaximumHorizontalErrorM =
			maximum(horizontalErrors)
		metrics.HorizontalRMSEM =
			rootMeanSquare(
				horizontalErrors,
			)
		metrics.MeanHorizontalErrorRatio =
			mean(horizontalRatios)
		metrics.
			HorizontalUncertaintyCoverageRatio =
			float64(horizontalCovered) /
				float64(len(points))
	}
	if len(altitudeErrors) > 0 {
		metrics.
			MeanAltitudeAbsoluteErrorM =
			mean(altitudeErrors)
		metrics.
			MedianAltitudeAbsoluteErrorM =
			percentileNearestRank(
				altitudeErrors,
				0.50,
			)
		metrics.
			P95AltitudeAbsoluteErrorM =
			percentileNearestRank(
				altitudeErrors,
				0.95,
			)
		metrics.
			MaximumAltitudeAbsoluteErrorM =
			maximum(altitudeErrors)
		metrics.AltitudeRMSEM =
			rootMeanSquare(
				altitudeErrors,
			)
	}
	if len(altitudeRatios) > 0 {
		metrics.MeanAltitudeErrorRatio =
			mean(altitudeRatios)
		metrics.
			VerticalUncertaintyCoverageRatio =
			float64(verticalCovered) /
				float64(
					len(altitudeRatios),
				)
	}

	return metrics
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	total := 0.0
	for _, value := range values {
		total += value
	}

	return total / float64(len(values))
}

func maximum(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	result := values[0]
	for _, value := range values[1:] {
		if value > result {
			result = value
		}
	}

	return result
}

func rootMeanSquare(
	values []float64,
) float64 {
	if len(values) == 0 {
		return 0
	}

	total := 0.0
	for _, value := range values {
		total += value * value
	}

	return math.Sqrt(
		total / float64(len(values)),
	)
}

func percentileNearestRank(
	values []float64,
	percentile float64,
) float64 {
	if len(values) == 0 {
		return 0
	}

	sortedValues := append(
		[]float64(nil),
		values...,
	)
	sort.Float64s(sortedValues)

	rank := int(
		math.Ceil(
			percentile *
				float64(len(sortedValues)),
		),
	)
	if rank < 1 {
		rank = 1
	}
	if rank > len(sortedValues) {
		rank = len(sortedValues)
	}

	return sortedValues[rank-1]
}
