package providerhealth

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type Policy struct {
	StaleAfter                        time.Duration
	UnavailableAfter                  time.Duration
	MinimumHealthyRequestSamples      int64
	MinimumHealthySuccessRatio        float64
	MaximumHealthyAverageLatency      time.Duration
	MaximumHealthyConsecutiveFailures int
	UnavailableConsecutiveFailures    int
	MaximumHealthyRejectionRatio      float64
}

func (policy Policy) Evaluate(input EvaluationInput) (Snapshot, error) {
	if err := policy.validate(); err != nil {
		return Snapshot{}, err
	}
	if err := validateInput(input); err != nil {
		return Snapshot{}, err
	}

	providerName := strings.TrimSpace(input.ProviderName)
	evaluatedAt := input.EvaluatedAt.UTC()
	firstRequestAt := utcTimePointer(input.FirstRequestAt)
	lastRequestAt := utcTimePointer(input.LastRequestAt)
	lastSuccessAt := utcTimePointer(input.LastSuccessAt)
	lastFailureAt := utcTimePointer(input.LastFailureAt)
	budget := normalizeBudget(
		input.Budget,
		evaluatedAt,
	)
	input.Budget = budget

	successRatio := ratio(input.RequestsSuccessful, input.RequestsTotal)
	rejectionRatio := ratio(
		input.Observations.Rejected,
		input.Observations.Received,
	)

	snapshot := Snapshot{
		ProviderName: providerName,
		Status:       StatusUnknown,
		EvaluatedAt:  evaluatedAt,

		FirstRequestAt: firstRequestAt,
		LastRequestAt:  lastRequestAt,
		LastSuccessAt:  lastSuccessAt,
		LastFailureAt:  lastFailureAt,

		RequestsTotal:       input.RequestsTotal,
		RequestsSuccessful:  input.RequestsSuccessful,
		SuccessRatio:        round4(successRatio),
		ConsecutiveFailures: input.ConsecutiveFailures,
		AverageLatency:      input.AverageLatency,
		LatestOutcome:       input.LatestOutcome,

		Observations:   input.Observations,
		RejectionRatio: round4(rejectionRatio),
		Budget:         budget,

		LastRequestAgeSeconds: ageSeconds(evaluatedAt, lastRequestAt),
		LastSuccessAgeSeconds: ageSeconds(evaluatedAt, lastSuccessAt),

		Limitations: buildLimitations(input),
	}

	if input.RequestsTotal == 0 {
		snapshot.Reasons = []string{"provider_request_history_absent"}
		return snapshot, nil
	}

	unavailableReasons := policy.unavailableReasons(input, lastSuccessAt, evaluatedAt)
	if len(unavailableReasons) > 0 {
		snapshot.Status = StatusUnavailable
		snapshot.Reasons = uniqueSorted(unavailableReasons)
		return snapshot, nil
	}

	degradedReasons := policy.degradedReasons(
		input,
		lastSuccessAt,
		evaluatedAt,
		successRatio,
		rejectionRatio,
	)
	if len(degradedReasons) > 0 {
		snapshot.Status = StatusDegraded
		snapshot.Reasons = uniqueSorted(degradedReasons)
		return snapshot, nil
	}

	snapshot.Status = StatusHealthy
	snapshot.Reasons = []string{"provider_operating_within_health_policy"}

	return snapshot, nil
}

func (policy Policy) validate() error {
	if policy.StaleAfter <= 0 {
		return errors.New("provider health stale-after duration must be positive")
	}
	if policy.UnavailableAfter <= policy.StaleAfter {
		return errors.New("provider health unavailable-after duration must be greater than stale-after duration")
	}
	if policy.MinimumHealthyRequestSamples <= 0 {
		return errors.New("minimum healthy request samples must be positive")
	}
	if !isRatio(policy.MinimumHealthySuccessRatio) {
		return errors.New("minimum healthy success ratio must be between zero and one")
	}
	if policy.MaximumHealthyAverageLatency <= 0 {
		return errors.New("maximum healthy average latency must be positive")
	}
	if policy.MaximumHealthyConsecutiveFailures < 0 {
		return errors.New("maximum healthy consecutive failures must be non-negative")
	}
	if policy.UnavailableConsecutiveFailures <= policy.MaximumHealthyConsecutiveFailures {
		return errors.New("unavailable consecutive failures must be greater than maximum healthy consecutive failures")
	}
	if !isRatio(policy.MaximumHealthyRejectionRatio) {
		return errors.New("maximum healthy rejection ratio must be between zero and one")
	}

	return nil
}

