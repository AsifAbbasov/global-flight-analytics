package projectionpatternconfidence

import (
	"errors"
	"fmt"
	"math"
	"time"
)

const (
	defaultMinimumSimilarityScore                 = 0.50
	defaultMaximumSimilarityStandardDeviation     = 0.20
	defaultContinuationAgreementSampleCount       = 4
	defaultContinuationDivergenceNormalizationMPS = 100.0
	defaultMaximumContinuationDivergenceMPS       = 200.0
	defaultContinuationAgreementWeight            = 0.20
)

var (
	ErrMinimumNeighborCountInvalid = errors.New(
		"minimum neighbor count must be at least two",
	)
	ErrTargetNeighborCountInvalid = errors.New(
		"target neighbor count must be greater than or equal to the minimum neighbor count",
	)
	ErrMinimumSimilarityScoreInvalid = errors.New(
		"minimum similarity score must be finite, greater than zero, and at most one",
	)
	ErrMaximumSimilarityStandardDeviationInvalid = errors.New(
		"maximum similarity standard deviation must be finite, greater than zero, and at most one half",
	)
	ErrAnchorDistanceNormalizationInvalid = errors.New(
		"anchor distance normalization must be finite and greater than zero",
	)
	ErrMinimumUsableScoreInvalid = errors.New(
		"minimum usable score must be finite, greater than zero, and at most one",
	)
	ErrConfidenceThresholdInvalid = errors.New(
		"confidence thresholds must satisfy zero < medium <= high <= one",
	)
	ErrContinuationAgreementSampleCountInvalid = errors.New(
		"continuation agreement sample count must be between one and thirty-two",
	)
	ErrContinuationDivergencePolicyInvalid = errors.New(
		"continuation divergence normalization and maximum must be finite, positive, and ordered",
	)
	ErrComponentWeightInvalid = errors.New(
		"pattern confidence component weights must be finite, positive, and sum to one",
	)
	ErrLegacyConfigConflict = errors.New(
		"legacy and canonical pattern confidence config values conflict",
	)

	// Deprecated compatibility aliases.
	ErrMaximumCandidateAgeInvalid = errors.New(
		"maximum candidate age is no longer part of pattern confidence policy",
	)
	ErrMaximumMeanAnchorDistanceInvalid = ErrAnchorDistanceNormalizationInvalid
)

type Config struct {
	MinimumNeighborCount int
	TargetNeighborCount  int

	MinimumSimilarityScore             float64
	MaximumSimilarityStandardDeviation float64
	AnchorDistanceNormalizationKM      float64

	MinimumUsableScore float64

	MediumConfidenceMinimum float64
	HighConfidenceMinimum   float64

	ContinuationAgreementSampleCount       int
	ContinuationDivergenceNormalizationMPS float64
	MaximumContinuationDivergenceMPS       float64

	SimilarityStrengthWeight    float64
	SupportWeight               float64
	SimilarityConsistencyWeight float64
	AnchorProximityWeight       float64
	ContinuationAgreementWeight float64

	// Deprecated: ignored. Freshness belongs exclusively to projectionfreshness.
	MaximumCandidateAge time.Duration
	// Deprecated: use AnchorDistanceNormalizationKM.
	MaximumMeanAnchorDistanceKM float64
	// Deprecated: use SimilarityStrengthWeight.
	SimilarityWeight float64
	// Deprecated: use SimilarityConsistencyWeight.
	FreshnessWeight float64
}

func (config Config) Validate() error {
	_, err := normalizeAndValidateConfig(config)
	return err
}

