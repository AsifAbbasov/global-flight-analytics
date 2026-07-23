package providerbudget

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type durableTestStore struct {
	fixedDecision Decision
	fixedErr      error

	reportedDecision Decision
	reportedErr      error

	observedProvider  providerpolicy.Provider
	observedRemaining int
	observedRetryAt   time.Time
	observedAt        time.Time
}

func (store *durableTestStore) AcquireFixedWindow(
	_ providerpolicy.Provider,
	_ []FixedWindowReservation,
	_ time.Time,
) (Decision, error) {
	return store.fixedDecision, store.fixedErr
}

func (store *durableTestStore) AcquireProviderReported(
	_ providerpolicy.Provider,
	_ time.Time,
	_ time.Duration,
) (Decision, error) {
	return store.reportedDecision, store.reportedErr
}

func (store *durableTestStore) ObserveProviderReportedBudget(
	provider providerpolicy.Provider,
	remaining int,
	retryAt time.Time,
	observedAt time.Time,
) error {
	store.observedProvider = provider
	store.observedRemaining = remaining
	store.observedRetryAt = retryAt
	store.observedAt = observedAt
	return nil
}

func TestNewDurableRequiresStateStore(t *testing.T) {
	_, err := NewDurable(nil, nil)
	if !errors.Is(err, ErrStateStoreRequired) {
		t.Fatalf("expected ErrStateStoreRequired, got %v", err)
	}
}

func TestDurableManagerDelegatesFixedWindowAcquisition(t *testing.T) {
	retryAt := time.Date(2026, time.July, 23, 14, 0, 1, 0, time.UTC)
	store := &durableTestStore{
		fixedDecision: Decision{
			Provider: providerpolicy.ProviderAirplanesLive,
			Allowed:  false,
			Reason:   DecisionReasonFixedWindowExhausted,
			RetryAt:  retryAt,
		},
	}
	manager, err := NewDurable(store, func() time.Time {
		return retryAt.Add(-time.Second)
	})
	if err != nil {
		t.Fatalf("create durable manager: %v", err)
	}

	decision, err := manager.Acquire(providerpolicy.ProviderAirplanesLive)
	if err != nil {
		t.Fatalf("acquire fixed window: %v", err)
	}
	if decision.Allowed ||
		decision.Reason != DecisionReasonFixedWindowExhausted ||
		!decision.RetryAt.Equal(retryAt) {
		t.Fatalf("unexpected durable fixed-window decision: %+v", decision)
	}
}

func TestExhaustedProviderReportedBudgetAlwaysHasRetryAt(t *testing.T) {
	currentTime := time.Date(2026, time.July, 23, 14, 30, 0, 0, time.UTC)
	manager, err := New(func() time.Time { return currentTime })
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}
	if err := manager.ObserveProviderReportedBudget(
		providerpolicy.ProviderOpenSky,
		1,
		0,
	); err != nil {
		t.Fatalf("observe remaining budget: %v", err)
	}

	first, err := manager.Acquire(providerpolicy.ProviderOpenSky)
	if err != nil || !first.Allowed {
		t.Fatalf("first acquire = %+v, err=%v", first, err)
	}
	second, err := manager.Acquire(providerpolicy.ProviderOpenSky)
	if err != nil {
		t.Fatalf("second acquire: %v", err)
	}
	if second.Allowed {
		t.Fatal("expected exhausted provider budget to deny")
	}
	if second.Reason != DecisionReasonProviderBudgetExhausted {
		t.Fatalf("reason = %s, want provider budget exhausted", second.Reason)
	}
	expectedRetryAt := currentTime.Add(time.Minute)
	if !second.RetryAt.Equal(expectedRetryAt) {
		t.Fatalf("retry at = %s, want %s", second.RetryAt, expectedRetryAt)
	}
}

func TestDurableObservationPersistsFallbackRetryAt(t *testing.T) {
	currentTime := time.Date(2026, time.July, 23, 15, 0, 0, 0, time.UTC)
	store := &durableTestStore{}
	manager, err := NewDurable(store, func() time.Time { return currentTime })
	if err != nil {
		t.Fatalf("create durable manager: %v", err)
	}

	if err := manager.ObserveProviderReportedBudget(
		providerpolicy.ProviderOpenSky,
		0,
		0,
	); err != nil {
		t.Fatalf("observe exhausted budget: %v", err)
	}

	if store.observedProvider != providerpolicy.ProviderOpenSky ||
		store.observedRemaining != 0 ||
		!store.observedAt.Equal(currentTime) {
		t.Fatalf("unexpected persisted observation: %+v", store)
	}
	expectedRetryAt := currentTime.Add(time.Minute)
	if !store.observedRetryAt.Equal(expectedRetryAt) {
		t.Fatalf("persisted retry at = %s, want %s", store.observedRetryAt, expectedRetryAt)
	}
}
