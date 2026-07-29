package projectionpatternconfidence

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestConfigValidateAcceptsExplicitDistributionPolicy(t *testing.T) {
	config := validConfidenceConfig()
	if err := config.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidateAcceptsLegacyWeightAliases(t *testing.T) {
	config := validConfidenceConfig()
	config.AnchorDistanceNormalizationKM = 0
	config.SimilarityStrengthWeight = 0
	config.SimilarityConsistencyWeight = 0
	config.MaximumMeanAnchorDistanceKM = 50
	config.SimilarityWeight = 0.35
	config.FreshnessWeight = 0.25

	evaluator, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if evaluator.config.AnchorDistanceNormalizationKM != 50 ||
		evaluator.config.SimilarityStrengthWeight != 0.35 ||
		evaluator.config.SimilarityConsistencyWeight != 0.25 {
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
				config.MinimumNeighborCount = 0
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
			name: "minimum similarity above one",
			mutate: func(config *Config) {
				config.MinimumSimilarityScore = 2
			},
			wantError: ErrMinimumSimilarityScoreInvalid,
		},
		{
			name: "similarity standard deviation above theoretical maximum",
			mutate: func(config *Config) {
				config.MaximumSimilarityStandardDeviation = 0.6
			},
			wantError: ErrMaximumSimilarityStandardDeviationInvalid,
		},
		{
			name: "anchor distance normalization",
			mutate: func(config *Config) {
				config.AnchorDistanceNormalizationKM = math.NaN()
			},
			wantError: ErrAnchorDistanceNormalizationInvalid,
		},
		{
			name: "minimum usable score above one",
			mutate: func(config *Config) {
				config.MinimumUsableScore = 2
			},
			wantError: ErrMinimumUsableScoreInvalid,
		},
		{
			name: "minimum usable score is zero",
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
			name: "component weights do not sum to one",
			mutate: func(config *Config) {
				config.SimilarityStrengthWeight = 1
			},
			wantError: ErrComponentWeightInvalid,
		},
		{
			name: "component weight is zero",
			mutate: func(config *Config) {
				config.SimilarityStrengthWeight = 0
				config.SimilarityWeight = 0
			},
			wantError: ErrComponentWeightInvalid,
		},
		{
			name: "legacy canonical conflict",
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
		MinimumNeighborCount:               2,
		TargetNeighborCount:                3,
		MinimumSimilarityScore:             0.55,
		MaximumSimilarityStandardDeviation: 0.15,
		AnchorDistanceNormalizationKM:      50,
		MinimumUsableScore:                 0.5,
		MediumConfidenceMinimum:            0.6,
		HighConfidenceMinimum:              0.8,
		SimilarityStrengthWeight:           0.35,
		SupportWeight:                      0.25,
		SimilarityConsistencyWeight:        0.25,
		AnchorProximityWeight:              0.15,
		MaximumCandidateAge:                7 * 24 * time.Hour,
	}
}
