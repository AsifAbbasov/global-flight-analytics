package projectionhorizon

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	Version                    = "projection-horizon-policy-v2"
	MaximumSupportedPointCount = 10000
)

var (
	ErrPolicyUnavailable = errors.New(
		"projection horizon policy is unavailable",
	)
	ErrPolicyNameRequired = errors.New(
		"projection horizon policy name is required and must be normalized",
	)
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
		"projection step must be greater than zero and must not exceed the minimum duration",
	)
	ErrDurationGridInvalid = errors.New(
		"configured projection durations must be exactly divisible by the projection step",
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
	ErrRequestedDurationGridInvalid = errors.New(
		"requested projection duration must be exactly divisible by the projection step",
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

func (config Config) Validate() error {
	normalizedName := strings.TrimSpace(config.Name)
	if normalizedName == "" || normalizedName != config.Name {
		return ErrPolicyNameRequired
	}
	if config.MinimumDuration <= 0 {
		return ErrMinimumDurationInvalid
	}
	if config.MaximumDuration < config.MinimumDuration {
		return ErrMaximumDurationInvalid
	}
	if config.DefaultDuration < config.MinimumDuration ||
		config.DefaultDuration > config.MaximumDuration {
		return ErrDefaultDurationInvalid
	}
	if config.Step <= 0 || config.Step > config.MinimumDuration {
		return ErrStepInvalid
	}
	if config.MinimumDuration%config.Step != 0 ||
		config.DefaultDuration%config.Step != 0 ||
		config.MaximumDuration%config.Step != 0 {
		return ErrDurationGridInvalid
	}
	if config.MaximumPointCount < 1 ||
		config.MaximumPointCount > MaximumSupportedPointCount {
		return fmt.Errorf(
			"%w: configured=%d supported_range=1..%d",
			ErrMaximumPointCountInvalid,
			config.MaximumPointCount,
			MaximumSupportedPointCount,
		)
	}

	requiredPointCountValue := config.MaximumDuration / config.Step
	if requiredPointCountValue > time.Duration(MaximumSupportedPointCount) {
		return fmt.Errorf(
			"%w: maximum duration requires %d points, supported maximum is %d",
			ErrMaximumPointCountInvalid,
			requiredPointCountValue,
			MaximumSupportedPointCount,
		)
	}
	requiredPointCount := int(requiredPointCountValue)
	if requiredPointCount > config.MaximumPointCount {
		return fmt.Errorf(
			"%w: maximum duration requires %d points, configured maximum is %d",
			ErrMaximumPointCountInvalid,
			requiredPointCount,
			config.MaximumPointCount,
		)
	}

	return nil
}

type Policy struct {
	config Config
}

func New(config Config) (*Policy, error) {
	if err := config.Validate(); err != nil {
		return nil, err
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

type DurationResolution struct {
	RequestedDuration time.Duration
	EffectiveDuration time.Duration
	Truncated         bool
	TruncationReason  TruncationReason
}

func (config Config) ResolveRequestedDuration(
	requestedDuration time.Duration,
) (DurationResolution, error) {
	if err := config.Validate(); err != nil {
		return DurationResolution{}, err
	}

	if requestedDuration == 0 {
		requestedDuration = config.DefaultDuration
	}
	if requestedDuration < config.MinimumDuration {
		return DurationResolution{},
			&DurationBelowMinimumError{
				Requested: requestedDuration,
				Minimum:   config.MinimumDuration,
			}
	}

	effectiveDuration := requestedDuration
	truncated := false
	truncationReason := TruncationReasonNone
	if effectiveDuration > config.MaximumDuration {
		effectiveDuration = config.MaximumDuration
		truncated = true
		truncationReason = TruncationReasonMaximumDuration
	}
	if effectiveDuration%config.Step != 0 {
		return DurationResolution{}, fmt.Errorf(
			"%w: requested=%s effective=%s step=%s",
			ErrRequestedDurationGridInvalid,
			requestedDuration,
			effectiveDuration,
			config.Step,
		)
	}

	return DurationResolution{
		RequestedDuration: requestedDuration,
		EffectiveDuration: effectiveDuration,
		Truncated:         truncated,
		TruncationReason:  truncationReason,
	}, nil
}

func (policy *Policy) BuildDefault(
	asOfTime time.Time,
) (Plan, error) {
	if policy == nil {
		return Plan{}, ErrPolicyUnavailable
	}

	return policy.Build(
		Request{
			AsOfTime:          asOfTime,
			RequestedDuration: policy.config.DefaultDuration,
		},
	)
}

func (policy *Policy) Build(request Request) (Plan, error) {
	if policy == nil {
		return Plan{}, ErrPolicyUnavailable
	}
	if request.AsOfTime.IsZero() {
		return Plan{}, ErrAsOfTimeRequired
	}

	resolution, err := policy.config.ResolveRequestedDuration(
		request.RequestedDuration,
	)
	if err != nil {
		return Plan{}, err
	}

	asOfTime := request.AsOfTime.UTC()
	forecastTimes := buildForecastTimes(
		asOfTime,
		resolution.EffectiveDuration,
		policy.config.Step,
	)
	if len(forecastTimes) > policy.config.MaximumPointCount {
		return Plan{}, fmt.Errorf(
			"%w: planned %d points, maximum is %d",
			ErrMaximumPointCountInvalid,
			len(forecastTimes),
			policy.config.MaximumPointCount,
		)
	}

	return FinalizePlan(
		Plan{
			Version:    Version,
			PolicyName: policy.config.Name,
			AsOfTime:   asOfTime,
			EndTime: asOfTime.Add(
				resolution.EffectiveDuration,
			),
			Step: policy.config.Step,
			RequestedDuration: resolution.
				RequestedDuration,
			EffectiveDuration: resolution.
				EffectiveDuration,
			ForecastTimes: forecastTimes,
			Truncated:     resolution.Truncated,
			TruncationReason: resolution.
				TruncationReason,
		},
	)
}

type DurationBelowMinimumError struct {
	Requested time.Duration
	Minimum   time.Duration
}

func (err *DurationBelowMinimumError) Error() string {
	return fmt.Sprintf(
		"requested projection duration %s is below minimum %s",
		err.Requested,
		err.Minimum,
	)
}

func (err *DurationBelowMinimumError) Unwrap() error {
	return ErrRequestedDurationBelowMinimum
}

func exactForecastPointCount(
	duration time.Duration,
	step time.Duration,
) int {
	if duration <= 0 || step <= 0 || duration%step != 0 {
		return 0
	}

	count := duration / step
	if count > time.Duration(MaximumSupportedPointCount) {
		return MaximumSupportedPointCount + 1
	}

	return int(count)
}

func buildForecastTimes(
	asOfTime time.Time,
	duration time.Duration,
	step time.Duration,
) []time.Time {
	pointCount := exactForecastPointCount(duration, step)
	result := make([]time.Time, pointCount)
	for index := range result {
		offset := time.Duration(index+1) * step
		result[index] = asOfTime.Add(offset)
	}

	return result
}
