package providerdecision

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfallback"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

var (
	ErrProviderRequired = errors.New(
		"provider decision provider is required",
	)
	ErrNoDecisionEvidence = errors.New(
		"provider decision evidence is not available",
	)
)

const (
	LimitationProcessLocal        = "provider decision evidence is process-local and resets on restart"
	LimitationFallbackNotObserved = "fallback selection evidence has not been observed"
)

type Evidence struct {
	Provider      providerpolicy.Provider
	RequestKey    string
	PublicationID string
	Allowed       bool
	Reason        providerbudget.DecisionReason
	RetryAt       time.Time
	DecidedAt     time.Time
}

type Snapshot struct {
	Provider       providerpolicy.Provider
	DecisionsTotal int64
	AllowedTotal   int64
	DeniedTotal    int64
	ReasonCounts   map[providerbudget.DecisionReason]int64
	Latest         Evidence

	FallbackObserved         bool
	FallbackDecisionsTotal   int64
	PrimarySelectedTotal     int64
	FallbackSelectedTotal    int64
	NoProviderAvailableTotal int64
	LatestFallback           providerfallback.Decision

	Limitations []string
}

type Recorder interface {
	RecordBudgetDecision(
		provider providerpolicy.Provider,
		requestKey string,
		publicationID string,
		decision providerbudget.Decision,
	)
}

type Collector struct {
	mu sync.RWMutex

	now func() time.Time

	states map[providerpolicy.Provider]*providerState
}

type providerState struct {
	decisionsTotal int64
	allowedTotal   int64
	deniedTotal    int64
	reasonCounts   map[providerbudget.DecisionReason]int64
	latest         Evidence

	fallbackObserved         bool
	fallbackDecisionsTotal   int64
	primarySelectedTotal     int64
	fallbackSelectedTotal    int64
	noProviderAvailableTotal int64
	latestFallback           providerfallback.Decision
}

func New(
	now func() time.Time,
) *Collector {
	if now == nil {
		now = time.Now
	}

	return &Collector{
		now: now,
		states: make(
			map[providerpolicy.Provider]*providerState,
		),
	}
}

func (
	collector *Collector,
) RecordBudgetDecision(
	provider providerpolicy.Provider,
	requestKey string,
	publicationID string,
	decision providerbudget.Decision,
) {
	if collector == nil {
		return
	}

	normalizedProvider := provider
	if normalizedProvider == "" {
		normalizedProvider = decision.Provider
	}

	evidence := Evidence{
		Provider: normalizedProvider,
		RequestKey: strings.TrimSpace(
			requestKey,
		),
		PublicationID: strings.TrimSpace(
			publicationID,
		),
		Allowed:   decision.Allowed,
		Reason:    decision.Reason,
		RetryAt:   decision.RetryAt.UTC(),
		DecidedAt: collector.now().UTC(),
	}

	collector.mu.Lock()
	defer collector.mu.Unlock()

	state := collector.state(
		normalizedProvider,
	)

	state.decisionsTotal++
	if decision.Allowed {
		state.allowedTotal++
	} else {
		state.deniedTotal++
	}

	state.reasonCounts[decision.Reason]++
	state.latest = evidence
}

func (
	collector *Collector,
) RecordFallbackDecision(
	decision providerfallback.Decision,
) {
	if collector == nil ||
		decision.PrimaryProvider == "" {
		return
	}

	normalizedDecision := decision
	normalizedDecision.DecidedAt = decision.DecidedAt.UTC()
	normalizedDecision.RetryAt = decision.RetryAt.UTC()
	normalizedDecision.ConsideredProviders = append(
		[]providerpolicy.Provider(nil),
		decision.ConsideredProviders...,
	)

	collector.mu.Lock()
	defer collector.mu.Unlock()

	state := collector.state(
		decision.PrimaryProvider,
	)

	state.fallbackObserved = true
	state.fallbackDecisionsTotal++
	state.latestFallback = normalizedDecision

	switch decision.Outcome {
	case providerfallback.OutcomePrimarySelected:
		state.primarySelectedTotal++
	case providerfallback.OutcomeFallbackSelected:
		state.fallbackSelectedTotal++
	case providerfallback.OutcomeNoProviderAvailable:
		state.noProviderAvailableTotal++
	}
}

func (
	collector *Collector,
) Snapshot(
	provider providerpolicy.Provider,
) (Snapshot, error) {
	if provider == "" {
		return Snapshot{}, ErrProviderRequired
	}

	collector.mu.RLock()
	defer collector.mu.RUnlock()

	state := collector.states[provider]
	if state == nil {
		return Snapshot{}, fmt.Errorf(
			"%w: %s",
			ErrNoDecisionEvidence,
			provider,
		)
	}

	reasonCounts := make(
		map[providerbudget.DecisionReason]int64,
		len(state.reasonCounts),
	)
	for reason, count := range state.reasonCounts {
		reasonCounts[reason] = count
	}

	latestFallback := state.latestFallback
	latestFallback.ConsideredProviders = append(
		[]providerpolicy.Provider(nil),
		state.latestFallback.ConsideredProviders...,
	)

	limitations := []string{
		LimitationProcessLocal,
	}
	if !state.fallbackObserved {
		limitations = append(
			limitations,
			LimitationFallbackNotObserved,
		)
	}

	return Snapshot{
		Provider:                 provider,
		DecisionsTotal:           state.decisionsTotal,
		AllowedTotal:             state.allowedTotal,
		DeniedTotal:              state.deniedTotal,
		ReasonCounts:             reasonCounts,
		Latest:                   state.latest,
		FallbackObserved:         state.fallbackObserved,
		FallbackDecisionsTotal:   state.fallbackDecisionsTotal,
		PrimarySelectedTotal:     state.primarySelectedTotal,
		FallbackSelectedTotal:    state.fallbackSelectedTotal,
		NoProviderAvailableTotal: state.noProviderAvailableTotal,
		LatestFallback:           latestFallback,
		Limitations:              limitations,
	}, nil
}

func (
	collector *Collector,
) state(
	provider providerpolicy.Provider,
) *providerState {
	state := collector.states[provider]
	if state != nil {
		return state
	}

	state = &providerState{
		reasonCounts: make(
			map[providerbudget.DecisionReason]int64,
		),
	}

	collector.states[provider] = state

	return state
}
