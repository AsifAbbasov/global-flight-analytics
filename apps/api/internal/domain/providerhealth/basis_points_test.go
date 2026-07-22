package providerhealth

import (
	"slices"
	"testing"
	"time"
)

func TestPolicyUsesExactIntegerRatioForLargeCounters(t *testing.T) {
	const denominator int64 = 9_007_199_254_740_993
	const numerator int64 = 9_007_199_254_740_992
	now := time.Now().UTC()
	request := now
	policy := testPolicy()
	policy.MinimumHealthySuccessRatio = 1
	snapshot, err := policy.Evaluate(EvaluationInput{
		ProviderName: "large-counter-provider", EvaluatedAt: now,
		FirstRequestAt: &request, LastRequestAt: &request, LastSuccessAt: &request,
		RequestsTotal: denominator, RequestsSuccessful: numerator,
		LatestOutcome: RequestOutcomeSuccess,
		Budget:        BudgetEvidence{State: BudgetStateUnknown},
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !slices.Contains(snapshot.Reasons, "provider_success_ratio_below_healthy_threshold") {
		t.Fatalf("reasons = %v", snapshot.Reasons)
	}
}

func TestRatioToBasisPointsUsesPolicyScale(t *testing.T) {
	if got := ratioToBasisPoints(0.95); got != 9_500 {
		t.Fatalf("threshold = %d, want 9500", got)
	}
	if !ratioAtLeastBasisPoints(19, 20, 9_500) {
		t.Fatal("19/20 must satisfy 9500 basis points")
	}
}
