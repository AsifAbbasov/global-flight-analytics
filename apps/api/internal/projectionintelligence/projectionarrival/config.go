package projectionarrival

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrArrivalRadiusInvalid = errors.New(
		"arrival radius must be finite and greater than zero",
	)
	ErrDestinationConfidenceInvalid = errors.New(
		"minimum destination confidence must be finite and between zero and one",
	)
	ErrSpeedSamplePolicyInvalid = errors.New(
		"speed sample policy is invalid",
	)
	ErrMinimumGroundSpeedInvalid = errors.New(
		"minimum ground speed must be finite and greater than zero",
	)
	ErrSpeedUncertaintyMultiplierInvalid = errors.New(
		"speed uncertainty multiplier must be finite and greater than zero",
	)
	ErrMinimumArrivalIntervalInvalid = errors.New(
		"minimum arrival interval must be greater than zero",
	)
	ErrMaximumArrivalDurationInvalid = errors.New(
		"maximum estimated arrival duration must be greater than zero",
	)
	ErrExtrapolationConfidenceLossInvalid = errors.New(
		"maximum extrapolation confidence loss must be finite and between zero and one",
	)
	ErrConfidenceWeightInvalid = errors.New(
		"arrival confidence weights must be finite, non-negative, and sum to one",
	)
	ErrConfidenceThresholdInvalid = errors.New(
		"confidence thresholds must satisfy zero < medium <= high <= one",
	)
)

type Config struct {
	ArrivalRadiusM float64

	MinimumDestinationConfidenceScore float64

	MinimumSpeedSampleCount int
	MaximumSpeedSampleCount int
	MinimumGroundSpeedMPS   float64

	SpeedUncertaintyMultiplier      float64
	MinimumArrivalInterval          time.Duration
	MaximumEstimatedArrivalDuration time.Duration

	MaximumExtrapolationConfidenceLoss float64

	ProjectionConfidenceWeight  float64
	DestinationConfidenceWeight float64
	SpeedStabilityWeight        float64

	MediumConfidenceMinimum float64
	HighConfidenceMinimum   float64
}

func (config Config) Validate() error {
	if !positiveFinite(config.ArrivalRadiusM) {
		return fmt.Errorf(
			"%w: %f",
			ErrArrivalRadiusInvalid,
			config.ArrivalRadiusM,
		)
	}
	if !unitInterval(
		config.MinimumDestinationConfidenceScore,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrDestinationConfidenceInvalid,
			config.MinimumDestinationConfidenceScore,
		)
	}
	if config.MinimumSpeedSampleCount < 1 ||
		config.MaximumSpeedSampleCount <
			config.MinimumSpeedSampleCount {
		return fmt.Errorf(
			"%w: minimum=%d maximum=%d",
			ErrSpeedSamplePolicyInvalid,
			config.MinimumSpeedSampleCount,
			config.MaximumSpeedSampleCount,
		)
	}
	if !positiveFinite(
		config.MinimumGroundSpeedMPS,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrMinimumGroundSpeedInvalid,
			config.MinimumGroundSpeedMPS,
		)
	}
	if !positiveFinite(
		config.SpeedUncertaintyMultiplier,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrSpeedUncertaintyMultiplierInvalid,
			config.SpeedUncertaintyMultiplier,
		)
	}
	if config.MinimumArrivalInterval <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMinimumArrivalIntervalInvalid,
			config.MinimumArrivalInterval,
		)
	}
	if config.MaximumEstimatedArrivalDuration <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumArrivalDurationInvalid,
			config.MaximumEstimatedArrivalDuration,
		)
	}
	if !unitInterval(
		config.MaximumExtrapolationConfidenceLoss,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrExtrapolationConfidenceLossInvalid,
			config.MaximumExtrapolationConfidenceLoss,
		)
	}

	weights := []float64{
		config.ProjectionConfidenceWeight,
		config.DestinationConfidenceWeight,
		config.SpeedStabilityWeight,
	}
	total := 0.0
	for _, weight := range weights {
		if !finite(weight) ||
			weight < 0 {
			return fmt.Errorf(
				"%w: %f",
				ErrConfidenceWeightInvalid,
				weight,
			)
		}
		total += weight
	}
	if math.Abs(total-1) > 1e-9 {
		return fmt.Errorf(
			"%w: total=%f",
			ErrConfidenceWeightInvalid,
			total,
		)
	}

	if !positiveFinite(
		config.MediumConfidenceMinimum,
	) ||
		!positiveFinite(
			config.HighConfidenceMinimum,
		) ||
		config.MediumConfidenceMinimum >
			config.HighConfidenceMinimum ||
		config.HighConfidenceMinimum > 1 {
		return fmt.Errorf(
			"%w: medium=%f high=%f",
			ErrConfidenceThresholdInvalid,
			config.MediumConfidenceMinimum,
			config.HighConfidenceMinimum,
		)
	}

	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func positiveFinite(value float64) bool {
	return finite(value) &&
		value > 0
}

func nonNegativeFinite(value float64) bool {
	return finite(value) &&
		value >= 0
}

func unitInterval(value float64) bool {
	return finite(value) &&
		value >= 0 &&
		value <= 1
}
