package projectionhorizon

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

var (
	ErrPlanInvalid = errors.New(
		"projection horizon plan is invalid",
	)
	ErrPlanFingerprintInvalid = errors.New(
		"projection horizon plan fingerprint is invalid",
	)
)

type Plan struct {
	Version    string
	PolicyName string

	AsOfTime time.Time
	EndTime  time.Time
	Step     time.Duration

	RequestedDuration time.Duration
	EffectiveDuration time.Duration

	ForecastTimes []time.Time

	Truncated        bool
	TruncationReason TruncationReason

	Fingerprint string
}

func (plan Plan) Clone() Plan {
	cloned := plan
	cloned.ForecastTimes = append(
		[]time.Time(nil),
		plan.ForecastTimes...,
	)

	return cloned
}

func (plan Plan) ContractHorizon() projectioncontract.Horizon {
	return projectioncontract.Horizon{
		AsOfTime: plan.AsOfTime,
		EndTime:  plan.EndTime,
		Step:     plan.Step,
	}
}

func FinalizePlan(plan Plan) (Plan, error) {
	plan.Fingerprint = calculatePlanFingerprint(plan)
	if err := plan.Validate(); err != nil {
		return Plan{}, err
	}

	return plan.Clone(), nil
}

func (plan Plan) Validate() error {
	if plan.Version != Version {
		return invalidPlan(
			"version must be %q",
			Version,
		)
	}
	if normalized := strings.TrimSpace(plan.PolicyName); normalized == "" ||
		normalized != plan.PolicyName {
		return invalidPlan(
			"policy name is required and must be normalized",
		)
	}
	if plan.AsOfTime.IsZero() || plan.AsOfTime.Location() != time.UTC {
		return invalidPlan(
			"as-of time is required and must be canonical UTC",
		)
	}
	if plan.EndTime.IsZero() || plan.EndTime.Location() != time.UTC ||
		!plan.EndTime.After(plan.AsOfTime) {
		return invalidPlan(
			"end time must be canonical UTC and after the as-of time",
		)
	}
	if plan.Step <= 0 {
		return invalidPlan("step must be greater than zero")
	}
	if plan.RequestedDuration <= 0 || plan.EffectiveDuration <= 0 {
		return invalidPlan(
			"requested and effective durations must be greater than zero",
		)
	}
	if plan.EffectiveDuration%plan.Step != 0 {
		return invalidPlan(
			"effective duration %s is not divisible by step %s",
			plan.EffectiveDuration,
			plan.Step,
		)
	}
	if plan.Step > plan.EffectiveDuration {
		return invalidPlan(
			"step %s exceeds effective duration %s",
			plan.Step,
			plan.EffectiveDuration,
		)
	}
	if !plan.EndTime.Equal(
		plan.AsOfTime.Add(plan.EffectiveDuration),
	) {
		return invalidPlan(
			"end time does not match as-of time plus effective duration",
		)
	}

	if plan.Truncated {
		if plan.RequestedDuration <= plan.EffectiveDuration ||
			plan.TruncationReason != TruncationReasonMaximumDuration {
			return invalidPlan(
				"truncated plan requires a longer requested duration and maximum-duration reason",
			)
		}
	} else if plan.RequestedDuration != plan.EffectiveDuration ||
		plan.TruncationReason != TruncationReasonNone {
		return invalidPlan(
			"non-truncated plan requires equal requested and effective durations with no reason",
		)
	}

	expectedPointCount := exactForecastPointCount(
		plan.EffectiveDuration,
		plan.Step,
	)
	if expectedPointCount < 1 ||
		expectedPointCount > MaximumSupportedPointCount ||
		len(plan.ForecastTimes) != expectedPointCount {
		return invalidPlan(
			"forecast point count is %d, expected %d within supported limit %d",
			len(plan.ForecastTimes),
			expectedPointCount,
			MaximumSupportedPointCount,
		)
	}
	for index, forecastTime := range plan.ForecastTimes {
		if forecastTime.Location() != time.UTC {
			return invalidPlan(
				"forecast time at index %d is not canonical UTC",
				index,
			)
		}
		expectedTime := plan.AsOfTime.Add(
			time.Duration(index+1) * plan.Step,
		)
		if !forecastTime.Equal(expectedTime) {
			return invalidPlan(
				"forecast time at index %d does not match the fixed step grid",
				index,
			)
		}
	}
	if !plan.ForecastTimes[len(plan.ForecastTimes)-1].Equal(
		plan.EndTime,
	) {
		return invalidPlan(
			"last forecast time does not match the end time",
		)
	}

	expectedFingerprint := calculatePlanFingerprint(plan)
	if plan.Fingerprint == "" || plan.Fingerprint != expectedFingerprint {
		return fmt.Errorf(
			"%w: %w",
			ErrPlanInvalid,
			ErrPlanFingerprintInvalid,
		)
	}

	return nil
}

func invalidPlan(format string, args ...any) error {
	return fmt.Errorf(
		"%w: %s",
		ErrPlanInvalid,
		fmt.Sprintf(format, args...),
	)
}
