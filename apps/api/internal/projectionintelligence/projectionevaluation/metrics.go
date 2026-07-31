package projectionevaluation

import (
	"math"
	"sort"
	"time"
)

func buildPositionMetrics(
	forecastPointCount int,
	points []PointEvaluation,
) PositionMetrics {
	horizontalErrors := make([]float64, 0, len(points))
	horizontalRatios := make([]float64, 0, len(points))
	altitudeErrors := make([]float64, 0, len(points))
	altitudeRatios := make([]float64, 0, len(points))

	horizontalCovered := 0
	verticalCovered := 0
	verticalEvaluated := 0

	for _, point := range points {
		horizontalErrors = append(horizontalErrors, point.HorizontalErrorM)
		horizontalRatios = append(horizontalRatios, point.HorizontalErrorRatio)
		if point.WithinHorizontalUncertainty {
			horizontalCovered++
		}
		if point.AltitudeAbsoluteErrorM != nil {
			altitudeErrors = append(altitudeErrors, *point.AltitudeAbsoluteErrorM)
			verticalEvaluated++
		}
		if point.AltitudeErrorRatio != nil {
			altitudeRatios = append(altitudeRatios, *point.AltitudeErrorRatio)
		}
		if point.WithinVerticalUncertainty != nil && *point.WithinVerticalUncertainty {
			verticalCovered++
		}
	}

	metrics := PositionMetrics{
		ForecastPointCount:          forecastPointCount,
		EvaluatedPointCount:         len(points),
		MissingActualPointCount:     forecastPointCount - len(points),
		AltitudeEvaluatedPointCount: verticalEvaluated,
	}
	if forecastPointCount > 0 {
		metrics.CoverageRatio = float64(len(points)) / float64(forecastPointCount)
	}
	if len(horizontalErrors) > 0 {
		metrics.MeanHorizontalErrorM = mean(horizontalErrors)
		metrics.MedianHorizontalErrorM = median(horizontalErrors)
		metrics.P95HorizontalErrorM = percentileNearestRank(horizontalErrors, 0.95)
		metrics.MaximumHorizontalErrorM = maximum(horizontalErrors)
		metrics.HorizontalRMSEM = rootMeanSquare(horizontalErrors)
		metrics.MeanHorizontalErrorRatio = mean(horizontalRatios)
		metrics.HorizontalUncertaintyCoverageRatio = float64(horizontalCovered) / float64(len(horizontalErrors))
	}
	if len(altitudeErrors) > 0 {
		metrics.MeanAltitudeAbsoluteErrorM = mean(altitudeErrors)
		metrics.MedianAltitudeAbsoluteErrorM = median(altitudeErrors)
		metrics.P95AltitudeAbsoluteErrorM = percentileNearestRank(altitudeErrors, 0.95)
		metrics.MaximumAltitudeAbsoluteErrorM = maximum(altitudeErrors)
		metrics.AltitudeRMSEM = rootMeanSquare(altitudeErrors)
	}
	if len(altitudeRatios) > 0 {
		metrics.MeanAltitudeErrorRatio = mean(altitudeRatios)
		metrics.VerticalUncertaintyCoverageRatio = float64(verticalCovered) / float64(len(altitudeRatios))
	}
	return metrics
}

func buildEndpointMetrics(
	horizonEnd time.Time,
	points []PointEvaluation,
) EndpointMetrics {
	for index := len(points) - 1; index >= 0; index-- {
		point := points[index]
		if !point.ForecastTime.Equal(horizonEnd.UTC()) {
			continue
		}
		metrics := EndpointMetrics{
			Available:                   true,
			ForecastTime:                point.ForecastTime,
			HorizontalErrorM:            point.HorizontalErrorM,
			HorizontalErrorRatio:        point.HorizontalErrorRatio,
			WithinHorizontalUncertainty: point.WithinHorizontalUncertainty,
		}
		if point.AltitudeAbsoluteErrorM != nil && point.AltitudeErrorRatio != nil {
			metrics.AltitudeAvailable = true
			metrics.AltitudeAbsoluteErrorM = *point.AltitudeAbsoluteErrorM
			metrics.AltitudeErrorRatio = *point.AltitudeErrorRatio
			if point.WithinVerticalUncertainty != nil {
				metrics.WithinVerticalUncertainty = *point.WithinVerticalUncertainty
			}
		}
		return metrics
	}
	return EndpointMetrics{}
}

