package projectionevaluation

import (
	"errors"
	"fmt"
	"math"
	"time"
)

const EvaluationPolicyVersion = "projection-replay-evaluation-policy-v2"

var (
	ErrMaximumInterpolationGapInvalid = errors.New(
		"maximum interpolation gap must be greater than zero",
	)
	ErrMaximumTruthGroundSpeedInvalid = errors.New(
		"maximum truth ground speed must be finite and greater than zero",
	)
	ErrMaximumTruthVerticalRateInvalid = errors.New(
		"maximum truth vertical rate must be finite and greater than zero",
	)
	ErrMinimumEvaluatedPointCountInvalid = errors.New(
		"minimum evaluated point count must be greater than zero",
	)
	ErrMaximumHorizontalErrorInvalid = errors.New(
		"maximum horizontal error must be finite and greater than zero",
	)
	ErrMaximumAltitudeErrorInvalid = errors.New(
		"maximum altitude error must be finite and greater than zero",
	)
	ErrLeadTimeBucketSizeInvalid = errors.New(
		"lead-time bucket size must be greater than zero",
	)
)

type Config struct {
	MaximumInterpolationGap     time.Duration
	MaximumTruthGroundSpeedMPS  float64
	MaximumTruthVerticalRateMPS float64
	MinimumEvaluatedPointCount  int

	MaximumHorizontalErrorM float64
	MaximumAltitudeErrorM   float64
	LeadTimeBucketSize      time.Duration
}

func (config Config) Validate() error {
	if config.MaximumInterpolationGap <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumInterpolationGapInvalid,
			config.MaximumInterpolationGap,
		)
	}
	if !positiveFinite(config.MaximumTruthGroundSpeedMPS) {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumTruthGroundSpeedInvalid,
			config.MaximumTruthGroundSpeedMPS,
		)
	}
	if !positiveFinite(config.MaximumTruthVerticalRateMPS) {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumTruthVerticalRateInvalid,
			config.MaximumTruthVerticalRateMPS,
		)
	}
	if config.MinimumEvaluatedPointCount < 1 {
		return fmt.Errorf(
			"%w: %d",
			ErrMinimumEvaluatedPointCountInvalid,
			config.MinimumEvaluatedPointCount,
		)
	}
	if !positiveFinite(config.MaximumHorizontalErrorM) {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumHorizontalErrorInvalid,
			config.MaximumHorizontalErrorM,
		)
	}
	if !positiveFinite(config.MaximumAltitudeErrorM) {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumAltitudeErrorInvalid,
			config.MaximumAltitudeErrorM,
		)
	}
	if config.LeadTimeBucketSize <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrLeadTimeBucketSizeInvalid,
			config.LeadTimeBucketSize,
		)
	}

	return nil
}

func (config Config) snapshot() EvaluationPolicy {
	policy := EvaluationPolicy{
		Version:                     EvaluationPolicyVersion,
		MaximumInterpolationGap:     config.MaximumInterpolationGap,
		MaximumTruthGroundSpeedMPS:  config.MaximumTruthGroundSpeedMPS,
		MaximumTruthVerticalRateMPS: config.MaximumTruthVerticalRateMPS,
		MinimumEvaluatedPointCount:  config.MinimumEvaluatedPointCount,
		MaximumHorizontalErrorM:     config.MaximumHorizontalErrorM,
		MaximumAltitudeErrorM:       config.MaximumAltitudeErrorM,
		LeadTimeBucketSize:          config.LeadTimeBucketSize,
	}
	policy.InputFingerprint = evaluationPolicyFingerprint(policy)
	return policy
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func positiveFinite(value float64) bool {
	return finite(value) && value > 0
}

func nonNegativeFinite(value float64) bool {
	return finite(value) && value >= 0
}

func unitInterval(value float64) bool {
	return finite(value) && value >= 0 && value <= 1
}

func clampUnit(value float64) float64 {
	switch {
	case !finite(value), value <= 0:
		return 0
	case value >= 1:
		return 1
	default:
		return value
	}
}
