package projectionevaluation

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrMaximumInterpolationGapInvalid = errors.New(
		"maximum interpolation gap must be greater than zero",
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
)

type Config struct {
	MaximumInterpolationGap    time.Duration
	MinimumEvaluatedPointCount int

	MaximumHorizontalErrorM float64
	MaximumAltitudeErrorM   float64
}

func (config Config) Validate() error {
	if config.MaximumInterpolationGap <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumInterpolationGapInvalid,
			config.MaximumInterpolationGap,
		)
	}
	if config.MinimumEvaluatedPointCount < 1 {
		return fmt.Errorf(
			"%w: %d",
			ErrMinimumEvaluatedPointCountInvalid,
			config.MinimumEvaluatedPointCount,
		)
	}
	if !positiveFinite(
		config.MaximumHorizontalErrorM,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumHorizontalErrorInvalid,
			config.MaximumHorizontalErrorM,
		)
	}
	if !positiveFinite(
		config.MaximumAltitudeErrorM,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumAltitudeErrorInvalid,
			config.MaximumAltitudeErrorM,
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
