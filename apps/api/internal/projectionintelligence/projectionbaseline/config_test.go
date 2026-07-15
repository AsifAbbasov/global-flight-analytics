package projectionbaseline

import (
	"errors"
	"math"
	"testing"
)

func TestConfigValidateAcceptsExplicitPolicy(
	t *testing.T,
) {
	config := validBaselineConfig()

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
			name: "horizon planner",
			mutate: func(config *Config) {
				config.HorizonPlanner = nil
			},
			wantError: ErrHorizonPlannerRequired,
		},
		{
			name: "eligibility evaluator",
			mutate: func(config *Config) {
				config.EligibilityEvaluator = nil
			},
			wantError: ErrEligibilityEvaluatorRequired,
		},
		{
			name: "horizontal initial",
			mutate: func(config *Config) {
				config.
					InitialHorizontalUncertaintyM = 0
			},
			wantError: ErrHorizontalUncertaintyInvalid,
		},
		{
			name: "horizontal growth",
			mutate: func(config *Config) {
				config.
					HorizontalUncertaintyGrowthMPS = -1
			},
			wantError: ErrHorizontalUncertaintyInvalid,
		},
		{
			name: "vertical initial",
			mutate: func(config *Config) {
				config.
					InitialVerticalUncertaintyM =
					math.NaN()
			},
			wantError: ErrVerticalUncertaintyInvalid,
		},
		{
			name: "vertical growth",
			mutate: func(config *Config) {
				config.
					VerticalUncertaintyGrowthMPS = -1
			},
			wantError: ErrVerticalUncertaintyInvalid,
		},
		{
			name: "confidence loss",
			mutate: func(config *Config) {
				config.MaximumConfidenceLoss = 2
			},
			wantError: ErrMaximumConfidenceLossInvalid,
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
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				config := validBaselineConfig()
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