func normalizeAndValidateConfig(config Config) (Config, error) {
	normalized := config

	if normalized.MinimumSimilarityScore == 0 {
		normalized.MinimumSimilarityScore = defaultMinimumSimilarityScore
	}
	if normalized.MaximumSimilarityStandardDeviation == 0 {
		normalized.MaximumSimilarityStandardDeviation =
			defaultMaximumSimilarityStandardDeviation
	}
	if normalized.ContinuationAgreementSampleCount == 0 {
		normalized.ContinuationAgreementSampleCount =
			defaultContinuationAgreementSampleCount
	}
	if normalized.ContinuationDivergenceNormalizationMPS == 0 {
		normalized.ContinuationDivergenceNormalizationMPS =
			defaultContinuationDivergenceNormalizationMPS
	}
	if normalized.MaximumContinuationDivergenceMPS == 0 {
		normalized.MaximumContinuationDivergenceMPS =
			defaultMaximumContinuationDivergenceMPS
	}

	if err := normalizeLegacyFloat(
		&normalized.AnchorDistanceNormalizationKM,
		config.MaximumMeanAnchorDistanceKM,
	); err != nil {
		return Config{}, fmt.Errorf("%w: anchor distance", err)
	}
	if err := normalizeLegacyFloat(
		&normalized.SimilarityStrengthWeight,
		config.SimilarityWeight,
	); err != nil {
		return Config{}, fmt.Errorf("%w: similarity strength weight", err)
	}
	if err := normalizeLegacyFloat(
		&normalized.SimilarityConsistencyWeight,
		config.FreshnessWeight,
	); err != nil {
		return Config{}, fmt.Errorf("%w: similarity consistency weight", err)
	}

	if normalized.ContinuationAgreementWeight == 0 {
		legacyTotal := normalized.SimilarityStrengthWeight +
			normalized.SupportWeight +
			normalized.SimilarityConsistencyWeight +
			normalized.AnchorProximityWeight
		if math.Abs(legacyTotal-1) > scoreComparisonTolerance {
			return Config{}, fmt.Errorf(
				"%w: legacy total=%f",
				ErrComponentWeightInvalid,
				legacyTotal,
			)
		}
		scale := 1 - defaultContinuationAgreementWeight
		normalized.SimilarityStrengthWeight *= scale
		normalized.SupportWeight *= scale
		normalized.SimilarityConsistencyWeight *= scale
		normalized.AnchorProximityWeight *= scale
		normalized.ContinuationAgreementWeight =
			defaultContinuationAgreementWeight
	}

	if normalized.MinimumNeighborCount < 2 {
		return Config{}, fmt.Errorf(
			"%w: %d",
			ErrMinimumNeighborCountInvalid,
			normalized.MinimumNeighborCount,
		)
	}
	if normalized.TargetNeighborCount < normalized.MinimumNeighborCount {
		return Config{}, fmt.Errorf(
			"%w: minimum=%d target=%d",
			ErrTargetNeighborCountInvalid,
			normalized.MinimumNeighborCount,
			normalized.TargetNeighborCount,
		)
	}
	if !positiveUnitInterval(normalized.MinimumSimilarityScore) {
		return Config{}, fmt.Errorf(
			"%w: %f",
			ErrMinimumSimilarityScoreInvalid,
			normalized.MinimumSimilarityScore,
		)
	}
	if !positiveFinite(normalized.MaximumSimilarityStandardDeviation) ||
		normalized.MaximumSimilarityStandardDeviation > 0.5 {
		return Config{}, fmt.Errorf(
			"%w: %f",
			ErrMaximumSimilarityStandardDeviationInvalid,
			normalized.MaximumSimilarityStandardDeviation,
		)
	}
	if !positiveFinite(normalized.AnchorDistanceNormalizationKM) {
		return Config{}, fmt.Errorf(
			"%w: %f",
			ErrAnchorDistanceNormalizationInvalid,
			normalized.AnchorDistanceNormalizationKM,
		)
	}
	if !positiveUnitInterval(normalized.MinimumUsableScore) {
		return Config{}, fmt.Errorf(
			"%w: %f",
			ErrMinimumUsableScoreInvalid,
			normalized.MinimumUsableScore,
		)
	}
	if !positiveFinite(normalized.MediumConfidenceMinimum) ||
		!positiveFinite(normalized.HighConfidenceMinimum) ||
		normalized.MediumConfidenceMinimum > normalized.HighConfidenceMinimum ||
		normalized.HighConfidenceMinimum > 1 {
		return Config{}, fmt.Errorf(
			"%w: medium=%f high=%f",
			ErrConfidenceThresholdInvalid,
			normalized.MediumConfidenceMinimum,
			normalized.HighConfidenceMinimum,
		)
	}
	if normalized.ContinuationAgreementSampleCount < 1 ||
		normalized.ContinuationAgreementSampleCount > 32 {
		return Config{}, fmt.Errorf(
			"%w: %d",
			ErrContinuationAgreementSampleCountInvalid,
			normalized.ContinuationAgreementSampleCount,
		)
	}
	if !positiveFinite(normalized.ContinuationDivergenceNormalizationMPS) ||
		!positiveFinite(normalized.MaximumContinuationDivergenceMPS) ||
		normalized.MaximumContinuationDivergenceMPS <
			normalized.ContinuationDivergenceNormalizationMPS {
		return Config{}, fmt.Errorf(
			"%w: normalization=%f maximum=%f",
			ErrContinuationDivergencePolicyInvalid,
			normalized.ContinuationDivergenceNormalizationMPS,
			normalized.MaximumContinuationDivergenceMPS,
		)
	}

	weights := []float64{
		normalized.SimilarityStrengthWeight,
		normalized.SupportWeight,
		normalized.SimilarityConsistencyWeight,
		normalized.AnchorProximityWeight,
		normalized.ContinuationAgreementWeight,
	}
	total := 0.0
	for _, weight := range weights {
		if !positiveFinite(weight) {
			return Config{}, fmt.Errorf(
				"%w: %f",
				ErrComponentWeightInvalid,
				weight,
			)
		}
		total += weight
	}
	if math.Abs(total-1) > scoreComparisonTolerance {
		return Config{}, fmt.Errorf(
			"%w: total=%f",
			ErrComponentWeightInvalid,
			total,
		)
	}

	return normalized, nil
}

