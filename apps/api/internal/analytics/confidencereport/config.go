package confidencereport

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrMediumThresholdInvalid = errors.New(
		"medium confidence threshold must be finite, greater than zero, and lower than high threshold",
	)
	ErrHighThresholdInvalid = errors.New(
		"high confidence threshold must be finite, greater than medium threshold, and at most one",
	)
	ErrMaximumPenaltyInvalid = errors.New(
		"maximum confidence penalty must be finite and between zero and one",
	)
	ErrDecimalPrecisionInvalid = errors.New(
		"confidence decimal precision must be between zero and twelve",
	)
)

type Config struct {
	MediumThreshold  float64
	HighThreshold    float64
	MaximumPenalty   float64
	DecimalPrecision int
}

func DefaultConfig() Config {
	return Config{
		MediumThreshold:  0.60,
		HighThreshold:    0.80,
		MaximumPenalty:   1.00,
		DecimalPrecision: 6,
	}
}

func (
	config Config,
) Validate() error {
	if !isFinite(config.MediumThreshold) ||
		config.MediumThreshold <= 0 ||
		config.MediumThreshold >= config.HighThreshold {
		return fmt.Errorf(
			"%w: %f",
			ErrMediumThresholdInvalid,
			config.MediumThreshold,
		)
	}

	if !isFinite(config.HighThreshold) ||
		config.HighThreshold <= config.MediumThreshold ||
		config.HighThreshold > 1 {
		return fmt.Errorf(
			"%w: %f",
			ErrHighThresholdInvalid,
			config.HighThreshold,
		)
	}

	if !isFinite(config.MaximumPenalty) ||
		config.MaximumPenalty < 0 ||
		config.MaximumPenalty > 1 {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumPenaltyInvalid,
			config.MaximumPenalty,
		)
	}

	if config.DecimalPrecision < 0 ||
		config.DecimalPrecision > 12 {
		return fmt.Errorf(
			"%w: %d",
			ErrDecimalPrecisionInvalid,
			config.DecimalPrecision,
		)
	}

	return nil
}

func isFinite(
	value float64,
) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}