func buildConfidenceMetrics(points []PointEvaluation) ConfidenceMetrics {
	if len(points) == 0 {
		return ConfidenceMetrics{}
	}
	forecast := make([]float64, 0, len(points))
	accuracy := make([]float64, 0, len(points))
	gaps := make([]float64, 0, len(points))
	for _, point := range points {
		forecast = append(forecast, point.ForecastConfidence.Score)
		accuracy = append(accuracy, point.NormalizedHorizontalAccuracy)
		gaps = append(gaps, point.AbsoluteConfidenceGap)
	}
	return ConfidenceMetrics{
		EvaluatedPointCount:              len(points),
		MeanForecastConfidence:           mean(forecast),
		MeanNormalizedHorizontalAccuracy: mean(accuracy),
		MeanAbsoluteConfidenceGap:        mean(gaps),
		ConfidenceCalibrationRMSE:        rootMeanSquare(gaps),
	}
}

func buildLeadTimeMetrics(
	points []PointEvaluation,
	bucketSize time.Duration,
) []LeadTimeMetrics {
	if len(points) == 0 || bucketSize <= 0 {
		return nil
	}
	type accumulator struct {
		start  time.Duration
		end    time.Duration
		points []PointEvaluation
	}
	byStart := make(map[time.Duration]*accumulator)
	for _, point := range points {
		start := (point.LeadTime / bucketSize) * bucketSize
		item := byStart[start]
		if item == nil {
			item = &accumulator{start: start, end: start + bucketSize}
			byStart[start] = item
		}
		item.points = append(item.points, point)
	}
	starts := make([]time.Duration, 0, len(byStart))
	for start := range byStart {
		starts = append(starts, start)
	}
	sort.Slice(starts, func(left, right int) bool { return starts[left] < starts[right] })
	result := make([]LeadTimeMetrics, 0, len(starts))
	for _, start := range starts {
		item := byStart[start]
		errors := make([]float64, 0, len(item.points))
		confidences := make([]float64, 0, len(item.points))
		accuracies := make([]float64, 0, len(item.points))
		gaps := make([]float64, 0, len(item.points))
		covered := 0
		for _, point := range item.points {
			errors = append(errors, point.HorizontalErrorM)
			confidences = append(confidences, point.ForecastConfidence.Score)
			accuracies = append(accuracies, point.NormalizedHorizontalAccuracy)
			gaps = append(gaps, point.AbsoluteConfidenceGap)
			if point.WithinHorizontalUncertainty {
				covered++
			}
		}
		result = append(result, LeadTimeMetrics{
			BucketStart:                        item.start,
			BucketEnd:                          item.end,
			EvaluatedPointCount:                len(item.points),
			MeanHorizontalErrorM:               mean(errors),
			MedianHorizontalErrorM:             median(errors),
			P95HorizontalErrorM:                percentileNearestRank(errors, 0.95),
			HorizontalRMSEM:                    rootMeanSquare(errors),
			HorizontalUncertaintyCoverageRatio: float64(covered) / float64(len(item.points)),
			MeanForecastConfidence:             mean(confidences),
			MeanNormalizedHorizontalAccuracy:   mean(accuracies),
			MeanAbsoluteConfidenceGap:          mean(gaps),
		})
	}
	return result
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

func rootMeanSquare(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value * value
	}
	return math.Sqrt(total / float64(len(values)))
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sortedValues := append([]float64(nil), values...)
	sort.Float64s(sortedValues)
	middle := len(sortedValues) / 2
	if len(sortedValues)%2 == 1 {
		return sortedValues[middle]
	}
	return (sortedValues[middle-1] + sortedValues[middle]) / 2
}

func percentileNearestRank(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sortedValues := append([]float64(nil), values...)
	sort.Float64s(sortedValues)
	rank := int(math.Ceil(percentile * float64(len(sortedValues))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(sortedValues) {
		rank = len(sortedValues)
	}
	return sortedValues[rank-1]
}
