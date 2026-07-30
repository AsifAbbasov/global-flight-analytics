package projectionfreshness

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrMaximumNewestNeighborAgeInvalid = errors.New(
		"maximum newest-neighbor age must be greater than zero",
	)
	ErrMaximumMeanNeighborAgeInvalid = errors.New(
		"maximum mean-neighbor age must be greater than zero",
	)
	ErrMaximumOldestNeighborAgeInvalid = errors.New(
		"maximum oldest-neighbor age must be greater than zero",
	)
	ErrFreshnessAgeThresholdOrderInvalid = errors.New(
		"freshness age thresholds must satisfy newest <= mean <= oldest",
	)
	ErrRecentNeighborAgeLimitInvalid = errors.New(
		"recent-neighbor age limit must be greater than zero and not exceed the oldest-neighbor maximum",
	)
	ErrRecentNeighborCountInvalid = errors.New(
		"recent-neighbor counts must satisfy zero < minimum <= target",
	)
	ErrFreshnessScoreThresholdInvalid = errors.New(
		"freshness score thresholds must satisfy zero < minimum <= complete <= one",
	)
	ErrFreshnessWeightInvalid = errors.New(
		"freshness component weights must be finite, non-negative, and sum to one",
	)
)

type Config struct {
	MaximumNewestNeighborAge time.Duration
	MaximumMeanNeighborAge   time.Duration
	MaximumOldestNeighborAge time.Duration

	RecentNeighborAgeLimit     time.Duration
	MinimumRecentNeighborCount int
	TargetRecentNeighborCount  int

	MinimumUsableScore   float64
	CompleteScoreMinimum float64

	NewestAgeWeight     float64
	MeanAgeWeight       float64
	OldestAgeWeight     float64
	RecentSupportWeight float64
}

func (config Config) Validate() error {
	if err := validateAgeThresholds(config); err != nil {
		return err
	}
	if err := validateRecentSupportPolicy(config); err != nil {
		return err
	}
	if err := validateScoreThresholds(config); err != nil {
		return err
	}
	return validateWeights(config)
}

func validateAgeThresholds(config Config) error {
	if config.MaximumNewestNeighborAge <= 0 {
		return fmt.Errorf("%w: %s", ErrMaximumNewestNeighborAgeInvalid, config.MaximumNewestNeighborAge)
	}
	if config.MaximumMeanNeighborAge <= 0 {
		return fmt.Errorf("%w: %s", ErrMaximumMeanNeighborAgeInvalid, config.MaximumMeanNeighborAge)
	}
	if config.MaximumOldestNeighborAge <= 0 {
		return fmt.Errorf("%w: %s", ErrMaximumOldestNeighborAgeInvalid, config.MaximumOldestNeighborAge)
	}
	if config.MaximumNewestNeighborAge > config.MaximumMeanNeighborAge ||
		config.MaximumMeanNeighborAge > config.MaximumOldestNeighborAge {
		return fmt.Errorf(
			"%w: newest=%s mean=%s oldest=%s",
			ErrFreshnessAgeThresholdOrderInvalid,
			config.MaximumNewestNeighborAge,
			config.MaximumMeanNeighborAge,
			config.MaximumOldestNeighborAge,
		)
	}
	return nil
}

func validateRecentSupportPolicy(config Config) error {
	if config.RecentNeighborAgeLimit <= 0 ||
		config.RecentNeighborAgeLimit > config.MaximumOldestNeighborAge {
		return fmt.Errorf(
			"%w: recent=%s oldest=%s",
			ErrRecentNeighborAgeLimitInvalid,
			config.RecentNeighborAgeLimit,
			config.MaximumOldestNeighborAge,
		)
	}
	if config.MinimumRecentNeighborCount < 1 ||
		config.TargetRecentNeighborCount < config.MinimumRecentNeighborCount {
		return fmt.Errorf(
			"%w: minimum=%d target=%d",
			ErrRecentNeighborCountInvalid,
			config.MinimumRecentNeighborCount,
			config.TargetRecentNeighborCount,
		)
	}
	return nil
}

func validateScoreThresholds(config Config) error {
	if !positiveUnitInterval(config.MinimumUsableScore) ||
		!positiveUnitInterval(config.CompleteScoreMinimum) ||
		config.MinimumUsableScore > config.CompleteScoreMinimum {
		return fmt.Errorf(
			"%w: minimum=%f complete=%f",
			ErrFreshnessScoreThresholdInvalid,
			config.MinimumUsableScore,
			config.CompleteScoreMinimum,
		)
	}
	return nil
}

func validateWeights(config Config) error {
	weights := []float64{
		config.NewestAgeWeight,
		config.MeanAgeWeight,
		config.OldestAgeWeight,
		config.RecentSupportWeight,
	}
	total := 0.0
	for _, weight := range weights {
		if !finite(weight) || weight < 0 {
			return fmt.Errorf("%w: %f", ErrFreshnessWeightInvalid, weight)
		}
		total += weight
	}
	if math.Abs(total-1) > scoreComparisonTolerance {
		return fmt.Errorf("%w: total=%f", ErrFreshnessWeightInvalid, total)
	}
	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func unitInterval(value float64) bool {
	return finite(value) && value >= 0 && value <= 1
}

func positiveUnitInterval(value float64) bool {
	return finite(value) && value > 0 && value <= 1
}

func clampUnit(value float64) float64 {
	if !finite(value) || value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}
	return value
}
