package providerhealth

import (
	"strings"
	"testing"
	"time"
)

func TestPolicyRejectsFutureRequestEvidence(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Nanosecond)
	tests := []struct {
		name   string
		mutate func(*EvaluationInput)
	}{
		{name: "first request", mutate: func(input *EvaluationInput) { input.FirstRequestAt = &future }},
		{name: "last request", mutate: func(input *EvaluationInput) { input.LastRequestAt = &future }},
		{name: "last success", mutate: func(input *EvaluationInput) { input.LastSuccessAt = &future }},
		{name: "last failure", mutate: func(input *EvaluationInput) { input.LastFailureAt = &future }},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			input := healthyBoundaryInput(now)
			test.mutate(&input)
			_, err := testPolicy().Evaluate(input)
			if err == nil || !strings.Contains(err.Error(), "cannot be after evaluation timestamp") {
				t.Fatalf("Evaluate() error = %v", err)
			}
		})
	}
}

func TestPolicyRejectsSuccessAfterLastRequest(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	input := healthyBoundaryInput(now)
	lastRequest := now.Add(-time.Minute)
	lastSuccess := lastRequest.Add(time.Nanosecond)
	input.LastRequestAt = &lastRequest
	input.LastSuccessAt = &lastSuccess
	_, err := testPolicy().Evaluate(input)
	if err == nil || !strings.Contains(err.Error(), "success timestamp cannot be after last request") {
		t.Fatalf("Evaluate() error = %v", err)
	}
}

func TestPolicyRejectsFailureAfterLastRequest(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	input := healthyBoundaryInput(now)
	lastRequest := now.Add(-time.Minute)
	lastSuccess := lastRequest
	lastFailure := lastRequest.Add(time.Nanosecond)
	input.LastRequestAt = &lastRequest
	input.LastSuccessAt = &lastSuccess
	input.LastFailureAt = &lastFailure
	_, err := testPolicy().Evaluate(input)
	if err == nil || !strings.Contains(err.Error(), "failure timestamp cannot be after last request") {
		t.Fatalf("Evaluate() error = %v", err)
	}
}

func TestPolicyObservationCounterValidationCannotOverflow(t *testing.T) {
	t.Parallel()
	const maximumInt64 = int64(^uint64(0) >> 1)
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	_, err := testPolicy().Evaluate(EvaluationInput{
		ProviderName:  "provider",
		EvaluatedAt:   now,
		LatestOutcome: RequestOutcomeUnknown,
		Observations:  ObservationEvidence{Received: maximumInt64, Accepted: maximumInt64, Rejected: 1},
		Budget:        BudgetEvidence{State: BudgetStateUnknown},
	})
	if err == nil || !strings.Contains(err.Error(), "cannot exceed received observations") {
		t.Fatalf("Evaluate() error = %v", err)
	}
}
