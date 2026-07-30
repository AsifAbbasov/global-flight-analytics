package projectionfreshness

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestConfigValidateAcceptsExplicitPolicy(
	t *testing.T,
) {
	config := validFreshnessConfig()

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
			name: "newest age",
			mutate: func(config *Config) {
				config.MaximumNewestNeighborAge = 0
			},
			wantError: ErrMaximumNewestNeighborAgeInvalid,
		},
		{
			name: "mean age",
			mutate: func(config *Config) {
				config.MaximumMeanNeighborAge = 0
			},
			wantError: ErrMaximumMeanNeighborAgeInvalid,
		},
		{
			name: "oldest age",
			mutate: func(config *Config) {
				config.MaximumOldestNeighborAge = 0
			},
			wantError: ErrMaximumOldestNeighborAgeInvalid,
		},
		{
			name: "recent age limit",
			mutate: func(config *Config) {
				config.RecentNeighborAgeLimit = 0
			},
			wantError: ErrRecentNeighborAgeLimitInvalid,
		},
		{
			name: "recent counts",
			mutate: func(config *Config) {
				config.MinimumRecentNeighborCount = 3
				config.TargetRecentNeighborCount = 2
			},
			wantError: ErrRecentNeighborCountInvalid,
		},
		{
			name: "score thresholds",
			mutate: func(config *Config) {
				config.MinimumUsableScore = 0.9
				config.CompleteScoreMinimum = 0.8
			},
			wantError: ErrFreshnessScoreThresholdInvalid,
		},
		{
			name: "weights",
			mutate: func(config *Config) {
				config.NewestAgeWeight =
					math.NaN()
			},
			wantError: ErrFreshnessWeightInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				config := validFreshnessConfig()
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

func TestConfigValidateRejectsZeroComponentWeights(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{
			name: "newest age weight",
			mutate: func(config *Config) {
				config.NewestAgeWeight = 0
				config.MeanAgeWeight = 0.40
				config.OldestAgeWeight = 0.30
				config.RecentSupportWeight = 0.30
			},
		},
		{
			name: "mean age weight",
			mutate: func(config *Config) {
				config.NewestAgeWeight = 0.40
				config.MeanAgeWeight = 0
				config.OldestAgeWeight = 0.30
				config.RecentSupportWeight = 0.30
			},
		},
		{
			name: "oldest age weight",
			mutate: func(config *Config) {
				config.NewestAgeWeight = 0.40
				config.MeanAgeWeight = 0.40
				config.OldestAgeWeight = 0
				config.RecentSupportWeight = 0.20
			},
		},
		{
			name: "recent support weight",
			mutate: func(config *Config) {
				config.NewestAgeWeight = 0.40
				config.MeanAgeWeight = 0.40
				config.OldestAgeWeight = 0.20
				config.RecentSupportWeight = 0
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validFreshnessConfig()
			test.mutate(&config)
			err := config.Validate()
			if !errors.Is(err, ErrFreshnessWeightInvalid) {
				t.Fatalf(
					"error = %v, want %v",
					err,
					ErrFreshnessWeightInvalid,
				)
			}
		})
	}
}

func validFreshnessConfig() Config {
	return Config{
		MaximumNewestNeighborAge: 7 * 24 * time.Hour,
		MaximumMeanNeighborAge:   14 * 24 * time.Hour,
		MaximumOldestNeighborAge: 30 * 24 * time.Hour,

		RecentNeighborAgeLimit:     10 * 24 * time.Hour,
		MinimumRecentNeighborCount: 2,
		TargetRecentNeighborCount:  3,

		MinimumUsableScore:   0.35,
		CompleteScoreMinimum: 0.65,

		NewestAgeWeight:     0.30,
		MeanAgeWeight:       0.30,
		OldestAgeWeight:     0.20,
		RecentSupportWeight: 0.20,
	}
}
