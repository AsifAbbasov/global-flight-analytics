package projectionroutefrequency

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrObservationCountInvalid = errors.New(
		"route observation counts must satisfy zero < minimum <= target",
	)
	ErrDistinctDayCountInvalid = errors.New(
		"route distinct-day counts must satisfy zero < minimum <= target",
	)
	ErrRecentWindowInvalid = errors.New(
		"recent route-history window must be greater than zero",
	)
	ErrRecentObservationCountInvalid = errors.New(
		"recent route observation counts must satisfy zero < minimum <= target",
	)
	ErrMaximumLatestObservationAgeInvalid = errors.New(
		"maximum latest route-observation age must be greater than zero",
	)
	ErrMinimumRouteConfidenceInvalid = errors.New(
		"minimum route confidence must be finite and between zero and one",
	)
	ErrScoreThresholdInvalid = errors.New(
		"route-frequency score thresholds must satisfy zero <= minimum <= complete <= one",
	)
	ErrComponentWeightInvalid = errors.New(
		"route-frequency component weights must be finite, non-negative, and sum to one",
	)
)

type Config struct {
	MinimumObservationCount int
	TargetObservationCount  int

	MinimumDistinctDayCount int
	TargetDistinctDayCount  int

	RecentWindow                  time.Duration
	MinimumRecentObservationCount int
	TargetRecentObservationCount  int

	MaximumLatestObservationAge time.Duration
	MinimumRouteConfidenceScore float64

	MinimumUsableScore   float64
	CompleteScoreMinimum float64

	ObservationCountWeight  float64
	DistinctDayWeight       float64
	RecentObservationWeight float64
	LatestObservationWeight float64
	RouteConfidenceWeight   float64
}

func (config Config) Validate() error {
	if config.MinimumObservationCount < 1 ||
		config.TargetObservationCount <
			config.MinimumObservationCount {
		return fmt.Errorf(
			"%w: minimum=%d target=%d",
			ErrObservationCountInvalid,
			config.MinimumObservationCount,
			config.TargetObservationCount,
		)
	}
	if config.MinimumDistinctDayCount < 1 ||
		config.TargetDistinctDayCount <
			config.MinimumDistinctDayCount {
		return fmt.Errorf(
			"%w: minimum=%d target=%d",
			ErrDistinctDayCountInvalid,
			config.MinimumDistinctDayCount,
			config.TargetDistinctDayCount,
		)
	}
	if config.RecentWindow <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrRecentWindowInvalid,
			config.RecentWindow,
		)
	}
	if config.MinimumRecentObservationCount < 1 ||
		config.TargetRecentObservationCount <
			config.MinimumRecentObservationCount {
		return fmt.Errorf(
			"%w: minimum=%d target=%d",
			ErrRecentObservationCountInvalid,
			config.MinimumRecentObservationCount,
			config.TargetRecentObservationCount,
		)
	}
	if config.MaximumLatestObservationAge <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumLatestObservationAgeInvalid,
			config.MaximumLatestObservationAge,
		)
	}
	if !unitInterval(
		config.MinimumRouteConfidenceScore,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrMinimumRouteConfidenceInvalid,
			config.MinimumRouteConfidenceScore,
		)
	}
	if !unitInterval(config.MinimumUsableScore) ||
		!unitInterval(config.CompleteScoreMinimum) ||
		config.MinimumUsableScore >
			config.CompleteScoreMinimum {
		return fmt.Errorf(
			"%w: minimum=%f complete=%f",
			ErrScoreThresholdInvalid,
			config.MinimumUsableScore,
			config.CompleteScoreMinimum,
		)
	}

	weights := []float64{
		config.ObservationCountWeight,
		config.DistinctDayWeight,
		config.RecentObservationWeight,
		config.LatestObservationWeight,
		config.RouteConfidenceWeight,
	}
	total := 0.0
	for _, weight := range weights {
		if !finite(weight) ||
			weight < 0 {
			return fmt.Errorf(
				"%w: %f",
				ErrComponentWeightInvalid,
				weight,
			)
		}
		total += weight
	}
	if math.Abs(total-1) > 1e-9 {
		return fmt.Errorf(
			"%w: total=%f",
			ErrComponentWeightInvalid,
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
