package confidencereport

import (
	"errors"
	"math"
	"testing"
)

func TestDefaultConfigIsValid(
	t *testing.T,
) {
	if err := DefaultConfig().Validate(); err != nil {
		t.Fatalf(
			"expected valid default config, got %v",
			err,
		)
	}
}

func TestConfigValidationRejectsInvalidValues(
	t *testing.T,
) {
	testCases := []struct {
		name     string
		config   Config
		expected error
	}{
		{
			name: "medium threshold not finite",
			config: Config{
				MediumThreshold:  math.NaN(),
				HighThreshold:    0.80,
				MaximumPenalty:   1,
				DecimalPrecision: 6,
			},
			expected: ErrMediumThresholdInvalid,
		},
		{
			name: "medium threshold not below high",
			config: Config{
				MediumThreshold:  0.80,
				HighThreshold:    0.80,
				MaximumPenalty:   1,
				DecimalPrecision: 6,
			},
			expected: ErrMediumThresholdInvalid,
		},
		{
			name: "high threshold above one",
			config: Config{
				MediumThreshold:  0.60,
				HighThreshold:    1.01,
				MaximumPenalty:   1,
				DecimalPrecision: 6,
			},
			expected: ErrHighThresholdInvalid,
		},
		{
			name: "maximum penalty below zero",
			config: Config{
				MediumThreshold:  0.60,
				HighThreshold:    0.80,
				MaximumPenalty:   -0.01,
				DecimalPrecision: 6,
			},
			expected: ErrMaximumPenaltyInvalid,
		},
		{
			name: "precision above twelve",
			config: Config{
				MediumThreshold:  0.60,
				HighThreshold:    0.80,
				MaximumPenalty:   1,
				DecimalPrecision: 13,
			},
			expected: ErrDecimalPrecisionInvalid,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				err := testCase.config.Validate()

				if !errors.Is(
					err,
					testCase.expected,
				) {
					t.Fatalf(
						"expected %v, got %v",
						testCase.expected,
						err,
					)
				}
			},
		)
	}
}
