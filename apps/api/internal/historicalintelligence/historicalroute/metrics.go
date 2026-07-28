package historicalroute

import (
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

type compensatedSum struct {
	total        float64
	compensation float64
}

func (sum *compensatedSum) Add(value float64) {
	corrected := value - sum.compensation
	next := sum.total + corrected
	sum.compensation = (next - sum.total) - corrected
	sum.total = next
}

type bucketAccumulator struct {
	routeKeys        map[string]struct{}
	observationCount int
	confidenceTotal  compensatedSum
	completeCount    int
	partialCount     int
	unavailableCount int
	distanceTotal    compensatedSum
	distanceCount    int
}

type metricBucketValue struct {
	value       float64
	sampleCount int
}

type metricCalculator func(bucketAccumulator) metricBucketValue

var routeMetricCalculators = map[historicalcontract.MetricName]metricCalculator{
	historicalcontract.MetricNameActiveRoutes:          activeRoutePairsValue,
	historicalcontract.MetricNameRouteObservations:     routeObservationsValue,
	historicalcontract.MetricNameRouteConfidence:       routeConfidenceValue,
	historicalcontract.MetricNameCompleteRouteRatio:    completeRouteRatioValue,
	historicalcontract.MetricNamePartialRouteRatio:     partialRouteRatioValue,
	historicalcontract.MetricNameUnavailableRouteRatio: unavailableRouteRatioValue,
	historicalcontract.MetricNameGreatCircleDistanceKM: greatCircleDistanceValue,
}

func routeValues(
	buckets []historicalwindow.Bucket,
	routes []routeEvidence,
	metricName historicalcontract.MetricName,
) ([]historicalseries.BucketValue, error) {
	calculator, exists := routeMetricCalculators[metricName]
	if !exists {
		return nil, ErrMetricUnsupported
	}

	values := make([]historicalseries.BucketValue, len(buckets))
	accumulators := make([]bucketAccumulator, len(buckets))
	for index, bucket := range buckets {
		values[index].Bucket = bucket
		accumulators[index].routeKeys = make(map[string]struct{})
	}

	for _, route := range routes {
		index := routeBucketIndex(buckets, route.result.Window.StartTime)
		if index < 0 {
			continue
		}
		accumulators[index].Add(route)
	}

	for index, accumulator := range accumulators {
		calculated := calculator(accumulator)
		values[index].Value = calculated.value
		values[index].SampleCount = calculated.sampleCount
	}
	return values, nil
}

func (accumulator *bucketAccumulator) Add(route routeEvidence) {
	accumulator.observationCount++
	accumulator.confidenceTotal.Add(route.result.Confidence.Score)

	switch route.result.Status {
	case routecontract.RouteStatusComplete:
		accumulator.completeCount++
		accumulator.routeKeys[route.origin+"-"+route.destination] = struct{}{}
		accumulator.distanceTotal.Add(route.distanceKM)
		accumulator.distanceCount++
	case routecontract.RouteStatusPartial:
		accumulator.partialCount++
	case routecontract.RouteStatusUnavailable:
		accumulator.unavailableCount++
	}
}

func activeRoutePairsValue(accumulator bucketAccumulator) metricBucketValue {
	return metricBucketValue{
		value:       float64(len(accumulator.routeKeys)),
		sampleCount: accumulator.observationCount,
	}
}

func routeObservationsValue(accumulator bucketAccumulator) metricBucketValue {
	return metricBucketValue{
		value:       float64(accumulator.observationCount),
		sampleCount: accumulator.observationCount,
	}
}

func routeConfidenceValue(accumulator bucketAccumulator) metricBucketValue {
	value := 0.0
	if accumulator.observationCount > 0 {
		value = accumulator.confidenceTotal.total /
			float64(accumulator.observationCount)
	}
	return metricBucketValue{
		value:       value,
		sampleCount: accumulator.observationCount,
	}
}

func completeRouteRatioValue(accumulator bucketAccumulator) metricBucketValue {
	return ratioMetricValue(accumulator.completeCount, accumulator.observationCount)
}

func partialRouteRatioValue(accumulator bucketAccumulator) metricBucketValue {
	return ratioMetricValue(accumulator.partialCount, accumulator.observationCount)
}

func unavailableRouteRatioValue(accumulator bucketAccumulator) metricBucketValue {
	return ratioMetricValue(accumulator.unavailableCount, accumulator.observationCount)
}

func ratioMetricValue(count int, total int) metricBucketValue {
	value := 0.0
	if total > 0 {
		value = float64(count) / float64(total)
	}
	return metricBucketValue{value: value, sampleCount: total}
}

func greatCircleDistanceValue(accumulator bucketAccumulator) metricBucketValue {
	value := 0.0
	if accumulator.distanceCount > 0 {
		value = accumulator.distanceTotal.total /
			float64(accumulator.distanceCount)
	}
	return metricBucketValue{
		value:       value,
		sampleCount: accumulator.distanceCount,
	}
}

func routeBucketIndex(
	buckets []historicalwindow.Bucket,
	value time.Time,
) int {
	if value.IsZero() {
		return -1
	}
	normalized := value.UTC()
	index := sort.Search(len(buckets), func(index int) bool {
		return buckets[index].EndTime.After(normalized)
	})
	if index >= len(buckets) || !buckets[index].Contains(normalized) {
		return -1
	}
	return index
}