func validateInput(input EvaluationInput) error {
	if strings.TrimSpace(input.ProviderName) == "" {
		return errors.New("provider name is required")
	}
	if input.EvaluatedAt.IsZero() {
		return errors.New("provider health evaluation timestamp is required")
	}
	if input.RequestsTotal < 0 || input.RequestsSuccessful < 0 {
		return errors.New("provider request counters must be non-negative")
	}
	if input.RequestsSuccessful > input.RequestsTotal {
		return errors.New("successful provider requests cannot exceed total requests")
	}
	if input.ConsecutiveFailures < 0 {
		return errors.New("consecutive provider failures must be non-negative")
	}
	if input.AverageLatency < 0 {
		return errors.New("provider average latency must be non-negative")
	}
	if err := validateObservationEvidence(input.Observations); err != nil {
		return err
	}
	if err := validateBudgetEvidence(input.Budget); err != nil {
		return err
	}
	if !validRequestOutcome(input.LatestOutcome) {
		return fmt.Errorf("unsupported provider request outcome %q", input.LatestOutcome)
	}
	if input.RequestsTotal > 0 {
		if input.LastRequestAt == nil {
			return errors.New("last provider request timestamp is required when request history exists")
		}
		if input.LatestOutcome == RequestOutcomeUnknown {
			return errors.New("latest provider request outcome is required when request history exists")
		}
	}
	if input.RequestsSuccessful > 0 && input.LastSuccessAt == nil {
		return errors.New("last provider success timestamp is required when successful requests exist")
	}
	if input.FirstRequestAt != nil && input.LastRequestAt != nil && input.FirstRequestAt.After(*input.LastRequestAt) {
		return errors.New("first provider request timestamp cannot be after last request timestamp")
	}

	return nil
}

func validateObservationEvidence(evidence ObservationEvidence) error {
	if evidence.Received < 0 || evidence.Accepted < 0 || evidence.Rejected < 0 {
		return errors.New("provider observation counters must be non-negative")
	}
	if evidence.Accepted+evidence.Rejected > evidence.Received {
		return errors.New("accepted and rejected observations cannot exceed received observations")
	}

	return nil
}

func validateBudgetEvidence(evidence BudgetEvidence) error {
	if !validBudgetState(evidence.State) {
		return fmt.Errorf("unsupported provider budget state %q", evidence.State)
	}
	if evidence.Limit < 0 || evidence.Remaining < 0 {
		return errors.New("provider budget counters must be non-negative")
	}
	if evidence.Limit > 0 && evidence.Remaining > evidence.Limit {
		return errors.New("provider budget remaining value cannot exceed its limit")
	}
	if evidence.State == BudgetStateExhausted && evidence.Remaining != 0 {
		return errors.New("exhausted provider budget must have zero remaining capacity")
	}

	return nil
}

func (policy Policy) unavailableReasons(
	input EvaluationInput,
	lastSuccessAt *time.Time,
	evaluatedAt time.Time,
) []string {
	reasons := make([]string, 0, 5)

	if input.LatestOutcome == RequestOutcomeUnauthorized {
		reasons = append(reasons, "provider_authentication_rejected")
	}
	if input.Budget.State == BudgetStateExhausted {
		reasons = append(reasons, "provider_budget_exhausted")
	}
	if input.RequestsSuccessful == 0 || lastSuccessAt == nil {
		reasons = append(reasons, "provider_has_no_successful_requests")
	} else if evaluatedAt.Sub(*lastSuccessAt) > policy.UnavailableAfter {
		reasons = append(reasons, "provider_last_success_exceeds_unavailable_threshold")
	}
	if input.ConsecutiveFailures >= policy.UnavailableConsecutiveFailures {
		reasons = append(reasons, "provider_consecutive_failure_limit_reached")
	}

	return reasons
}

