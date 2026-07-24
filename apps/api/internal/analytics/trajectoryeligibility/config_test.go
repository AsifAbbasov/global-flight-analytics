package trajectoryeligibility

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestDefaultConfigIsValid(
	t *testing.T,
) {
	if err := DefaultConfig().Validate(); err != nil {
		t.Fatalf(
			"expected default config to be valid, got %v",
			err,
		)
	}
}

func TestPolicyValidationRejectsInvalidValues(
	t *testing.T,
) {
	testCases := []struct {
		name     string
		policy   Policy
		expected error
	}{
		{
			name: "negative minimum point count",
			policy: Policy{
				MinimumPointCount: -1,
			},
			expected: ErrMinimumPointCountInvalid,
		},
		{
			name: "quality score below zero",
			policy: Policy{
				MinimumQualityScore: -0.01,
			},
			expected: ErrMinimumQualityScoreInvalid,
		},
		{
			name: "quality score above one",
			policy: Policy{
				MinimumQualityScore: 1.01,
			},
			expected: ErrMinimumQualityScoreInvalid,
		},
		{
			name: "quality score not a number",
			policy: Policy{
				MinimumQualityScore: math.NaN(),
			},
			expected: ErrMinimumQualityScoreInvalid,
		},
		{
			name: "coverage gap count below unlimited sentinel",
			policy: Policy{
				MaximumCoverageGapCount: -2,
			},
			expected: ErrMaximumCoverageGapCountInvalid,
		},
		{
			name: "negative minimum duration",
			policy: Policy{
				MinimumDuration: -time.Second,
			},
			expected: ErrMinimumDurationInvalid,
		},
		{
			name: "negative maximum duration",
			policy: Policy{
				MaximumDuration: -time.Second,
			},
			expected: ErrMaximumDurationInvalid,
		},
		{
			name: "maximum duration below minimum",
			policy: Policy{
				MinimumDuration: 2 * time.Minute,
				MaximumDuration: time.Minute,
			},
			expected: ErrDurationRangeInvalid,
		},
		{
			name: "negative maximum observation age",
			policy: Policy{
				MaximumObservationAge: -time.Second,
			},
			expected: ErrMaximumObservationAgeInvalid,
		},
		{
			name: "negative maximum future observation skew",
			policy: Policy{
				MaximumFutureObservationSkew: -time.Second,
			},
			expected: ErrMaximumFutureObservationSkewInvalid,
		},
		{
			name: "negative maximum recent point gap",
			policy: Policy{
				MaximumRecentPointGap: -time.Second,
			},
			expected: ErrMaximumRecentPointGapInvalid,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				err := testCase.policy.Validate()

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

func TestConfigValidationIdentifiesCapability(
	t *testing.T,
) {
	config := DefaultConfig()
	config.RouteInference.MinimumPointCount = -1

	err := config.Validate()
	if err == nil {
		t.Fatal("expected invalid route inference config")
	}

	if !errors.Is(
		err,
		ErrMinimumPointCountInvalid,
	) {
		t.Fatalf(
			"expected minimum point count error, got %v",
			err,
		)
	}
}