func (config Config) policySnapshot() Policy {
	return Policy{
		MinimumNeighborCount:                   config.MinimumNeighborCount,
		TargetNeighborCount:                    config.TargetNeighborCount,
		MinimumSimilarityScore:                 config.MinimumSimilarityScore,
		MaximumSimilarityStandardDeviation:     config.MaximumSimilarityStandardDeviation,
		AnchorDistanceNormalizationKM:          config.AnchorDistanceNormalizationKM,
		MinimumUsableScore:                     config.MinimumUsableScore,
		MediumConfidenceMinimum:                config.MediumConfidenceMinimum,
		HighConfidenceMinimum:                  config.HighConfidenceMinimum,
		ContinuationAgreementSampleCount:       config.ContinuationAgreementSampleCount,
		ContinuationDivergenceNormalizationMPS: config.ContinuationDivergenceNormalizationMPS,
		MaximumContinuationDivergenceMPS:       config.MaximumContinuationDivergenceMPS,
		SimilarityStrengthWeight:               config.SimilarityStrengthWeight,
		SupportWeight:                          config.SupportWeight,
		SimilarityConsistencyWeight:            config.SimilarityConsistencyWeight,
		AnchorProximityWeight:                  config.AnchorProximityWeight,
		ContinuationAgreementWeight:            config.ContinuationAgreementWeight,
	}
}

func normalizeLegacyFloat(canonical *float64, legacy float64) error {
	if *canonical > 0 && legacy > 0 &&
		math.Abs(*canonical-legacy) > scoreComparisonTolerance {
		return ErrLegacyConfigConflict
	}
	if *canonical == 0 {
		*canonical = legacy
	}
	return nil
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

func positiveUnitInterval(value float64) bool {
	return finite(value) && value > 0 && value <= 1
}
