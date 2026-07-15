package projectionneighbors

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestConfigValidateAcceptsExplicitPolicy(
	t *testing.T,
) {
	config := validSelectorConfig()

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
			name: "similarity engine",
			mutate: func(config *Config) {
				config.SimilarityEngine = nil
			},
			wantError: ErrSimilarityEngineRequired,
		},
		{
			name: "similarity policy key",
			mutate: func(config *Config) {
				config.SimilarityPolicyKey = ""
			},
			wantError: ErrSimilarityPolicyKeyRequired,
		},
		{
			name: "minimum points",
			mutate: func(config *Config) {
				config.MinimumCurrentPointCount = 1
			},
			wantError: ErrMinimumCurrentPointCountInvalid,
		},
		{
			name: "maximum candidates",
			mutate: func(config *Config) {
				config.MaximumCandidateCount = 0
			},
			wantError: ErrMaximumCandidateCountInvalid,
		},
		{
			name: "selection limit",
			mutate: func(config *Config) {
				config.SelectionLimit = 0
			},
			wantError: ErrSelectionLimitInvalid,
		},
		{
			name: "similarity score",
			mutate: func(config *Config) {
				config.MinimumSimilarityScore =
					math.NaN()
			},
			wantError: ErrMinimumSimilarityScoreInvalid,
		},
		{
			name: "anchor distance",
			mutate: func(config *Config) {
				config.MaximumAnchorDistanceKM = 0
			},
			wantError: ErrMaximumAnchorDistanceInvalid,
		},
		{
			name: "candidate age",
			mutate: func(config *Config) {
				config.MaximumCandidateAge =
					-time.Hour
			},
			wantError: ErrMaximumCandidateAgeInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				config := validSelectorConfig()
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
