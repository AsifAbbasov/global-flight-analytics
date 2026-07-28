package projectionbaseline

import (
	"math"
	"time"
)

const maximumObservationAgeConfidenceLoss = 0.50

type pointConfidenceInput struct {
	TrajectoryQuality     float64
	LatestObservedAt      time.Time
	AsOfTime              time.Time
	ForecastTime          time.Time
	MaximumObservationAge time.Duration
	EffectiveHorizon      time.Duration
	MaximumHorizonLoss    float64
}

func calculatePointConfidence(input pointConfidenceInput) (float64, bool) {
	if !unitInterval(input.TrajectoryQuality) ||
		input.LatestObservedAt.IsZero() ||
		input.AsOfTime.IsZero() ||
		input.ForecastTime.IsZero() ||
		input.MaximumObservationAge <= 0 ||
		input.EffectiveHorizon <= 0 ||
		!unitInterval(input.MaximumHorizonLoss) {
		return 0, false
	}

	observationAge := input.AsOfTime.UTC().Sub(
		input.LatestObservedAt.UTC(),
	)
	if observationAge < 0 {
		return 0, false
	}

	ageProgress := float64(observationAge) /
		float64(input.MaximumObservationAge)
	ageProgress = boundedUnit(ageProgress)
	ageFactor := 1 -
		maximumObservationAgeConfidenceLoss*ageProgress

	horizonProgress := float64(
		input.ForecastTime.UTC().Sub(input.AsOfTime.UTC()),
	) / float64(input.EffectiveHorizon)
	if horizonProgress < 0 || horizonProgress > 1 ||
		math.IsNaN(horizonProgress) || math.IsInf(horizonProgress, 0) {
		return 0, false
	}
	horizonFactor := 1 -
		input.MaximumHorizonLoss*horizonProgress

	score := input.TrajectoryQuality * ageFactor * horizonFactor
	if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 || score > 1 {
		return 0, false
	}

	return boundedUnit(score), true
}

func boundedUnit(value float64) float64 {
	if value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}
	return value
}
