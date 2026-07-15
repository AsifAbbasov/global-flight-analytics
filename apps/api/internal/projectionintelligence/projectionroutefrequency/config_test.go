package projectionroutefrequency

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestConfigValidateAcceptsExplicitPolicy(
	t *testing.T,
) {
	config := validRouteFrequencyConfig()

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
			name: "observation counts",
			mutate: func(config *Config) {
				config.MinimumObservationCount = 5
				config.TargetObservationCount = 4
			},
			wantError: ErrObservationCountInvalid,
		},
		{
			name: "distinct days",
			mutate: func(config *Config) {
				config.MinimumDistinctDayCount = 3
				config.TargetDistinctDayCount = 2
			},
			wantError: ErrDistinctDayCountInvalid,
		},
		{
			name: "recent window",
			mutate: func(config *Config) {
				config.RecentWindow = 0
			},
			wantError: ErrRecentWindowInvalid,
		},
		{
			name: "recent observations",
			mutate: func(config *Config) {
				config.MinimumRecentObservationCount = 4
				config.TargetRecentObservationCount = 3
			},
			wantError: ErrRecentObservationCountInvalid,
		},
		{
			name: "latest observation age",
			mutate: func(config *Config) {
				config.MaximumLatestObservationAge = 0
			},
			wantError: ErrMaximumLatestObservationAgeInvalid,
		},
		{
			name: "route confidence",
			mutate: func(config *Config) {
				config.MinimumRouteConfidenceScore = 2
			},
			wantError: ErrMinimumRouteConfidenceInvalid,
		},
		{
			name: "score thresholds",
			mutate: func(config *Config) {
				config.MinimumUsableScore = 0.9
				config.CompleteScoreMinimum = 0.8
			},
			wantError: ErrScoreThresholdInvalid,
		},
		{
			name: "weights",
			mutate: func(config *Config) {
				config.ObservationCountWeight =
					math.NaN()
			},
			wantError: ErrComponentWeightInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				config :=
					validRouteFrequencyConfig()
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

func validRouteFrequencyConfig() Config {
	return Config{
		MinimumObservationCount: 5,
		TargetObservationCount:  10,

		MinimumDistinctDayCount: 3,
		TargetDistinctDayCount:  6,

		RecentWindow:                  14 * 24 * time.Hour,
		MinimumRecentObservationCount: 2,
		TargetRecentObservationCount:  4,

		MaximumLatestObservationAge: 7 * 24 * time.Hour,
		MinimumRouteConfidenceScore: 0.6,

		MinimumUsableScore:   0.45,
		CompleteScoreMinimum: 0.75,

		ObservationCountWeight:  0.25,
		DistinctDayWeight:       0.20,
		RecentObservationWeight: 0.20,
		LatestObservationWeight: 0.20,
		RouteConfidenceWeight:   0.15,
	}
}
