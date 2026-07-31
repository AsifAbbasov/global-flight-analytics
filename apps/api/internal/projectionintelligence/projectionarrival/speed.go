package projectionarrival

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func buildPositionEvidence(
	current trajectory.FlightTrajectory,
	projection projectioncontract.Result,
) positionEvidence {
	samples := make(
		[]positionSample,
		0,
		len(projection.Points)+1,
	)

	currentEndpoint, currentEndpointPresent :=
		currentEndpointAt(
			current,
			projection.Horizon.AsOfTime,
		)
	if currentEndpointPresent {
		samples = append(samples, currentEndpoint)
	}

	for _, point := range projection.Points {
		samples = append(
			samples,
			positionSample{
				source:     positionSampleSourceProjection,
				sourceID:   strconv.Itoa(point.Sequence),
				sourceName: strings.TrimSpace(projection.Method.Name),
				sequence:   point.Sequence,
				timeValue:  point.ForecastTime.UTC(),
				latitude:   point.Position.Latitude,
				longitude:  point.Position.Longitude,
				horizontalUncertaintyM: point.Uncertainty.
					HorizontalRadiusM,
			},
		)
	}

	sort.Slice(
		samples,
		func(left int, right int) bool {
			return positionSampleLess(
				samples[left],
				samples[right],
			)
		},
	)

	result := make(
		[]positionSample,
		0,
		len(samples),
	)
	for _, sample := range samples {
		if !validPositionSample(sample) {
			continue
		}
		if len(result) > 0 &&
			sample.timeValue.Equal(
				result[len(result)-1].timeValue,
			) {
			// Canonical sorting places observed current evidence before an
			// estimated projection point. Preserve the stronger source when
			// two inputs claim the same instant.
			continue
		}
		result = append(result, sample)
	}

	return positionEvidence{
		samples:                result,
		currentEndpoint:        currentEndpoint,
		currentEndpointPresent: currentEndpointPresent,
		fingerprint:            positionSamplesFingerprint(result),
	}
}

func validPositionSample(sample positionSample) bool {
	return !sample.timeValue.IsZero() &&
		validLatitude(sample.latitude) &&
		validLongitude(sample.longitude) &&
		nonNegativeFinite(sample.horizontalUncertaintyM)
}

func positionSampleLess(left positionSample, right positionSample) bool {
	leftTime := left.timeValue.UTC()
	rightTime := right.timeValue.UTC()
	if !leftTime.Equal(rightTime) {
		return leftTime.Before(rightTime)
	}
	if left.source != right.source {
		return left.source < right.source
	}
	if left.sourceID != right.sourceID {
		return left.sourceID < right.sourceID
	}
	if left.sourceName != right.sourceName {
		return left.sourceName < right.sourceName
	}
	if left.sequence != right.sequence {
		return left.sequence < right.sequence
	}
	if left.latitude != right.latitude {
		return math.Float64bits(left.latitude) <
			math.Float64bits(right.latitude)
	}
	if left.longitude != right.longitude {
		return math.Float64bits(left.longitude) <
			math.Float64bits(right.longitude)
	}
	return math.Float64bits(left.horizontalUncertaintyM) <
		math.Float64bits(right.horizontalUncertaintyM)
}

func currentEndpointAt(
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
) (positionSample, bool) {
	candidates := make(
		[]trajectory.TrackPoint4D,
		0,
		len(item.Points),
	)
	for _, point := range item.Points {
		if point.ObservedAt.IsZero() ||
			point.ObservedAt.UTC().After(asOfTime.UTC()) ||
			!validLatitude(point.Latitude) ||
			!validLongitude(point.Longitude) {
			continue
		}
		candidates = append(candidates, point)
	}

	if len(candidates) == 0 {
		return positionSample{}, false
	}

	sort.Slice(
		candidates,
		func(left int, right int) bool {
			return trajectoryPointLess(
				candidates[left],
				candidates[right],
			)
		},
	)

	point := candidates[len(candidates)-1]
	sourceName := strings.TrimSpace(point.SourceName)
	if sourceName == "" {
		sourceName = strings.TrimSpace(item.SourceName)
	}
	if sourceName == "" {
		sourceName = "trajectory_observation"
	}

	return positionSample{
		source:                 positionSampleSourceCurrent,
		sourceID:               strings.TrimSpace(point.ID),
		sourceName:             sourceName,
		sequence:               -1,
		timeValue:              point.ObservedAt.UTC(),
		latitude:               point.Latitude,
		longitude:              point.Longitude,
		horizontalUncertaintyM: 0,
	}, true
}

