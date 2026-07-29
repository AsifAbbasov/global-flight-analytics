package projectionpatternconfidence

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestConfigValidateAcceptsContinuationPolicy(t *testing.T) {
	config := validConfidenceConfig()
	if err := config.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidateMigratesVersionThreeWeights(t *testing.T) {
	config := validConfidenceConfig()
	config.SimilarityStrengthWeight = 0.35
	config.SupportWeight = 0.25
	config.SimilarityConsistencyWeight = 0.25
	config.AnchorProximityWeight = 0.15
	config.ContinuationAgreementWeight = 0

	evaluator, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if math.Abs(evaluator.config.SimilarityStrengthWeight-0.28) > scoreComparisonTolerance ||
		math.Abs(evaluator.config.SupportWeight-0.20) > scoreComparisonTolerance ||
		math.Abs(evaluator.config.SimilarityConsistencyWeight-0.20) > scoreComparisonTolerance ||
		math.Abs(evaluator.config.AnchorProximityWeight-0.12) > scoreComparisonTolerance ||
		math.Abs(evaluator.config.ContinuationAgreementWeight-0.20) > scoreComparisonTolerance {
		t.Fatalf("version three weights were not migrated: %#v", evaluator.config)
	}
}

func TestConfigValidateAcceptsLegacyAliases(t *testing.T) {
	config := validConfidenceConfig()
	config.AnchorDistanceNormalizationKM = 0
	config.SimilarityStrengthWeight = 0
	config.SimilarityConsistencyWeight = 0
	config.MaximumMeanAnchorDistanceKM = 50
	config.SimilarityWeight = 0.28
	config.FreshnessWeight = 0.20

	evaluator, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if evaluator.config.AnchorDistanceNormalizationKM != 50 ||
		evaluator.config.SimilarityStrengthWeight != 0.28 ||
		evaluator.config.SimilarityConsistencyWeight != 0.20 {
		t.Fatalf("legacy config was not normalized: %#v", evaluator.config)
	}
}

func TestConfigValidateRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Config)
		wantError error
	}{
		{
			name: "minimum neighbors",
			mutate: func(config *Config) {
				config.MinimumNeighborCount = 1
			},
			wantError: ErrMinimumNeighborCountInvalid,
		},
		{
			name: "target neighbors",
			mutate: func(config *Config) {
				config.TargetNeighborCount = 1
			},
			wantError: ErrTargetNeighborCountInvalid,
		},
		{
			name: "minimum similarity",
			mutate: func(config *Config) {
				config.MinimumSimilarityScore = 2
			},
			wantError: ErrMinimumSimilarityScoreInvalid,
		},
		{
			name: "similarity deviation",
			mutate: func(config *Config) {
				config.MaximumSimilarityStandardDeviation = 0.6
			},
			wantError: ErrMaximumSimilarityStandardDeviationInvalid,
		},
		{
			name: "anchor normalization",
			mutate: func(config *Config) {
				config.AnchorDistanceNormalizationKM = math.NaN()
			},
			wantError: ErrAnchorDistanceNormalizationInvalid,
		},
		{
			name: "minimum usable score",
			mutate: func(config *Config) {
				config.MinimumUsableScore = 0
			},
			wantError: ErrMinimumUsableScoreInvalid,
		},
		{
			name: "confidence thresholds",
			mutate: func(config *Config) {
				config.MediumConfidenceMinimum = 0.9
				config.HighConfidenceMinimum = 0.8
			},
			wantError: ErrConfidenceThresholdInvalid,
		},
		{
			name: "continuation samples",
			mutate: func(config *Config) {
				config.ContinuationAgreementSampleCount = 33
			},
			wantError: ErrContinuationAgreementSampleCountInvalid,
		},
		{
			name: "continuation divergence",
			mutate: func(config *Config) {
				config.MaximumContinuationDivergenceMPS = 10
			},
			wantError: ErrContinuationDivergencePolicyInvalid,
		},
		{
			name: "component weight",
			mutate: func(config *Config) {
				config.SimilarityStrengthWeight = 0
				config.SimilarityWeight = 0
			},
			wantError: ErrComponentWeightInvalid,
		},
		{
			name: "legacy conflict",
			mutate: func(config *Config) {
				config.SimilarityWeight = 0.2
			},
			wantError: ErrLegacyConfigConflict,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validConfidenceConfig()
			test.mutate(&config)
			err := config.Validate()
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func validConfidenceConfig() Config {
	return Config{
		MinimumNeighborCount:                   2,
		TargetNeighborCount:                    3,
		MinimumSimilarityScore:                 0.55,
		MaximumSimilarityStandardDeviation:     0.15,
		AnchorDistanceNormalizationKM:          50,
		MinimumUsableScore:                     0.5,
		MediumConfidenceMinimum:                0.6,
		HighConfidenceMinimum:                  0.8,
		ContinuationAgreementSampleCount:       4,
		ContinuationDivergenceNormalizationMPS: 25,
		MaximumContinuationDivergenceMPS:       50,
		SimilarityStrengthWeight:               0.28,
		SupportWeight:                          0.20,
		SimilarityConsistencyWeight:            0.20,
		AnchorProximityWeight:                  0.12,
		ContinuationAgreementWeight:            0.20,
		MaximumCandidateAge:                    7 * 24 * time.Hour,
	}
}
