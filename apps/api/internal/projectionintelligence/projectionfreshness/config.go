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
	ErrRecentNeighborAgeLimitInvalid = errors.New(
		"recent-neighbor age limit must be greater than zero",
	)
	ErrRecentNeighborCountInvalid = errors.New(
		"recent-neighbor counts must satisfy zero < minimum <= target",
	)
	ErrFreshnessScoreThresholdInvalid = errors.New(
		"freshness score thresholds must satisfy zero <= minimum <= complete <= one",
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
	if config.MaximumNewestNeighborAge <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumNewestNeighborAgeInvalid,
			config.MaximumNewestNeighborAge,
		)
	}
	if config.MaximumMeanNeighborAge <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumMeanNeighborAgeInvalid,
			config.MaximumMeanNeighborAge,
		)
	}
	if config.MaximumOldestNeighborAge <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumOldestNeighborAgeInvalid,
			config.MaximumOldestNeighborAge,
		)
	}
	if config.RecentNeighborAgeLimit <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrRecentNeighborAgeLimitInvalid,
			config.RecentNeighborAgeLimit,
		)
	}
	if config.MinimumRecentNeighborCount < 1 ||
		config.TargetRecentNeighborCount <
			config.MinimumRecentNeighborCount {
		return fmt.Errorf(
			"%w: minimum=%d target=%d",
			ErrRecentNeighborCountInvalid,
			config.MinimumRecentNeighborCount,
			config.TargetRecentNeighborCount,
		)
	}
	if !unitInterval(config.MinimumUsableScore) ||
		!unitInterval(config.CompleteScoreMinimum) ||
		config.MinimumUsableScore >
			config.CompleteScoreMinimum {
		return fmt.Errorf(
			"%w: minimum=%f complete=%f",
			ErrFreshnessScoreThresholdInvalid,
			config.MinimumUsableScore,
			config.CompleteScoreMinimum,
		)
	}

	weights := []float64{
		config.NewestAgeWeight,
		config.MeanAgeWeight,
		config.OldestAgeWeight,
		config.RecentSupportWeight,
	}
	total := 0.0
	for _, weight := range weights {
		if !finite(weight) ||
			weight < 0 {
			return fmt.Errorf(
				"%w: %f",
				ErrFreshnessWeightInvalid,
				weight,
			)
		}
		total += weight
	}
	if math.Abs(total-1) > 1e-9 {
		return fmt.Errorf(
			"%w: total=%f",
			ErrFreshnessWeightInvalid,
			total,
		)
	}

	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) &&
		!math.IsInf(value, 0)
}

func unitInterval(value float64) bool {
	return finite(value) &&
		value >= 0 &&
		value <= 1
}

func clampUnit(value float64) float64 {
	if !finite(value) ||
		value <= 0 {
		return 0
	}
	if value >= 1 {
		return 1
	}

	return value
}
