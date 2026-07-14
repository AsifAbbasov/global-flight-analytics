package trajectorybuilder

import (
	"sort"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

type samplingMetrics struct {
	available      bool
	meanSeconds    float64
	maximumSeconds float64
}

func calculateSamplingMetrics(
	points []trajectory.TrackPoint4D,
) (
	samplingMetrics,
	[]flightfeatures.FeatureLimitation,
) {
	timestamps := make(
		[]time.Time,
		0,
		len(points),
	)
	zeroTimestampCount := 0
	nonMonotonicCount := 0
	var previousInputTimestamp time.Time

	for _, point := range points {
		if point.ObservedAt.IsZero() {
			zeroTimestampCount++
			continue
		}

		observedAt := point.ObservedAt.UTC()
		if !previousInputTimestamp.IsZero() &&
			observedAt.Before(previousInputTimestamp) {
			nonMonotonicCount++
		}
		previousInputTimestamp = observedAt
		timestamps = append(timestamps, observedAt)
	}

	limitations := make(
		[]flightfeatures.FeatureLimitation,
		0,
		3,
	)
	if zeroTimestampCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_sampling_timestamp_missing",
				Message: "One or more trajectory points have no observation timestamp and were excluded from sampling metrics.",
			},
		)
	}
	if nonMonotonicCount > 0 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_sampling_input_nonmonotonic",
				Message: "Non-zero point timestamps decrease in input order; sampling metrics use a sorted timestamp copy.",
			},
		)
	}
	if len(timestamps) < 2 {
		limitations = append(
			limitations,
			flightfeatures.FeatureLimitation{
				Code:    "trajectory_sampling_evidence_insufficient",
				Message: "At least two non-zero point timestamps are required for sampling interval metrics.",
			},
		)

		return samplingMetrics{}, limitations
	}

	sort.SliceStable(
		timestamps,
		func(left int, right int) bool {
			return timestamps[left].Before(
				timestamps[right],
			)
		},
	)

	totalSeconds := 0.0
	maximumSeconds := 0.0
	for index := 1; index < len(timestamps); index++ {
		intervalSeconds := timestamps[index].
			Sub(timestamps[index-1]).
			Seconds()
		totalSeconds += intervalSeconds
		if intervalSeconds > maximumSeconds {
			maximumSeconds = intervalSeconds
		}
	}

	return samplingMetrics{
		available: true,
		meanSeconds: totalSeconds /
			float64(len(timestamps)-1),
		maximumSeconds: maximumSeconds,
	}, limitations
}
