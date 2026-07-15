package projectioncontinuation

import (
	"errors"
	"math"
	"testing"
)

func TestConfigValidateAcceptsExplicitPolicy(
	t *testing.T,
) {
	config := validContinuationConfig(t)

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
			name: "neighbor selector",
			mutate: func(config *Config) {
				config.NeighborSelector = nil
			},
			wantError: ErrNeighborSelectorRequired,
		},
		{
			name: "pattern evaluator",
			mutate: func(config *Config) {
				config.PatternConfidenceEvaluator = nil
			},
			wantError: ErrPatternConfidenceEvaluatorRequired,
		},
		{
			name: "fallback projector",
			mutate: func(config *Config) {
				config.FallbackProjector = nil
			},
			wantError: ErrFallbackProjectorRequired,
		},
		{
			name: "point support",
			mutate: func(config *Config) {
				config.MinimumPointSupport = 0
			},
			wantError: ErrMinimumPointSupportInvalid,
		},
		{
			name: "altitude support",
			mutate: func(config *Config) {
				config.MinimumAltitudeSupport =
					config.MinimumPointSupport + 1
			},
			wantError: ErrMinimumAltitudeSupportInvalid,
		},
		{
			name: "horizontal uncertainty",
			mutate: func(config *Config) {
				config.
					InitialHorizontalUncertaintyM = 0
			},
			wantError: ErrHorizontalUncertaintyInvalid,
		},
		{
			name: "vertical uncertainty",
			mutate: func(config *Config) {
				config.
					VerticalUncertaintyGrowthMPS = -1
			},
			wantError: ErrVerticalUncertaintyInvalid,
		},
		{
			name: "spread multiplier",
			mutate: func(config *Config) {
				config.NeighborSpreadMultiplier =
					math.NaN()
			},
			wantError: ErrNeighborSpreadMultiplierInvalid,
		},
		{
			name: "confidence loss",
			mutate: func(config *Config) {
				config.MaximumConfidenceLoss = 2
			},
			wantError: ErrMaximumConfidenceLossInvalid,
		},
		{
			name: "confidence threshold",
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
				config :=
					validContinuationConfig(t)
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
