package providerhealth

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	providerhealthdomain "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerhealth"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
)

var (
	ErrProviderRequired = errors.New(
		"provider health collector provider is required",
	)
	ErrInvalidObservationEvidence = errors.New(
		"provider health observation evidence is invalid",
	)
	ErrUnsupportedTransportFailure = errors.New(
		"unsupported provider transport failure outcome",
	)
	ErrInvalidLatency = errors.New(
		"provider request latency must be non-negative",
	)
)

type Collector struct {
	mu sync.Mutex

	now    func() time.Time
	policy providerhealthdomain.Policy
	states map[providerpolicy.Provider]*providerState
}

type providerState struct {
	firstRequestAt time.Time
	lastRequestAt  time.Time
	lastSuccessAt  time.Time
	lastFailureAt  time.Time

	requestsTotal       int64
	requestsSuccessful  int64
	consecutiveFailures int
	latestOutcome       providerhealthdomain.RequestOutcome

	totalLatency   time.Duration
	latencySamples int64

	observations                providerhealthdomain.ObservationEvidence
	observationEvidenceObserved bool
	budget                      providerhealthdomain.BudgetEvidence
}

func New(now func() time.Time) *Collector {
	if now == nil {
		now = time.Now
	}

	return &Collector{
		now:    now,
		policy: DefaultPolicy(),
		states: make(map[providerpolicy.Provider]*providerState),
	}
}

func DefaultPolicy() providerhealthdomain.Policy {
	return providerhealthdomain.Policy{
		StaleAfter:                        2 * time.Minute,
		UnavailableAfter:                  10 * time.Minute,
		MinimumHealthyRequestSamples:      5,
		MinimumHealthySuccessRatio:        providerhealthdomain.MustBasisPointsFromRatio(0.90),
		MaximumHealthyAverageLatency:      5 * time.Second,
		MaximumHealthyConsecutiveFailures: 1,
		UnavailableConsecutiveFailures:    3,
		MaximumHealthyRejectionRatio:      providerhealthdomain.MustBasisPointsFromRatio(0.20),
	}
}

func (collector *Collector) RecordHTTPResponse(
	observation providerresponse.Observation,
	latency time.Duration,
) error {
	provider := observation.Provider
	if strings.TrimSpace(string(provider)) == "" {
		return ErrProviderRequired
	}
	if latency < 0 {
		return ErrInvalidLatency
	}

	now := collector.now().UTC()
	outcome := outcomeFromHTTPStatus(
		observation.StatusCode,
	)

	collector.mu.Lock()
	defer collector.mu.Unlock()

	state := collector.state(provider)
	state.recordRequest(
		now,
		outcome,
		latency,
		true,
	)
	state.budget = budgetFromHTTPObservation(
		state.budget,
		observation,
		now,
		outcome,
	)

	return nil
}

func (collector *Collector) RecordTransportFailure(
	provider providerpolicy.Provider,
	outcome providerhealthdomain.RequestOutcome,
	latency time.Duration,
) error {
	if strings.TrimSpace(string(provider)) == "" {
		return ErrProviderRequired
	}
	if outcome != providerhealthdomain.RequestOutcomeTimeout &&
		outcome != providerhealthdomain.RequestOutcomeNetworkError {
		return ErrUnsupportedTransportFailure
	}
	if latency < 0 {
		return ErrInvalidLatency
	}

	now := collector.now().UTC()

	collector.mu.Lock()
	defer collector.mu.Unlock()

	collector.state(provider).recordRequest(
		now,
		outcome,
		latency,
		true,
	)

	return nil
}

func (collector *Collector) RecordResponseFailure(
	provider providerpolicy.Provider,
	latency time.Duration,
) error {
	if strings.TrimSpace(string(provider)) == "" {
		return ErrProviderRequired
	}
	if latency < 0 {
		return ErrInvalidLatency
	}

	now := collector.now().UTC()

	collector.mu.Lock()
	defer collector.mu.Unlock()

	collector.state(provider).recordRequest(
		now,
		providerhealthdomain.RequestOutcomeInvalidResponse,
		latency,
		true,
	)

	return nil
}

func (collector *Collector) RecordObservationEvidence(
	provider providerpolicy.Provider,
	received int64,
	accepted int64,
	rejected int64,
) error {
	if strings.TrimSpace(string(provider)) == "" {
		return ErrProviderRequired
	}
	if received < 0 || accepted < 0 || rejected < 0 ||
		accepted+rejected > received {
		return ErrInvalidObservationEvidence
	}

	collector.mu.Lock()
	defer collector.mu.Unlock()

	state := collector.state(provider)
	state.observationEvidenceObserved = true
	state.observations.Received += received
	state.observations.Accepted += accepted
	state.observations.Rejected += rejected

	return nil
}

func (collector *Collector) Snapshot(
	provider providerpolicy.Provider,
) (providerhealthdomain.Snapshot, error) {
	providerName := strings.TrimSpace(
		string(provider),
	)
	if providerName == "" {
		return providerhealthdomain.Snapshot{},
			ErrProviderRequired
	}

	evaluatedAt := collector.now().UTC()

	collector.mu.Lock()
	state, exists := collector.states[provider]
	input, latencyObserved, observationEvidenceObserved := evaluationInput(
		providerName,
		evaluatedAt,
		state,
		exists,
	)
	collector.mu.Unlock()

	snapshot, err := collector.policy.Evaluate(input)
	if err != nil {
		return providerhealthdomain.Snapshot{}, err
	}

	snapshot.Limitations = append(
		snapshot.Limitations,
		"provider_health_history_is_process_local",
	)
	if !latencyObserved {
		snapshot.Limitations = append(
			snapshot.Limitations,
			"provider_request_latency_not_observed",
		)
	}
	if !observationEvidenceObserved {
		snapshot.Limitations = append(
			snapshot.Limitations,
			"provider_observation_quality_not_observed",
		)
	}

	snapshot.Limitations = uniqueSorted(
		snapshot.Limitations,
	)

	return snapshot, nil
}

