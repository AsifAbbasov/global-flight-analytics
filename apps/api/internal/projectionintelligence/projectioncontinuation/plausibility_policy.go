package projectioncontinuation

import (
	"errors"
	"fmt"
	"time"
)

const (
	DefaultMaximumInterpolationGap   = 5 * time.Minute
	DefaultMaximumHorizontalSpeedMPS = 400.0
	DefaultMaximumVerticalSpeedMPS   = 100.0
)

var ErrPlausibilityPolicyInvalid = errors.New(
	"historical continuation plausibility policy is invalid",
)

type PlausibilityPolicy struct {
	MaximumInterpolationGap   time.Duration
	MaximumHorizontalSpeedMPS float64
	MaximumVerticalSpeedMPS   float64
}

func DefaultPlausibilityPolicy() PlausibilityPolicy {
	return PlausibilityPolicy{
		MaximumInterpolationGap:   DefaultMaximumInterpolationGap,
		MaximumHorizontalSpeedMPS: DefaultMaximumHorizontalSpeedMPS,
		MaximumVerticalSpeedMPS:   DefaultMaximumVerticalSpeedMPS,
	}
}

func (
	policy PlausibilityPolicy,
) Validate() error {
	if policy.MaximumInterpolationGap <= 0 ||
		!positiveFinite(
			policy.MaximumHorizontalSpeedMPS,
		) ||
		!positiveFinite(
			policy.MaximumVerticalSpeedMPS,
		) {
		return fmt.Errorf(
			"%w: gap=%s horizontal_mps=%f vertical_mps=%f",
			ErrPlausibilityPolicyInvalid,
			policy.MaximumInterpolationGap,
			policy.MaximumHorizontalSpeedMPS,
			policy.MaximumVerticalSpeedMPS,
		)
	}

	return nil
}

func (
	policy PlausibilityPolicy,
) isZero() bool {
	return policy.MaximumInterpolationGap == 0 &&
		policy.MaximumHorizontalSpeedMPS == 0 &&
		policy.MaximumVerticalSpeedMPS == 0
}

func (
	config Config,
) normalized() Config {
	if config.PlausibilityPolicy.isZero() {
		config.PlausibilityPolicy =
			DefaultPlausibilityPolicy()
	}

	return config
}
