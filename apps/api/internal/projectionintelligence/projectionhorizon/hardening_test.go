package projectionhorizon

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

func TestNewRejectsNonNormalizedPolicyName(t *testing.T) {
	for _, name := range []string{"", "   ", " policy "} {
		config := validPolicyConfig()
		config.Name = name

		policy, err := New(config)
		if policy != nil {
			t.Fatalf("policy = %#v, want nil", policy)
		}
		if !errors.Is(err, ErrPolicyNameRequired) {
			t.Fatalf("name %q error = %v, want %v", name, err, ErrPolicyNameRequired)
		}
	}
}

func TestNewRejectsNonDivisibleConfiguredGrid(t *testing.T) {
	config := validPolicyConfig()
	config.DefaultDuration = 90 * time.Second

	policy, err := New(config)
	if policy != nil {
		t.Fatalf("policy = %#v, want nil", policy)
	}
	if !errors.Is(err, ErrDurationGridInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrDurationGridInvalid)
	}
}

func TestNewRejectsStepAboveMinimumDuration(t *testing.T) {
	config := validPolicyConfig()
	config.Step = 2 * time.Minute

	policy, err := New(config)
	if policy != nil {
		t.Fatalf("policy = %#v, want nil", policy)
	}
	if !errors.Is(err, ErrStepInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrStepInvalid)
	}
}

func TestNewRejectsUnsafeMaximumPointCount(t *testing.T) {
	config := validPolicyConfig()
	config.MaximumPointCount = MaximumSupportedPointCount + 1

	policy, err := New(config)
	if policy != nil {
		t.Fatalf("policy = %#v, want nil", policy)
	}
	if !errors.Is(err, ErrMaximumPointCountInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrMaximumPointCountInvalid)
	}
}

func TestNewRejectsExtremeConfigurationBeforeScheduleAllocation(t *testing.T) {
	maximum := time.Duration(math.MaxInt64)
	maximum -= maximum % time.Second
	config := Config{
		Name:              "extreme-policy",
		MinimumDuration:   time.Second,
		DefaultDuration:   time.Second,
		MaximumDuration:   maximum,
		Step:              time.Second,
		MaximumPointCount: MaximumSupportedPointCount,
	}

	policy, err := New(config)
	if policy != nil {
		t.Fatalf("policy = %#v, want nil", policy)
	}
	if !errors.Is(err, ErrMaximumPointCountInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrMaximumPointCountInvalid)
	}
}

func TestNilPolicyReturnsLifecycleError(t *testing.T) {
	var policy *Policy

	_, err := policy.Build(Request{AsOfTime: horizonTestAsOfTime()})
	if !errors.Is(err, ErrPolicyUnavailable) {
		t.Fatalf("Build() error = %v, want %v", err, ErrPolicyUnavailable)
	}

	_, err = policy.BuildDefault(horizonTestAsOfTime())
	if !errors.Is(err, ErrPolicyUnavailable) {
		t.Fatalf("BuildDefault() error = %v, want %v", err, ErrPolicyUnavailable)
	}
}

func TestBuildDefaultUsesExplicitDefaultPath(t *testing.T) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	plan, err := policy.BuildDefault(horizonTestAsOfTime())
	if err != nil {
		t.Fatalf("BuildDefault() error = %v", err)
	}
	if plan.RequestedDuration != 5*time.Minute ||
		plan.EffectiveDuration != 5*time.Minute {
		t.Fatalf("unexpected default plan: %#v", plan)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("default plan validation error = %v", err)
	}
}

func TestBuildNormalizesNonUTCTime(t *testing.T) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	input := time.Date(
		2026,
		time.July,
		15,
		21,
		0,
		0,
		0,
		time.FixedZone("AZT", 4*60*60),
	)

	plan, err := policy.Build(
		Request{
			AsOfTime:          input,
			RequestedDuration: time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if plan.AsOfTime.Location() != time.UTC ||
		plan.EndTime.Location() != time.UTC ||
		plan.ForecastTimes[0].Location() != time.UTC ||
		!plan.AsOfTime.Equal(input.UTC()) {
		t.Fatalf("plan was not normalized to canonical UTC: %#v", plan)
	}
}

func TestPlanFingerprintCoversRequestedAndTruncationEvidence(t *testing.T) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	asOfTime := horizonTestAsOfTime()

	exact, err := policy.Build(
		Request{
			AsOfTime:          asOfTime,
			RequestedDuration: 10 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("exact Build() error = %v", err)
	}
	truncated, err := policy.Build(
		Request{
			AsOfTime:          asOfTime,
			RequestedDuration: 15 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("truncated Build() error = %v", err)
	}

	if exact.EndTime != truncated.EndTime || exact.Step != truncated.Step {
		t.Fatalf("test precondition failed: exact=%#v truncated=%#v", exact, truncated)
	}
	if exact.Fingerprint == truncated.Fingerprint {
		t.Fatal("different requested and truncation evidence produced the same fingerprint")
	}
	if !strings.HasPrefix(exact.Fingerprint, "sha256:") ||
		len(exact.Fingerprint) != len("sha256:")+64 {
		t.Fatalf("unexpected fingerprint = %q", exact.Fingerprint)
	}
	if err := exact.Validate(); err != nil {
		t.Fatalf("exact plan validation error = %v", err)
	}
	if err := truncated.Validate(); err != nil {
		t.Fatalf("truncated plan validation error = %v", err)
	}
}

func TestPlanValidationRejectsManualInvariantCorruption(t *testing.T) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	plan, err := policy.BuildDefault(horizonTestAsOfTime())
	if err != nil {
		t.Fatalf("BuildDefault() error = %v", err)
	}

	plan.EndTime = plan.EndTime.Add(time.Second)
	if !errors.Is(plan.Validate(), ErrPlanInvalid) {
		t.Fatalf("corrupted plan validation error = %v", plan.Validate())
	}
}

func TestPlanValidationRejectsFingerprintCorruption(t *testing.T) {
	policy, err := New(validPolicyConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	plan, err := policy.BuildDefault(horizonTestAsOfTime())
	if err != nil {
		t.Fatalf("BuildDefault() error = %v", err)
	}

	plan.Fingerprint = "sha256:" + strings.Repeat("0", 64)
	err = plan.Validate()
	if !errors.Is(err, ErrPlanInvalid) ||
		!errors.Is(err, ErrPlanFingerprintInvalid) {
		t.Fatalf("error = %v, want plan and fingerprint errors", err)
	}
}
