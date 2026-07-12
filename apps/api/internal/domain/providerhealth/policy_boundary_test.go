package providerhealth

import (
	"slices"
	"testing"
	"time"
)

func TestPolicyMinimumHealthyRequestSampleBoundary(
	t *testing.T,
) {
	t.Parallel()

	policy := testPolicy()
	now := time.Date(
		2026,
		time.July,
		12,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	atBoundary := healthyBoundaryInput(now)
	atBoundary.RequestsTotal = policy.MinimumHealthyRequestSamples
	atBoundary.RequestsSuccessful = policy.MinimumHealthyRequestSamples

	snapshot, err := policy.Evaluate(atBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if snapshot.Status != StatusHealthy {
		t.Fatalf(
			"status at sample boundary = %q, want %q",
			snapshot.Status,
			StatusHealthy,
		)
	}

	belowBoundary := atBoundary
	belowBoundary.RequestsTotal--
	belowBoundary.RequestsSuccessful--

	snapshot, err = policy.Evaluate(belowBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertReason(
		t,
		snapshot,
		"provider_request_sample_below_healthy_threshold",
	)
}

func TestPolicyMinimumHealthySuccessRatioBoundary(
	t *testing.T,
) {
	t.Parallel()

	policy := testPolicy()
	now := time.Date(
		2026,
		time.July,
		12,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	atBoundary := healthyBoundaryInput(now)
	atBoundary.RequestsTotal = 20
	atBoundary.RequestsSuccessful = 19

	snapshot, err := policy.Evaluate(atBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if snapshot.Status != StatusHealthy {
		t.Fatalf(
			"status at success-ratio boundary = %q, want %q",
			snapshot.Status,
			StatusHealthy,
		)
	}

	belowBoundary := atBoundary
	belowBoundary.RequestsSuccessful = 18

	snapshot, err = policy.Evaluate(belowBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertReason(
		t,
		snapshot,
		"provider_success_ratio_below_healthy_threshold",
	)
}

func TestPolicyMaximumHealthyAverageLatencyBoundary(
	t *testing.T,
) {
	t.Parallel()

	policy := testPolicy()
	now := time.Date(
		2026,
		time.July,
		12,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	atBoundary := healthyBoundaryInput(now)
	atBoundary.AverageLatency = policy.MaximumHealthyAverageLatency

	snapshot, err := policy.Evaluate(atBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if snapshot.Status != StatusHealthy {
		t.Fatalf(
			"status at latency boundary = %q, want %q",
			snapshot.Status,
			StatusHealthy,
		)
	}

	aboveBoundary := atBoundary
	aboveBoundary.AverageLatency++

	snapshot, err = policy.Evaluate(aboveBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertReason(
		t,
		snapshot,
		"provider_average_latency_above_healthy_threshold",
	)
}

func TestPolicyMaximumHealthyRejectionRatioBoundary(
	t *testing.T,
) {
	t.Parallel()

	policy := testPolicy()
	now := time.Date(
		2026,
		time.July,
		12,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	atBoundary := healthyBoundaryInput(now)
	atBoundary.Observations = ObservationEvidence{
		Received: 100,
		Accepted: 80,
		Rejected: 20,
	}

	snapshot, err := policy.Evaluate(atBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if snapshot.Status != StatusHealthy {
		t.Fatalf(
			"status at rejection-ratio boundary = %q, want %q",
			snapshot.Status,
			StatusHealthy,
		)
	}

	aboveBoundary := atBoundary
	aboveBoundary.Observations = ObservationEvidence{
		Received: 100,
		Accepted: 79,
		Rejected: 21,
	}

	snapshot, err = policy.Evaluate(aboveBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertReason(
		t,
		snapshot,
		"provider_observation_rejection_ratio_above_healthy_threshold",
	)
}

func TestPolicyStaleAndUnavailableAgeBoundaries(
	t *testing.T,
) {
	t.Parallel()

	policy := testPolicy()
	now := time.Date(
		2026,
		time.July,
		12,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	atStaleBoundary := healthyBoundaryInput(now)
	lastSuccess := now.Add(-policy.StaleAfter)
	atStaleBoundary.LastSuccessAt = &lastSuccess

	snapshot, err := policy.Evaluate(atStaleBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if slices.Contains(
		snapshot.Reasons,
		"provider_last_success_is_stale",
	) {
		t.Fatalf(
			"reasons at stale boundary = %v",
			snapshot.Reasons,
		)
	}

	beyondStaleBoundary := atStaleBoundary
	lastSuccess = now.Add(-policy.StaleAfter - time.Nanosecond)
	beyondStaleBoundary.LastSuccessAt = &lastSuccess

	snapshot, err = policy.Evaluate(beyondStaleBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertReason(
		t,
		snapshot,
		"provider_last_success_is_stale",
	)

	atUnavailableBoundary := healthyBoundaryInput(now)
	lastSuccess = now.Add(-policy.UnavailableAfter)
	atUnavailableBoundary.LastSuccessAt = &lastSuccess

	snapshot, err = policy.Evaluate(atUnavailableBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if snapshot.Status == StatusUnavailable {
		t.Fatalf(
			"status at unavailable boundary = %q",
			snapshot.Status,
		)
	}

	beyondUnavailableBoundary := atUnavailableBoundary
	lastSuccess = now.Add(
		-policy.UnavailableAfter - time.Nanosecond,
	)
	beyondUnavailableBoundary.LastSuccessAt = &lastSuccess

	snapshot, err = policy.Evaluate(beyondUnavailableBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if snapshot.Status != StatusUnavailable {
		t.Fatalf(
			"status beyond unavailable boundary = %q, want %q",
			snapshot.Status,
			StatusUnavailable,
		)
	}
	assertReason(
		t,
		snapshot,
		"provider_last_success_exceeds_unavailable_threshold",
	)
}

func TestPolicyConsecutiveFailureBoundaries(
	t *testing.T,
) {
	t.Parallel()

	policy := testPolicy()
	policy.MaximumHealthyConsecutiveFailures = 1
	now := time.Date(
		2026,
		time.July,
		12,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	atHealthyBoundary := healthyBoundaryInput(now)
	atHealthyBoundary.ConsecutiveFailures =
		policy.MaximumHealthyConsecutiveFailures

	snapshot, err := policy.Evaluate(atHealthyBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if slices.Contains(
		snapshot.Reasons,
		"provider_has_recent_consecutive_failures",
	) {
		t.Fatalf(
			"reasons at healthy failure boundary = %v",
			snapshot.Reasons,
		)
	}

	aboveHealthyBoundary := atHealthyBoundary
	aboveHealthyBoundary.ConsecutiveFailures++

	snapshot, err = policy.Evaluate(aboveHealthyBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertReason(
		t,
		snapshot,
		"provider_has_recent_consecutive_failures",
	)

	atUnavailableBoundary := healthyBoundaryInput(now)
	atUnavailableBoundary.ConsecutiveFailures =
		policy.UnavailableConsecutiveFailures
	atUnavailableBoundary.LatestOutcome =
		RequestOutcomeNetworkError
	lastFailure := now
	atUnavailableBoundary.LastFailureAt = &lastFailure

	snapshot, err = policy.Evaluate(atUnavailableBoundary)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if snapshot.Status != StatusUnavailable {
		t.Fatalf(
			"status at unavailable failure boundary = %q, want %q",
			snapshot.Status,
			StatusUnavailable,
		)
	}
	assertReason(
		t,
		snapshot,
		"provider_consecutive_failure_limit_reached",
	)
}

func TestPolicyBudgetResetBoundary(
	t *testing.T,
) {
	t.Parallel()

	policy := testPolicy()
	now := time.Date(
		2026,
		time.July,
		12,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	beforeReset := healthyBoundaryInput(now)
	resetAt := now.Add(time.Nanosecond)
	beforeReset.Budget = BudgetEvidence{
		State:     BudgetStateExhausted,
		Limit:     100,
		Remaining: 0,
		ResetsAt:  &resetAt,
	}

	snapshot, err := policy.Evaluate(beforeReset)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if snapshot.Status != StatusUnavailable {
		t.Fatalf(
			"status before budget reset = %q, want %q",
			snapshot.Status,
			StatusUnavailable,
		)
	}
	if snapshot.Budget.State != BudgetStateExhausted {
		t.Fatalf(
			"budget before reset = %q, want %q",
			snapshot.Budget.State,
			BudgetStateExhausted,
		)
	}

	atReset := healthyBoundaryInput(now)
	resetAt = now
	atReset.Budget = BudgetEvidence{
		State:     BudgetStateExhausted,
		Limit:     100,
		Remaining: 0,
		ResetsAt:  &resetAt,
	}

	snapshot, err = policy.Evaluate(atReset)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if snapshot.Budget.State != BudgetStateUnknown {
		t.Fatalf(
			"budget at reset = %q, want %q",
			snapshot.Budget.State,
			BudgetStateUnknown,
		)
	}
	if snapshot.Status != StatusHealthy {
		t.Fatalf(
			"status at budget reset = %q, want %q",
			snapshot.Status,
			StatusHealthy,
		)
	}
}

func healthyBoundaryInput(
	now time.Time,
) EvaluationInput {
	firstRequest := now.Add(-15 * time.Minute)
	lastRequest := now
	lastSuccess := now

	return EvaluationInput{
		ProviderName:        "boundary-provider",
		EvaluatedAt:         now,
		FirstRequestAt:      &firstRequest,
		LastRequestAt:       &lastRequest,
		LastSuccessAt:       &lastSuccess,
		RequestsTotal:       20,
		RequestsSuccessful:  20,
		ConsecutiveFailures: 0,
		AverageLatency:      time.Second,
		LatestOutcome:       RequestOutcomeSuccess,
		Observations: ObservationEvidence{
			Received: 100,
			Accepted: 100,
			Rejected: 0,
		},
		Budget: BudgetEvidence{
			State:     BudgetStateAvailable,
			Limit:     100,
			Remaining: 50,
		},
	}
}

func assertReason(
	t *testing.T,
	snapshot Snapshot,
	reason string,
) {
	t.Helper()

	if !slices.Contains(
		snapshot.Reasons,
		reason,
	) {
		t.Fatalf(
			"reasons = %v, missing %q",
			snapshot.Reasons,
			reason,
		)
	}
}