func trajectoryPointLess(
	left trajectory.TrackPoint4D,
	right trajectory.TrackPoint4D,
) bool {
	leftTime := left.ObservedAt.UTC()
	rightTime := right.ObservedAt.UTC()
	if !leftTime.Equal(rightTime) {
		return leftTime.Before(rightTime)
	}
	leftID := strings.TrimSpace(left.ID)
	rightID := strings.TrimSpace(right.ID)
	if leftID != rightID {
		return leftID < rightID
	}
	leftSource := strings.TrimSpace(left.SourceName)
	rightSource := strings.TrimSpace(right.SourceName)
	if leftSource != rightSource {
		return leftSource < rightSource
	}
	if left.Latitude != right.Latitude {
		return math.Float64bits(left.Latitude) <
			math.Float64bits(right.Latitude)
	}
	return math.Float64bits(left.Longitude) <
		math.Float64bits(right.Longitude)
}

func calculateClosingSpeedProfile(
	samples []positionSample,
	distances []float64,
	maximumGroundSpeedMPS float64,
	maximumSampleCount int,
) (speedProfile, bool) {
	if len(samples) < 2 ||
		len(samples) != len(distances) ||
		!positiveFinite(maximumGroundSpeedMPS) ||
		maximumSampleCount < 1 {
		return speedProfile{}, false
	}

	type speedSample struct {
		closingMPS float64
		groundMPS  float64
	}

	values := make(
		[]speedSample,
		0,
		len(samples)-1,
	)
	for index := 1; index < len(samples); index++ {
		durationSeconds := samples[index].timeValue.Sub(
			samples[index-1].timeValue,
		).Seconds()
		if !positiveFinite(durationSeconds) {
			return speedProfile{}, false
		}

		groundDistanceM := greatCircleDistanceM(
			samples[index-1].latitude,
			samples[index-1].longitude,
			samples[index].latitude,
			samples[index].longitude,
		)
		groundSpeedMPS := groundDistanceM / durationSeconds
		if !nonNegativeFinite(groundSpeedMPS) ||
			groundSpeedMPS > maximumGroundSpeedMPS {
			return speedProfile{}, false
		}

		closingSpeedMPS :=
			(distances[index-1] - distances[index]) /
				durationSeconds
		if !finite(closingSpeedMPS) {
			return speedProfile{}, false
		}

		values = append(
			values,
			speedSample{
				closingMPS: closingSpeedMPS,
				groundMPS:  groundSpeedMPS,
			},
		)
	}
	if len(values) > maximumSampleCount {
		values = values[len(values)-maximumSampleCount:]
	}
	if len(values) == 0 {
		return speedProfile{}, false
	}

	mean := 0.0
	minimum := values[0].closingMPS
	maximum := values[0].closingMPS
	maximumGround := values[0].groundMPS
	for _, value := range values {
		mean += value.closingMPS
		minimum = math.Min(minimum, value.closingMPS)
		maximum = math.Max(maximum, value.closingMPS)
		maximumGround = math.Max(maximumGround, value.groundMPS)
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, value := range values {
		delta := value.closingMPS - mean
		variance += delta * delta
	}
	variance /= float64(len(values))
	stdDev := math.Sqrt(variance)
	if !finite(mean) ||
		!nonNegativeFinite(stdDev) ||
		!finite(minimum) ||
		!finite(maximum) ||
		!nonNegativeFinite(maximumGround) {
		return speedProfile{}, false
	}

	return speedProfile{
		sampleCount:            len(values),
		meanClosingSpeedMPS:    mean,
		closingSpeedStdDevMPS:  stdDev,
		minimumClosingSpeedMPS: minimum,
		maximumClosingSpeedMPS: maximum,
		maximumGroundSpeedMPS:  maximumGround,
	}, true
}

func enforceMinimumArrivalInterval(
	asOfTime time.Time,
	estimatedTime time.Time,
	earliestTime time.Time,
	latestTime time.Time,
	minimumInterval time.Duration,
) (
	time.Time,
	time.Time,
	time.Time,
) {
	if estimatedTime.Before(asOfTime) {
		estimatedTime = asOfTime
	}

	halfInterval := minimumInterval / 2
	minimumEarliest := estimatedTime.Add(-halfInterval)
	minimumLatest := estimatedTime.Add(halfInterval)

	if earliestTime.IsZero() || earliestTime.After(minimumEarliest) {
		earliestTime = minimumEarliest
	}
	if earliestTime.Before(asOfTime) {
		earliestTime = asOfTime
	}

	if latestTime.IsZero() || latestTime.Before(minimumLatest) {
		latestTime = minimumLatest
	}
	if latestTime.Before(estimatedTime) {
		latestTime = estimatedTime
	}
	if earliestTime.After(estimatedTime) {
		earliestTime = estimatedTime
	}

	return earliestTime, estimatedTime, latestTime
}