func (collector *Collector) state(
	provider providerpolicy.Provider,
) *providerState {
	state := collector.states[provider]
	if state != nil {
		return state
	}

	state = &providerState{
		latestOutcome: providerhealthdomain.RequestOutcomeUnknown,
		budget: providerhealthdomain.BudgetEvidence{
			State: providerhealthdomain.BudgetStateUnknown,
		},
	}
	collector.states[provider] = state

	return state
}

func (state *providerState) recordRequest(
	observedAt time.Time,
	outcome providerhealthdomain.RequestOutcome,
	latency time.Duration,
	latencyObserved bool,
) {
	if state.firstRequestAt.IsZero() {
		state.firstRequestAt = observedAt
	}

	state.lastRequestAt = observedAt
	state.requestsTotal++
	state.latestOutcome = outcome

	if latencyObserved {
		state.totalLatency += latency
		state.latencySamples++
	}

	if outcome == providerhealthdomain.RequestOutcomeSuccess {
		state.requestsSuccessful++
		state.consecutiveFailures = 0
		state.lastSuccessAt = observedAt
		return
	}

	state.consecutiveFailures++
	state.lastFailureAt = observedAt
}

func evaluationInput(
	providerName string,
	evaluatedAt time.Time,
	state *providerState,
	exists bool,
) (providerhealthdomain.EvaluationInput, bool, bool) {
	if !exists || state == nil {
		return providerhealthdomain.EvaluationInput{
			ProviderName:  providerName,
			EvaluatedAt:   evaluatedAt,
			LatestOutcome: providerhealthdomain.RequestOutcomeUnknown,
			Budget: providerhealthdomain.BudgetEvidence{
				State: providerhealthdomain.BudgetStateUnknown,
			},
		}, false, false
	}

	input := providerhealthdomain.EvaluationInput{
		ProviderName:        providerName,
		EvaluatedAt:         evaluatedAt,
		FirstRequestAt:      timePointer(state.firstRequestAt),
		LastRequestAt:       timePointer(state.lastRequestAt),
		LastSuccessAt:       timePointer(state.lastSuccessAt),
		LastFailureAt:       timePointer(state.lastFailureAt),
		RequestsTotal:       state.requestsTotal,
		RequestsSuccessful:  state.requestsSuccessful,
		ConsecutiveFailures: state.consecutiveFailures,
		AverageLatency:      averageLatency(state),
		LatestOutcome:       state.latestOutcome,
		Observations:        state.observations,
		Budget:              state.budget,
	}

	return input,
		state.latencySamples > 0,
		state.observationEvidenceObserved
}

func averageLatency(state *providerState) time.Duration {
	if state.latencySamples <= 0 {
		return 0
	}

	return time.Duration(
		int64(state.totalLatency) / state.latencySamples,
	)
}

func budgetFromHTTPObservation(
	current providerhealthdomain.BudgetEvidence,
	observation providerresponse.Observation,
	observedAt time.Time,
	outcome providerhealthdomain.RequestOutcome,
) providerhealthdomain.BudgetEvidence {
	if current.State == "" {
		current.State = providerhealthdomain.BudgetStateUnknown
	}

	if outcome == providerhealthdomain.RequestOutcomeRateLimited {
		current.State = providerhealthdomain.BudgetStateExhausted
		current.Remaining = 0
		current.ResetsAt = nil

		if !observation.CooldownUntil.IsZero() {
			current.ResetsAt = timePointer(
				observation.CooldownUntil,
			)
		} else if observation.RetryAfterKnown &&
			observation.RetryAfter > 0 {
			current.ResetsAt = timePointer(
				observedAt.Add(observation.RetryAfter),
			)
		}

		return current
	}

	if observation.RemainingKnown {
		current.Remaining = int64(observation.Remaining)
		current.ResetsAt = nil

		if observation.Remaining <= 0 {
			current.State = providerhealthdomain.BudgetStateExhausted
		} else {
			current.State = providerhealthdomain.BudgetStateAvailable
		}

		return current
	}

	if observation.RetryAfterKnown && observation.RetryAfter > 0 {
		current.State = providerhealthdomain.BudgetStateConstrained
		current.ResetsAt = timePointer(
			observedAt.Add(observation.RetryAfter),
		)
		return current
	}

	if outcome == providerhealthdomain.RequestOutcomeSuccess &&
		(current.State == providerhealthdomain.BudgetStateExhausted ||
			current.State == providerhealthdomain.BudgetStateConstrained) {
		return providerhealthdomain.BudgetEvidence{
			State: providerhealthdomain.BudgetStateUnknown,
		}
	}

	return current
}

func outcomeFromHTTPStatus(
	statusCode int,
) providerhealthdomain.RequestOutcome {
	switch {
	case statusCode >= http.StatusOK &&
		statusCode < http.StatusMultipleChoices:
		return providerhealthdomain.RequestOutcomeSuccess
	case statusCode == http.StatusTooManyRequests:
		return providerhealthdomain.RequestOutcomeRateLimited
	case statusCode == http.StatusUnauthorized ||
		statusCode == http.StatusForbidden:
		return providerhealthdomain.RequestOutcomeUnauthorized
	case statusCode >= http.StatusInternalServerError:
		return providerhealthdomain.RequestOutcomeServerError
	default:
		return providerhealthdomain.RequestOutcomeClientError
	}
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}

	utc := value.UTC()
	return &utc
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
