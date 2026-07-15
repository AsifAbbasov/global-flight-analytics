package projectionarrival

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestConfigValidateAcceptsExplicitPolicy(
	t *testing.T,
) {
	config := validArrivalConfig()

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
			name: "arrival radius",
			mutate: func(config *Config) {
				config.ArrivalRadiusM = 0
			},
			wantError: ErrArrivalRadiusInvalid,
		},
		{
			name: "destination confidence",
			mutate: func(config *Config) {
				config.
					MinimumDestinationConfidenceScore = 2
			},
			wantError: ErrDestinationConfidenceInvalid,
		},
		{
			name: "speed samples",
			mutate: func(config *Config) {
				config.MinimumSpeedSampleCount = 3
				config.MaximumSpeedSampleCount = 2
			},
			wantError: ErrSpeedSamplePolicyInvalid,
		},
		{
			name: "minimum speed",
			mutate: func(config *Config) {
				config.MinimumGroundSpeedMPS =
					math.NaN()
			},
			wantError: ErrMinimumGroundSpeedInvalid,
		},
		{
			name: "speed uncertainty",
			mutate: func(config *Config) {
				config.SpeedUncertaintyMultiplier = 0
			},
			wantError: ErrSpeedUncertaintyMultiplierInvalid,
		},
		{
			name: "arrival interval",
			mutate: func(config *Config) {
				config.MinimumArrivalInterval = 0
			},
			wantError: ErrMinimumArrivalIntervalInvalid,
		},
		{
			name: "maximum duration",
			mutate: func(config *Config) {
				config.
					MaximumEstimatedArrivalDuration = 0
			},
			wantError: ErrMaximumArrivalDurationInvalid,
		},
		{
			name: "extrapolation loss",
			mutate: func(config *Config) {
				config.
					MaximumExtrapolationConfidenceLoss = -1
			},
			wantError: ErrExtrapolationConfidenceLossInvalid,
		},
		{
			name: "confidence weights",
			mutate: func(config *Config) {
				config.ProjectionConfidenceWeight = 1
			},
			wantError: ErrConfidenceWeightInvalid,
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
				config := validArrivalConfig()
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

func validArrivalConfig() Config {
	return Config{
		ArrivalRadiusM: 1000,

		MinimumDestinationConfidenceScore: 0.6,

		MinimumSpeedSampleCount: 2,
		MaximumSpeedSampleCount: 4,
		MinimumGroundSpeedMPS:   5,

		SpeedUncertaintyMultiplier:      2,
		MinimumArrivalInterval:          2 * time.Minute,
		MaximumEstimatedArrivalDuration: 2 * time.Hour,

		MaximumExtrapolationConfidenceLoss: 0.5,

		ProjectionConfidenceWeight:  0.4,
		DestinationConfidenceWeight: 0.4,
		SpeedStabilityWeight:        0.2,

		MediumConfidenceMinimum: 0.6,
		HighConfidenceMinimum:   0.8,
	}
}
