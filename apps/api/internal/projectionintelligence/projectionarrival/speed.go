package projectionarrival

import (
	"math"
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

func buildPositionSamples(
	current trajectory.FlightTrajectory,
	projection projectioncontract.Result,
) []positionSample {
	samples := make(
		[]positionSample,
		0,
		len(projection.Points)+1,
	)

	if endpoint, exists :=
		currentEndpointAt(
			current,
			projection.Horizon.AsOfTime,
		); exists {
		samples = append(
			samples,
			endpoint,
		)
	}

	for _, point := range projection.Points {
		samples = append(
			samples,
			positionSample{
				timeValue: point.ForecastTime.UTC(),
				latitude:  point.Position.Latitude,
				longitude: point.Position.Longitude,
				horizontalUncertaintyM: point.Uncertainty.
					HorizontalRadiusM,
			},
		)
	}

	sort.SliceStable(
		samples,
		func(left int, right int) bool {
			return samples[left].
				timeValue.Before(
				samples[right].
					timeValue,
			)
		},
	)

	result := make(
		[]positionSample,
		0,
		len(samples),
	)
	for _, sample := range samples {
		if sample.timeValue.IsZero() ||
			!validLatitude(sample.latitude) ||
			!validLongitude(sample.longitude) ||
			!nonNegativeFinite(
				sample.
					horizontalUncertaintyM,
			) {
			continue
		}
		if len(result) > 0 &&
			sample.timeValue.Equal(
				result[len(result)-1].
					timeValue,
			) {
			result[len(result)-1] =
				sample
			continue
		}
		result = append(
			result,
			sample,
		)
	}

	return result
}

func currentEndpointAt(
	item trajectory.FlightTrajectory,
	asOfTime time.Time,
) (positionSample, bool) {
	type indexedPoint struct {
		point trajectory.TrackPoint4D
		index int
	}

	candidates := make(
		[]indexedPoint,
		0,
		len(item.Points),
	)
	for index, point := range item.Points {
		if point.ObservedAt.IsZero() ||
			point.ObservedAt.UTC().After(
				asOfTime.UTC(),
			) ||
			!validLatitude(
				point.Latitude,
			) ||
			!validLongitude(
				point.Longitude,
			) {
			continue
		}
		candidates = append(
			candidates,
			indexedPoint{
				point: point,
				index: index,
			},
		)
	}

	if len(candidates) == 0 {
		return positionSample{}, false
	}

	sort.SliceStable(
		candidates,
		func(left int, right int) bool {
			leftTime := candidates[left].
				point.ObservedAt.UTC()
			rightTime := candidates[right].
				point.ObservedAt.UTC()
			if leftTime.Equal(rightTime) {
				return candidates[left].index <
					candidates[right].index
			}

			return leftTime.Before(
				rightTime,
			)
		},
	)

	point := candidates[len(candidates)-1].
		point

	return positionSample{
		timeValue:              point.ObservedAt.UTC(),
		latitude:               point.Latitude,
		longitude:              point.Longitude,
		horizontalUncertaintyM: 0,
	}, true
}

func calculateSpeedProfile(
	samples []positionSample,
	minimumGroundSpeedMPS float64,
	maximumSampleCount int,
) (speedProfile, bool) {
	speeds := make(
		[]float64,
		0,
		len(samples)-1,
	)

	for index := 1; index < len(samples); index++ {
		durationSeconds := samples[index].
			timeValue.Sub(
			samples[index-1].
				timeValue,
		).Seconds()
		if durationSeconds <= 0 ||
			!finite(durationSeconds) {
			continue
		}

		distanceM := greatCircleDistanceM(
			samples[index-1].latitude,
			samples[index-1].longitude,
			samples[index].latitude,
			samples[index].longitude,
		)
		if !nonNegativeFinite(distanceM) {
			continue
		}

		speedMPS :=
			distanceM /
				durationSeconds
		if !finite(speedMPS) ||
			speedMPS <
				minimumGroundSpeedMPS {
			continue
		}

		speeds = append(
			speeds,
			speedMPS,
		)
	}

	if len(speeds) == 0 {
		return speedProfile{}, false
	}
	if len(speeds) > maximumSampleCount {
		speeds = speeds[len(speeds)-maximumSampleCount:]
	}

	mean := 0.0
	minimum := speeds[0]
	maximum := speeds[0]
	for _, speed := range speeds {
		mean += speed
		if speed < minimum {
			minimum = speed
		}
		if speed > maximum {
			maximum = speed
		}
	}
	mean /= float64(len(speeds))

	variance := 0.0
	for _, speed := range speeds {
		delta := speed - mean
		variance += delta * delta
	}
	variance /= float64(len(speeds))
	stdDev := math.Sqrt(variance)

	if !positiveFinite(mean) ||
		!nonNegativeFinite(stdDev) {
		return speedProfile{}, false
	}

	return speedProfile{
		sampleCount: len(speeds),
		meanMPS:     mean,
		stdDevMPS:   stdDev,
		minimumMPS:  minimum,
		maximumMPS:  maximum,
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

	halfInterval :=
		minimumInterval / 2
	minimumEarliest :=
		estimatedTime.Add(
			-halfInterval,
		)
	minimumLatest :=
		estimatedTime.Add(
			halfInterval,
		)

	if earliestTime.IsZero() ||
		earliestTime.After(
			minimumEarliest,
		) {
		earliestTime =
			minimumEarliest
	}
	if earliestTime.Before(asOfTime) {
		earliestTime = asOfTime
	}

	if latestTime.IsZero() ||
		latestTime.Before(
			minimumLatest,
		) {
		latestTime =
			minimumLatest
	}
	if latestTime.Before(
		estimatedTime,
	) {
		latestTime =
			estimatedTime
	}
	if earliestTime.After(
		estimatedTime,
	) {
		earliestTime =
			estimatedTime
	}

	return earliestTime,
		estimatedTime,
		latestTime
}
