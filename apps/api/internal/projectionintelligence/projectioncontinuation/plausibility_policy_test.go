package projectioncontinuation

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestDefaultPlausibilityPolicyIsValid(
	t *testing.T,
) {
	policy := DefaultPlausibilityPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf(
			"DefaultPlausibilityPolicy().Validate() error = %v",
			err,
		)
	}
}

func TestNewNormalizesZeroPlausibilityPolicy(
	t *testing.T,
) {
	config := validContinuationConfig(t)
	config.PlausibilityPolicy =
		PlausibilityPolicy{}

	baseline, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	want := DefaultPlausibilityPolicy()
	if baseline.config.PlausibilityPolicy != want {
		t.Fatalf(
			"normalized policy = %#v, want %#v",
			baseline.config.PlausibilityPolicy,
			want,
		)
	}
}

func TestPlausibilityPolicyRejectsInvalidValues(
	t *testing.T,
) {
	tests := []struct {
		name   string
		policy PlausibilityPolicy
	}{
		{
			name: "partial policy",
			policy: PlausibilityPolicy{
				MaximumInterpolationGap: time.Minute,
			},
		},
		{
			name: "negative gap",
			policy: PlausibilityPolicy{
				MaximumInterpolationGap:   -time.Second,
				MaximumHorizontalSpeedMPS: 400,
				MaximumVerticalSpeedMPS:   100,
			},
		},
		{
			name: "non-finite horizontal speed",
			policy: PlausibilityPolicy{
				MaximumInterpolationGap:   time.Minute,
				MaximumHorizontalSpeedMPS: math.NaN(),
				MaximumVerticalSpeedMPS:   100,
			},
		},
		{
			name: "zero vertical speed",
			policy: PlausibilityPolicy{
				MaximumInterpolationGap:   time.Minute,
				MaximumHorizontalSpeedMPS: 400,
				MaximumVerticalSpeedMPS:   0,
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				if err := test.policy.Validate(); !errors.Is(
					err,
					ErrPlausibilityPolicyInvalid,
				) {
					t.Fatalf(
						"Validate() error = %v",
						err,
					)
				}
			},
		)
	}
}
