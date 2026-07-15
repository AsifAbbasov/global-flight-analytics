package projectionevaluation

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestConfigValidateAcceptsExplicitPolicy(
	t *testing.T,
) {
	config := validEvaluationConfig()

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
			name: "interpolation gap",
			mutate: func(config *Config) {
				config.MaximumInterpolationGap = 0
			},
			wantError: ErrMaximumInterpolationGapInvalid,
		},
		{
			name: "minimum evaluated points",
			mutate: func(config *Config) {
				config.MinimumEvaluatedPointCount = 0
			},
			wantError: ErrMinimumEvaluatedPointCountInvalid,
		},
		{
			name: "maximum horizontal error",
			mutate: func(config *Config) {
				config.MaximumHorizontalErrorM =
					math.NaN()
			},
			wantError: ErrMaximumHorizontalErrorInvalid,
		},
		{
			name: "maximum altitude error",
			mutate: func(config *Config) {
				config.MaximumAltitudeErrorM = 0
			},
			wantError: ErrMaximumAltitudeErrorInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				config := validEvaluationConfig()
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

func validEvaluationConfig() Config {
	return Config{
		MaximumInterpolationGap:    3 * time.Minute,
		MinimumEvaluatedPointCount: 1,
		MaximumHorizontalErrorM:    10000,
		MaximumAltitudeErrorM:      1000,
	}
}
