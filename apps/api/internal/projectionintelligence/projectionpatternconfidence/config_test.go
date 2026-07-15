package projectionpatternconfidence

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestConfigValidateAcceptsExplicitPolicy(
	t *testing.T,
) {
	config := validConfidenceConfig()

	if err := config.Validate(); err != nil {
		t.Fatalf(
			"Validate() error = %v",
			err,
		)
	}
}

func TestConfigValidateRejectsInvalidValues(
	t *testing.T,
) {
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
			name: "maximum age",
			mutate: func(config *Config) {
				config.MaximumCandidateAge = 0
			},
			wantError: ErrMaximumCandidateAgeInvalid,
		},
		{
			name: "anchor distance",
			mutate: func(config *Config) {
				config.MaximumMeanAnchorDistanceKM =
					math.NaN()
			},
			wantError: ErrMaximumMeanAnchorDistanceInvalid,
		},
		{
			name: "minimum usable score",
			mutate: func(config *Config) {
				config.MinimumUsableScore = 2
			},
			wantError: ErrMinimumUsableScoreInvalid,
		},
		{
			name: "confidence thresholds",
			mutate: func(config *Config) {
				config.MediumConfidenceMinimum =
					0.9
				config.HighConfidenceMinimum =
					0.8
			},
			wantError: ErrConfidenceThresholdInvalid,
		},
		{
			name: "component weights",
			mutate: func(config *Config) {
				config.SimilarityWeight = 1
			},
			wantError: ErrComponentWeightInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				config := validConfidenceConfig()
				test.mutate(&config)

				err := config.Validate()
				if !errors.Is(
					err,
					test.wantError,
				) {
					t.Fatalf(
						"error = %v, want %v",
						err,
						test.wantError,
					)
				}
			},
		)
	}
}

func validConfidenceConfig() Config {
	return Config{
		MinimumNeighborCount: 2,
		TargetNeighborCount:  3,

		MaximumCandidateAge:         7 * 24 * time.Hour,
		MaximumMeanAnchorDistanceKM: 50,

		MinimumUsableScore: 0.5,

		MediumConfidenceMinimum: 0.6,
		HighConfidenceMinimum:   0.8,

		SimilarityWeight:      0.4,
		SupportWeight:         0.3,
		FreshnessWeight:       0.2,
		AnchorProximityWeight: 0.1,
	}
}
