package projectionhorizon

import (
	"errors"
	"testing"
	"time"
)

func TestNewAcceptsExplicitBoundedPolicy(
	t *testing.T,
) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	config := policy.Config()
	if config.Name !=
		"short-horizon-baseline" ||
		config.MinimumDuration !=
			time.Minute ||
		config.DefaultDuration !=
			5*time.Minute ||
		config.MaximumDuration !=
			10*time.Minute ||
		config.Step != time.Minute ||
		config.MaximumPointCount != 10 {
		t.Fatalf(
			"unexpected policy config: %#v",
			config,
		)
	}
}

func TestNewRejectsInvalidPolicy(
	t *testing.T,
) {
	tests := []struct {
		name      string
		mutate    func(*Config)
		wantError error
	}{
		{
			name: "minimum duration",
			mutate: func(config *Config) {
				config.MinimumDuration = 0
			},
			wantError: ErrMinimumDurationInvalid,
		},
		{
			name: "maximum duration",
			mutate: func(config *Config) {
				config.MaximumDuration =
					30 * time.Second
			},
			wantError: ErrMaximumDurationInvalid,
		},
		{
			name: "default duration",
			mutate: func(config *Config) {
				config.DefaultDuration =
					20 * time.Minute
			},
			wantError: ErrDefaultDurationInvalid,
		},
		{
			name: "step",
			mutate: func(config *Config) {
				config.Step = 0
			},
			wantError: ErrStepInvalid,
		},
		{
			name: "maximum point count",
			mutate: func(config *Config) {
				config.MaximumPointCount = 9
			},
			wantError: ErrMaximumPointCountInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				config := validPolicyConfig()
				test.mutate(&config)

				policy, err := New(config)
				if policy != nil {
					t.Fatalf(
						"policy = %#v, want nil",
						policy,
					)
				}
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

func TestBuildUsesDefaultDuration(
	t *testing.T,
) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}
	asOfTime := horizonTestAsOfTime()

	plan, err := policy.Build(
		Request{
			AsOfTime: asOfTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"Build() error = %v",
			err,
		)
	}

	if plan.Version != Version ||
		plan.PolicyName !=
			"short-horizon-baseline" ||
		plan.RequestedDuration !=
			5*time.Minute ||
		plan.EffectiveDuration !=
			5*time.Minute ||
		plan.Truncated ||
		plan.TruncationReason !=
			TruncationReasonNone ||
		len(plan.ForecastTimes) != 5 ||
		!plan.EndTime.Equal(
			asOfTime.Add(
				5*time.Minute,
			),
		) {
		t.Fatalf(
			"unexpected plan: %#v",
			plan,
		)
	}

	horizon := plan.ContractHorizon()
	if !horizon.AsOfTime.Equal(asOfTime) ||
		!horizon.EndTime.Equal(plan.EndTime) ||
		horizon.Step != time.Minute {
		t.Fatalf(
			"unexpected contract horizon: %#v",
			horizon,
		)
	}
}

func TestBuildTruncatesAboveMaximumDuration(
	t *testing.T,
) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	plan, err := policy.Build(
		Request{
			AsOfTime:          horizonTestAsOfTime(),
			RequestedDuration: 15 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf(
			"Build() error = %v",
			err,
		)
	}

	if !plan.Truncated ||
		plan.TruncationReason !=
			TruncationReasonMaximumDuration ||
		plan.RequestedDuration !=
			15*time.Minute ||
		plan.EffectiveDuration !=
			10*time.Minute ||
		len(plan.ForecastTimes) != 10 {
		t.Fatalf(
			"unexpected truncated plan: %#v",
			plan,
		)
	}
}

func TestBuildIncludesUnalignedFinalEndpoint(
	t *testing.T,
) {
	config := validPolicyConfig()
	config.MinimumDuration = 30 * time.Second
	config.DefaultDuration =
		90 * time.Second
	config.MaximumDuration =
		90 * time.Second
	config.Step = time.Minute
	config.MaximumPointCount = 2

	policy, err := New(config)
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	asOfTime := horizonTestAsOfTime()
	plan, err := policy.Build(
		Request{
			AsOfTime: asOfTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"Build() error = %v",
			err,
		)
	}

	if len(plan.ForecastTimes) != 2 ||
		!plan.ForecastTimes[0].Equal(
			asOfTime.Add(time.Minute),
		) ||
		!plan.ForecastTimes[1].Equal(
			asOfTime.Add(
				90*time.Second,
			),
		) {
		t.Fatalf(
			"forecast times = %#v",
			plan.ForecastTimes,
		)
	}
}

func TestBuildRejectsMissingAsOfAndShortDuration(
	t *testing.T,
) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}

	_, err = policy.Build(
		Request{},
	)
	if !errors.Is(
		err,
		ErrAsOfTimeRequired,
	) {
		t.Fatalf(
			"missing as-of error = %v",
			err,
		)
	}

	_, err = policy.Build(
		Request{
			AsOfTime:          horizonTestAsOfTime(),
			RequestedDuration: 30 * time.Second,
		},
	)
	if !errors.Is(
		err,
		ErrRequestedDurationBelowMinimum,
	) {
		t.Fatalf(
			"short duration error = %v",
			err,
		)
	}
}

func TestPlanCloneDoesNotShareForecastTimes(
	t *testing.T,
) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf(
			"New() error = %v",
			err,
		)
	}
	plan, err := policy.Build(
		Request{
			AsOfTime: horizonTestAsOfTime(),
		},
	)
	if err != nil {
		t.Fatalf(
			"Build() error = %v",
			err,
		)
	}

	cloned := plan.Clone()
	cloned.ForecastTimes[0] =
		cloned.ForecastTimes[0].Add(
			time.Hour,
		)

	if cloned.ForecastTimes[0].Equal(
		plan.ForecastTimes[0],
	) {
		t.Fatal(
			"Plan.Clone() shared forecast times",
		)
	}
}

func validPolicyConfig() Config {
	return Config{
		Name: "short-horizon-baseline",

		MinimumDuration: time.Minute,
		DefaultDuration: 5 * time.Minute,
		MaximumDuration: 10 * time.Minute,
		Step:            time.Minute,

		MaximumPointCount: 10,
	}
}

func horizonTestAsOfTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		17,
		0,
		0,
		0,
		time.FixedZone(
			"AZT",
			4*60*60,
		),
	).UTC()
}
