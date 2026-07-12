package providerhealth

import (
	"slices"
	"testing"
	"time"
)

func TestPolicyEvaluateHealthyProvider(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	firstRequest := now.Add(-15 * time.Minute)
	lastRequest := now.Add(-10 * time.Second)
	lastSuccess := lastRequest

	snapshot, err := testPolicy().Evaluate(EvaluationInput{
		ProviderName:        "airplanes.live",
		EvaluatedAt:         now,
		FirstRequestAt:      &firstRequest,
		LastRequestAt:       &lastRequest,
		LastSuccessAt:       &lastSuccess,
		RequestsTotal:       100,
		RequestsSuccessful:  99,
		ConsecutiveFailures: 0,
		AverageLatency:      800 * time.Millisecond,
		LatestOutcome:       RequestOutcomeSuccess,
		Observations: ObservationEvidence{
			Received: 1_000,
			Accepted: 980,
			Rejected: 20,
		},
		Budget: BudgetEvidence{
			State:     BudgetStateAvailable,
			Limit:     100,
			Remaining: 70,
		},
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if snapshot.Status != StatusHealthy {
		t.Fatalf("status = %q, want %q", snapshot.Status, StatusHealthy)
	}
	if snapshot.SuccessRatio != 0.99 {
		t.Fatalf("success ratio = %v, want 0.99", snapshot.SuccessRatio)
	}
	if snapshot.RejectionRatio != 0.02 {
		t.Fatalf("rejection ratio = %v, want 0.02", snapshot.RejectionRatio)
	}
	if !slices.Contains(snapshot.Reasons, "provider_operating_within_health_policy") {
		t.Fatalf("reasons = %v", snapshot.Reasons)
	}
}

func TestPolicyEvaluateDegradedProvider(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	firstRequest := now.Add(-15 * time.Minute)
	lastRequest := now.Add(-20 * time.Second)
	lastSuccess := now.Add(-90 * time.Second)
	lastFailure := lastRequest

	snapshot, err := testPolicy().Evaluate(EvaluationInput{
		ProviderName:        "secondary-provider",
		EvaluatedAt:         now,
		FirstRequestAt:      &firstRequest,
		LastRequestAt:       &lastRequest,
		LastSuccessAt:       &lastSuccess,
		LastFailureAt:       &lastFailure,
		RequestsTotal:       100,
		RequestsSuccessful:  90,
		ConsecutiveFailures: 1,
		AverageLatency:      4 * time.Second,
		LatestOutcome:       RequestOutcomeRateLimited,
		Observations: ObservationEvidence{
			Received: 100,
			Accepted: 70,
			Rejected: 30,
		},
		Budget: BudgetEvidence{
			State:     BudgetStateConstrained,
			Limit:     100,
			Remaining: 5,
		},
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if snapshot.Status != StatusDegraded {
		t.Fatalf("status = %q, want %q", snapshot.Status, StatusDegraded)
	}

	wantReasons := []string{
		"provider_average_latency_above_healthy_threshold",
		"provider_budget_is_constrained",
		"provider_has_recent_consecutive_failures",
		"provider_last_success_is_stale",
		"provider_latest_request_was_not_successful",
		"provider_observation_rejection_ratio_above_healthy_threshold",
		"provider_success_ratio_below_healthy_threshold",
	}
	for _, reason := range wantReasons {
		if !slices.Contains(snapshot.Reasons, reason) {
			t.Fatalf("reasons = %v, missing %q", snapshot.Reasons, reason)
		}
	}
}

func TestPolicyEvaluateUnavailableProvider(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	firstRequest := now.Add(-30 * time.Minute)
	lastRequest := now.Add(-10 * time.Second)
	lastSuccess := now.Add(-11 * time.Minute)
	lastFailure := lastRequest

	snapshot, err := testPolicy().Evaluate(EvaluationInput{
		ProviderName:        "primary-provider",
		EvaluatedAt:         now,
		FirstRequestAt:      &firstRequest,
		LastRequestAt:       &lastRequest,
		LastSuccessAt:       &lastSuccess,
		LastFailureAt:       &lastFailure,
		RequestsTotal:       100,
		RequestsSuccessful:  80,
		ConsecutiveFailures: 3,
		AverageLatency:      2 * time.Second,
		LatestOutcome:       RequestOutcomeTimeout,
		Budget: BudgetEvidence{
			State:     BudgetStateAvailable,
			Limit:     100,
			Remaining: 50,
		},
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if snapshot.Status != StatusUnavailable {
		t.Fatalf("status = %q, want %q", snapshot.Status, StatusUnavailable)
	}
	if !slices.Contains(snapshot.Reasons, "provider_consecutive_failure_limit_reached") {
		t.Fatalf("reasons = %v", snapshot.Reasons)
	}
	if !slices.Contains(snapshot.Reasons, "provider_last_success_exceeds_unavailable_threshold") {
		t.Fatalf("reasons = %v", snapshot.Reasons)
	}
}

func TestPolicyEvaluateUnknownWithoutHistory(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)

	snapshot, err := testPolicy().Evaluate(EvaluationInput{
		ProviderName:  "new-provider",
		EvaluatedAt:   now,
		LatestOutcome: RequestOutcomeUnknown,
		Budget: BudgetEvidence{
			State: BudgetStateUnknown,
		},
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if snapshot.Status != StatusUnknown {
		t.Fatalf("status = %q, want %q", snapshot.Status, StatusUnknown)
	}
	if !slices.Contains(snapshot.Limitations, "provider_request_history_absent") {
		t.Fatalf("limitations = %v", snapshot.Limitations)
	}
	if !slices.Contains(snapshot.Limitations, "provider_budget_state_unknown") {
		t.Fatalf("limitations = %v", snapshot.Limitations)
	}
}

func TestPolicyEvaluateDegradedWithSmallRequestSample(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)
	lastRequest := now.Add(-10 * time.Second)
	lastSuccess := lastRequest

	snapshot, err := testPolicy().Evaluate(EvaluationInput{
		ProviderName:        "new-provider",
		EvaluatedAt:         now,
		FirstRequestAt:      &lastRequest,
		LastRequestAt:       &lastRequest,
		LastSuccessAt:       &lastSuccess,
		RequestsTotal:       1,
		RequestsSuccessful:  1,
		ConsecutiveFailures: 0,
		LatestOutcome:       RequestOutcomeSuccess,
		Budget: BudgetEvidence{
			State: BudgetStateUnknown,
		},
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if snapshot.Status != StatusDegraded {
		t.Fatalf("status = %q, want %q", snapshot.Status, StatusDegraded)
	}
	if !slices.Contains(snapshot.Reasons, "provider_request_sample_below_healthy_threshold") {
		t.Fatalf("reasons = %v", snapshot.Reasons)
	}
}

func TestPolicyRejectsInconsistentObservationCounters(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 12, 12, 0, 0, 0, time.UTC)

	_, err := testPolicy().Evaluate(EvaluationInput{
		ProviderName:  "provider",
		EvaluatedAt:   now,
		LatestOutcome: RequestOutcomeUnknown,
		Observations: ObservationEvidence{
			Received: 10,
			Accepted: 9,
			Rejected: 2,
		},
		Budget: BudgetEvidence{
			State: BudgetStateUnknown,
		},
	})
	if err == nil {
		t.Fatal("Evaluate() error = nil, want error")
	}
}

func testPolicy() Policy {
	return Policy{
		StaleAfter:                        time.Minute,
		UnavailableAfter:                  10 * time.Minute,
		MinimumHealthyRequestSamples:      5,
		MinimumHealthySuccessRatio:        0.95,
		MaximumHealthyAverageLatency:      3 * time.Second,
		MaximumHealthyConsecutiveFailures: 0,
		UnavailableConsecutiveFailures:    3,
		MaximumHealthyRejectionRatio:      0.20,
	}
}
