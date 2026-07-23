package providerbudget

import (
	"errors"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

const defaultProviderReportedFallbackRetryAfter = time.Minute

var ErrStateStoreRequired = errors.New(
	"provider budget state store is required",
)

type FixedWindowReservation struct {
	LimitIndex  int
	WindowStart time.Time
	WindowEnd   time.Time
	MaxRequests int
}

type StateStore interface {
	AcquireFixedWindow(
		provider providerpolicy.Provider,
		reservations []FixedWindowReservation,
		now time.Time,
	) (Decision, error)

	AcquireProviderReported(
		provider providerpolicy.Provider,
		now time.Time,
		fallbackRetryAfter time.Duration,
	) (Decision, error)

	ObserveProviderReportedBudget(
		provider providerpolicy.Provider,
		remaining int,
		retryAt time.Time,
		observedAt time.Time,
	) error
}

func NewDurable(
	store StateStore,
	now func() time.Time,
) (*Manager, error) {
	if store == nil {
		return nil, ErrStateStoreRequired
	}

	manager, err := NewWithPolicies(
		providerpolicy.All(),
		now,
	)
	if err != nil {
		return nil, err
	}

	manager.stateStore = store

	return manager, nil
}

func (manager *Manager) acquireDurableFixedWindow(
	policy providerpolicy.Policy,
) (Decision, error) {
	now := manager.now().UTC()
	reservations := make(
		[]FixedWindowReservation,
		0,
		len(policy.RequestLimits),
	)

	for limitIndex, limit := range policy.RequestLimits {
		windowStart, windowEnd, err := windowBounds(
			now,
			limit.Window,
		)
		if err != nil {
			return Decision{}, err
		}

		reservations = append(
			reservations,
			FixedWindowReservation{
				LimitIndex:  limitIndex,
				WindowStart: windowStart,
				WindowEnd:   windowEnd,
				MaxRequests: limit.MaxRequests,
			},
		)
	}

	return manager.stateStore.AcquireFixedWindow(
		policy.Provider,
		reservations,
		now,
	)
}

func (manager *Manager) providerReportedFallback() time.Duration {
	if manager.providerReportedFallbackRetryAfter <= 0 {
		return defaultProviderReportedFallbackRetryAfter
	}

	return manager.providerReportedFallbackRetryAfter
}
