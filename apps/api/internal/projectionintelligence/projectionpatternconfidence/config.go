package projectionpatternconfidence

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrMinimumNeighborCountInvalid = errors.New(
		"minimum neighbor count must be greater than zero",
	)
	ErrTargetNeighborCountInvalid = errors.New(
		"target neighbor count must be greater than or equal to the minimum neighbor count",
	)
	ErrMaximumCandidateAgeInvalid = errors.New(
		"maximum candidate age must be greater than zero",
	)
	ErrMaximumMeanAnchorDistanceInvalid = errors.New(
		"maximum mean anchor distance must be finite and greater than zero",
	)
	ErrMinimumUsableScoreInvalid = errors.New(
		"minimum usable score must be finite and between zero and one",
	)
	ErrConfidenceThresholdInvalid = errors.New(
		"confidence thresholds must satisfy zero < medium <= high <= one",
	)
	ErrComponentWeightInvalid = errors.New(
		"pattern confidence component weights must be finite, non-negative, and sum to one",
	)
)

type Config struct {
	MinimumNeighborCount int
	TargetNeighborCount  int

	MaximumCandidateAge         time.Duration
	MaximumMeanAnchorDistanceKM float64

	MinimumUsableScore float64

	MediumConfidenceMinimum float64
	HighConfidenceMinimum   float64

	SimilarityWeight      float64
	SupportWeight         float64
	FreshnessWeight       float64
	AnchorProximityWeight float64
}

func (config Config) Validate() error {
	if config.MinimumNeighborCount < 1 {
		return fmt.Errorf(
			"%w: %d",
			ErrMinimumNeighborCountInvalid,
			config.MinimumNeighborCount,
		)
	}
	if config.TargetNeighborCount <
		config.MinimumNeighborCount {
		return fmt.Errorf(
			"%w: minimum=%d target=%d",
			ErrTargetNeighborCountInvalid,
			config.MinimumNeighborCount,
			config.TargetNeighborCount,
		)
	}
	if config.MaximumCandidateAge <= 0 {
		return fmt.Errorf(
			"%w: %s",
			ErrMaximumCandidateAgeInvalid,
			config.MaximumCandidateAge,
		)
	}
	if !finite(
		config.MaximumMeanAnchorDistanceKM,
	) ||
		config.MaximumMeanAnchorDistanceKM <= 0 {
		return fmt.Errorf(
			"%w: %f",
			ErrMaximumMeanAnchorDistanceInvalid,
			config.MaximumMeanAnchorDistanceKM,
		)
	}
	if !unitInterval(
		config.MinimumUsableScore,
	) {
		return fmt.Errorf(
			"%w: %f",
			ErrMinimumUsableScoreInvalid,
			config.MinimumUsableScore,
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

	weights := []float64{
		config.SimilarityWeight,
		config.SupportWeight,
		config.FreshnessWeight,
		config.AnchorProximityWeight,
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

func positiveFinite(value float64) bool {
	return finite(value) &&
		value > 0
}

func unitInterval(value float64) bool {
	return finite(value) &&
		value >= 0 &&
		value <= 1
}
