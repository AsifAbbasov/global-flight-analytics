package trajectorybuilder

import (
	"context"
	"math"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func calculateSamplingMetrics(
	ctx context.Context,
	points []canonicalPoint,
) (samplingMetrics, []flightfeatures.FeatureLimitation, error) {
	if len(points) < 2 {
		return samplingMetrics{}, []flightfeatures.FeatureLimitation{{
			Code:    flightfeatures.TrajectoryLimitationSamplingEvidenceInsufficient,
			Message: "At least two unique temporally eligible point timestamps are required for sampling interval metrics.",
		}}, nil
	}

	total := kahanAccumulator{}
	maximumSeconds := 0.0
	for index := 1; index < len(points); index++ {
		if index%contextCheckInterval == 0 {
			if err := ctx.Err(); err != nil {
				return samplingMetrics{}, nil, err
			}
		}
		intervalSeconds := points[index].observedAt.Sub(points[index-1].observedAt).Seconds()
		if intervalSeconds <= 0 || math.IsNaN(intervalSeconds) || math.IsInf(intervalSeconds, 0) {
			return samplingMetrics{}, []flightfeatures.FeatureLimitation{{
				Code:    flightfeatures.TrajectoryLimitationSamplingIntervalInvalid,
				Message: "Canonical point timestamps produced a non-positive or non-finite sampling interval.",
			}}, nil
		}
		total.Add(intervalSeconds)
		if intervalSeconds > maximumSeconds {
			maximumSeconds = intervalSeconds
		}
	}
	meanSeconds := total.Value() / float64(len(points)-1)
	if math.IsNaN(meanSeconds) || math.IsInf(meanSeconds, 0) {
		return samplingMetrics{}, []flightfeatures.FeatureLimitation{{
			Code:    flightfeatures.TrajectoryLimitationSamplingAggregateNonFinite,
			Message: "Sampling interval aggregation produced a non-finite result.",
		}}, nil
	}
	return samplingMetrics{
		available:      true,
		meanSeconds:    meanSeconds,
		maximumSeconds: maximumSeconds,
	}, nil, nil
}

type kahanAccumulator struct {
	sum          float64
	compensation float64
}

func (accumulator *kahanAccumulator) Add(value float64) {
	adjusted := value - accumulator.compensation
	next := accumulator.sum + adjusted
	accumulator.compensation = (next - accumulator.sum) - adjusted
	accumulator.sum = next
}

func (accumulator kahanAccumulator) Value() float64 {
	return accumulator.sum
}
