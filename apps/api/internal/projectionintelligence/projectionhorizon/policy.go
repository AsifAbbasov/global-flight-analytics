package projectionhorizon

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
)

const Version = "projection-horizon-policy-v1"

var (
	ErrMinimumDurationInvalid = errors.New(
		"minimum projection duration must be greater than zero",
	)
	ErrDefaultDurationInvalid = errors.New(
		"default projection duration must be inside the configured duration range",
	)
	ErrMaximumDurationInvalid = errors.New(
		"maximum projection duration must not be below the minimum duration",
	)
	ErrStepInvalid = errors.New(
		"projection step must be greater than zero",
	)
	ErrMaximumPointCountInvalid = errors.New(
		"maximum projection point count is invalid",
	)
	ErrAsOfTimeRequired = errors.New(
		"projection as-of time is required",
	)
	ErrRequestedDurationBelowMinimum = errors.New(
		"requested projection duration is below the configured minimum",
	)
)

type Config struct {
	Name string

	MinimumDuration time.Duration
	DefaultDuration time.Duration
	MaximumDuration time.Duration
	Step            time.Duration

	MaximumPointCount int
}

type Policy struct {
	config Config
}

func New(
	config Config,
) (*Policy, error) {
	if config.MinimumDuration <= 0 {
		return nil,
			ErrMinimumDurationInvalid
	}
	if config.MaximumDuration <
		config.MinimumDuration {
		return nil,
			ErrMaximumDurationInvalid
	}
	if config.DefaultDuration <
		config.MinimumDuration ||
		config.DefaultDuration >
			config.MaximumDuration {
		return nil,
			ErrDefaultDurationInvalid
	}
	if config.Step <= 0 {
		return nil,
			ErrStepInvalid
	}
	if config.MaximumPointCount < 1 {
		return nil,
			ErrMaximumPointCountInvalid
	}

	requiredPointCount := forecastPointCount(
		config.MaximumDuration,
		config.Step,
	)
	if requiredPointCount >
		config.MaximumPointCount {
		return nil,
			fmt.Errorf(
				"%w: maximum duration requires %d points, configured maximum is %d",
				ErrMaximumPointCountInvalid,
				requiredPointCount,
				config.MaximumPointCount,
			)
	}

	return &Policy{
		config: config,
	}, nil
}

func (policy *Policy) Config() Config {
	if policy == nil {
		return Config{}
	}

	return policy.config
}

type Request struct {
	AsOfTime          time.Time
	RequestedDuration time.Duration
}

type TruncationReason string

const (
	TruncationReasonNone            TruncationReason = ""
	TruncationReasonMaximumDuration TruncationReason = "maximum_duration"
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

func (policy *Policy) Build(
	request Request,
) (Plan, error) {
	if policy == nil {
		return Plan{},
			ErrMaximumPointCountInvalid
	}
	if request.AsOfTime.IsZero() {
		return Plan{},
			ErrAsOfTimeRequired
	}

	requestedDuration :=
		request.RequestedDuration
	if requestedDuration == 0 {
		requestedDuration =
			policy.config.DefaultDuration
	}
	if requestedDuration <
		policy.config.MinimumDuration {
		return Plan{},
			&DurationBelowMinimumError{
				Requested: requestedDuration,
				Minimum: policy.config.
					MinimumDuration,
			}
	}

	effectiveDuration := requestedDuration
	truncated := false
	truncationReason :=
		TruncationReasonNone
	if effectiveDuration >
		policy.config.MaximumDuration {
		effectiveDuration =
			policy.config.MaximumDuration
		truncated = true
		truncationReason =
			TruncationReasonMaximumDuration
	}

	asOfTime := request.AsOfTime.UTC()
	forecastTimes := buildForecastTimes(
		asOfTime,
		effectiveDuration,
		policy.config.Step,
	)
	if len(forecastTimes) >
		policy.config.MaximumPointCount {
		return Plan{},
			fmt.Errorf(
				"%w: planned %d points, maximum is %d",
				ErrMaximumPointCountInvalid,
				len(forecastTimes),
				policy.config.
					MaximumPointCount,
			)
	}

	return Plan{
		Version:    Version,
		PolicyName: policy.config.Name,

		AsOfTime: asOfTime,
		EndTime: asOfTime.Add(
			effectiveDuration,
		),
		Step: policy.config.Step,

		RequestedDuration: requestedDuration,
		EffectiveDuration: effectiveDuration,

		ForecastTimes: forecastTimes,

		Truncated:        truncated,
		TruncationReason: truncationReason,
	}.Clone(), nil
}

type DurationBelowMinimumError struct {
	Requested time.Duration
	Minimum   time.Duration
}

func (
	err *DurationBelowMinimumError,
) Error() string {
	return fmt.Sprintf(
		"requested projection duration %s is below minimum %s",
		err.Requested,
		err.Minimum,
	)
}

func (
	err *DurationBelowMinimumError,
) Unwrap() error {
	return ErrRequestedDurationBelowMinimum
}

func forecastPointCount(
	duration time.Duration,
	step time.Duration,
) int {
	if duration <= 0 ||
		step <= 0 {
		return 0
	}

	count := int(
		duration / step,
	)
	if duration%step != 0 {
		count++
	}

	return count
}

func buildForecastTimes(
	asOfTime time.Time,
	duration time.Duration,
	step time.Duration,
) []time.Time {
	pointCount := forecastPointCount(
		duration,
		step,
	)
	result := make(
		[]time.Time,
		0,
		pointCount,
	)
	endTime := asOfTime.Add(
		duration,
	)

	for offset := step; offset < duration; offset += step {
		result = append(
			result,
			asOfTime.Add(offset),
		)
	}
	result = append(
		result,
		endTime,
	)

	return result
}