func (policy Policy) degradedReasons(
	input EvaluationInput,
	lastSuccessAt *time.Time,
	evaluatedAt time.Time,
	successRatio float64,
	rejectionRatio float64,
) []string {
	reasons := make([]string, 0, 9)

	if input.RequestsTotal < policy.MinimumHealthyRequestSamples {
		reasons = append(reasons, "provider_request_sample_below_healthy_threshold")
	}
	if lastSuccessAt != nil && evaluatedAt.Sub(*lastSuccessAt) > policy.StaleAfter {
		reasons = append(reasons, "provider_last_success_is_stale")
	}
	if successRatio < policy.MinimumHealthySuccessRatio {
		reasons = append(reasons, "provider_success_ratio_below_healthy_threshold")
	}
	if input.AverageLatency > policy.MaximumHealthyAverageLatency {
		reasons = append(reasons, "provider_average_latency_above_healthy_threshold")
	}
	if input.ConsecutiveFailures > policy.MaximumHealthyConsecutiveFailures {
		reasons = append(reasons, "provider_has_recent_consecutive_failures")
	}
	if input.LatestOutcome != RequestOutcomeSuccess {
		reasons = append(reasons, "provider_latest_request_was_not_successful")
	}
	if input.Budget.State == BudgetStateConstrained {
		reasons = append(reasons, "provider_budget_is_constrained")
	}
	if input.Observations.Received > 0 && rejectionRatio > policy.MaximumHealthyRejectionRatio {
		reasons = append(reasons, "provider_observation_rejection_ratio_above_healthy_threshold")
	}

	return reasons
}

func buildLimitations(input EvaluationInput) []string {
	limitations := []string{
		"provider_health_is_policy_based_operational_evidence",
		"provider_health_does_not_measure_global_coverage",
	}

	if input.Budget.State == BudgetStateUnknown {
		limitations = append(limitations, "provider_budget_state_unknown")
	}
	if input.RequestsTotal == 0 {
		limitations = append(limitations, "provider_request_history_absent")
	}

	return uniqueSorted(limitations)
}

func normalizeBudget(
	evidence BudgetEvidence,
	evaluatedAt time.Time,
) BudgetEvidence {
	evidence.ResetsAt = utcTimePointer(
		evidence.ResetsAt,
	)

	if evidence.ResetsAt != nil &&
		!evaluatedAt.Before(*evidence.ResetsAt) &&
		(evidence.State == BudgetStateExhausted ||
			evidence.State == BudgetStateConstrained) {
		return BudgetEvidence{
			State: BudgetStateUnknown,
			Limit: evidence.Limit,
		}
	}

	return evidence
}

func utcTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	utc := value.UTC()
	return &utc
}

func ageSeconds(reference time.Time, value *time.Time) *int64 {
	if value == nil {
		return nil
	}

	age := reference.Sub(*value)
	if age < 0 {
		age = 0
	}

	seconds := int64(age / time.Second)
	return &seconds
}

func ratio(numerator int64, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}

	return clamp01(float64(numerator) / float64(denominator))
}

func round4(value float64) float64 {
	return math.Round(value*10_000) / 10_000
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func isRatio(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		result = append(result, value)
	}

	sort.Strings(result)
	return result
}

func validRequestOutcome(outcome RequestOutcome) bool {
	switch outcome {
	case RequestOutcomeUnknown,
		RequestOutcomeSuccess,
		RequestOutcomeRateLimited,
		RequestOutcomeUnauthorized,
		RequestOutcomeClientError,
		RequestOutcomeServerError,
		RequestOutcomeTimeout,
		RequestOutcomeNetworkError,
		RequestOutcomeInvalidResponse:
		return true
	default:
		return false
	}
}

func validBudgetState(state BudgetState) bool {
	switch state {
	case BudgetStateUnknown,
		BudgetStateAvailable,
		BudgetStateConstrained,
		BudgetStateExhausted:
		return true
	default:
		return false
	}
}
