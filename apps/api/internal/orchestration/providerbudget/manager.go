package providerbudget

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

var (
	ErrDuplicateProviderPolicy = errors.New(
		"duplicate provider policy",
	)

	ErrPublicationIDRequired = errors.New(
		"publication identifier is required",
	)

	ErrPublicationAccessRequired = errors.New(
		"publication-driven provider requires publication access",
	)

	ErrProviderReportedModeRequired = errors.New(
		"provider-reported budget mode is required",
	)

	ErrInvalidRemainingBudget = errors.New(
		"remaining provider budget must not be negative",
	)

	ErrInvalidRetryAfter = errors.New(
		"provider retry-after duration must not be negative",
	)
)

type DecisionReason string

const (
	DecisionReasonAllowed DecisionReason = "allowed"

	DecisionReasonFixedWindowExhausted DecisionReason = "fixed-window-exhausted"

	DecisionReasonProviderBudgetExhausted DecisionReason = "provider-budget-exhausted"

	DecisionReasonProviderCooldown DecisionReason = "provider-cooldown"

	DecisionReasonPublicationAlreadyProcessed DecisionReason = "publication-already-processed"
)

type Decision struct {
	Provider providerpolicy.Provider
	Allowed  bool
	Reason   DecisionReason
	RetryAt  time.Time
}

type Manager struct {
	mu sync.Mutex

	now func() time.Time

	policies map[providerpolicy.Provider]providerpolicy.Policy

	fixedWindowCounters map[fixedWindowKey]fixedWindowCounter

	providerReportedStates map[providerpolicy.Provider]providerReportedState

	processedPublications map[providerpolicy.Provider]string
}

type fixedWindowKey struct {
	Provider   providerpolicy.Provider
	LimitIndex int
}

type fixedWindowCounter struct {
	WindowStart time.Time
	Count       int
}

type providerReportedState struct {
	RemainingKnown bool
	Remaining      int
	CooldownUntil  time.Time
}

func New(
	now func() time.Time,
) (*Manager, error) {
	return NewWithPolicies(
		providerpolicy.All(),
		now,
	)
}

func NewWithPolicies(
	policies []providerpolicy.Policy,
	now func() time.Time,
) (*Manager, error) {
	if now == nil {
		now = time.Now
	}

	policyIndex := make(
		map[providerpolicy.Provider]providerpolicy.Policy,
		len(policies),
	)

	for _, policy := range policies {
		if err := providerpolicy.Validate(
			policy,
		); err != nil {
			return nil, fmt.Errorf(
				"validate provider %s policy: %w",
				policy.Provider,
				err,
			)
		}

		if _, exists := policyIndex[policy.Provider]; exists {
			return nil, fmt.Errorf(
				"%w: %s",
				ErrDuplicateProviderPolicy,
				policy.Provider,
			)
		}

		policyIndex[policy.Provider] = policy
	}

	return &Manager{
		now:                    now,
		policies:               policyIndex,
		fixedWindowCounters:    make(map[fixedWindowKey]fixedWindowCounter),
		providerReportedStates: make(map[providerpolicy.Provider]providerReportedState),
		processedPublications:  make(map[providerpolicy.Provider]string),
	}, nil
}

func (manager *Manager) Acquire(
	provider providerpolicy.Provider,
) (Decision, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	policy, err := manager.policy(provider)
	if err != nil {
		return Decision{}, err
	}

	switch policy.BudgetMode {
	case providerpolicy.BudgetModeFixedWindow:
		return manager.acquireFixedWindow(
			policy,
		)

	case providerpolicy.BudgetModeProviderReported:
		return manager.acquireProviderReported(
			policy,
		), nil

	case providerpolicy.BudgetModePublicationDriven:
		return Decision{}, fmt.Errorf(
			"%w: %s",
			ErrPublicationAccessRequired,
			provider,
		)

	default:
		return Decision{}, fmt.Errorf(
			"unsupported provider budget mode: %s",
			policy.BudgetMode,
		)
	}
}

func (manager *Manager) ObserveProviderReportedBudget(
	provider providerpolicy.Provider,
	remaining int,
	retryAfter time.Duration,
) error {
	if remaining < 0 {
		return ErrInvalidRemainingBudget
	}

	if retryAfter < 0 {
		return ErrInvalidRetryAfter
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	policy, err := manager.policy(provider)
	if err != nil {
		return err
	}

	if policy.BudgetMode != providerpolicy.BudgetModeProviderReported {
		return fmt.Errorf(
			"%w: %s",
			ErrProviderReportedModeRequired,
			provider,
		)
	}

	now := manager.now().UTC()

	state := providerReportedState{
		RemainingKnown: true,
		Remaining:      remaining,
	}

	if retryAfter > 0 {
		state.CooldownUntil = now.Add(
			retryAfter,
		)
	}

	manager.providerReportedStates[provider] = state

	return nil
}

func (manager *Manager) AcquirePublication(
	provider providerpolicy.Provider,
	publicationID string,
) (Decision, error) {
	normalizedPublicationID := strings.TrimSpace(
		publicationID,
	)

	if normalizedPublicationID == "" {
		return Decision{}, ErrPublicationIDRequired
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()

	policy, err := manager.policy(provider)
	if err != nil {
		return Decision{}, err
	}

	if policy.BudgetMode != providerpolicy.BudgetModePublicationDriven {
		return Decision{}, fmt.Errorf(
			"provider %s is not publication-driven",
			provider,
		)
	}

	if manager.processedPublications[provider] == normalizedPublicationID {
		return Decision{
			Provider: provider,
			Allowed:  false,
			Reason:   DecisionReasonPublicationAlreadyProcessed,
		}, nil
	}

	manager.processedPublications[provider] = normalizedPublicationID

	return Decision{
		Provider: provider,
		Allowed:  true,
		Reason:   DecisionReasonAllowed,
	}, nil
}

func (manager *Manager) acquireFixedWindow(
	policy providerpolicy.Policy,
) (Decision, error) {
	now := manager.now().UTC()

	retryAt := time.Time{}

	for limitIndex, limit := range policy.RequestLimits {
		windowStart, windowEnd, err := windowBounds(
			now,
			limit.Window,
		)
		if err != nil {
			return Decision{}, err
		}

		key := fixedWindowKey{
			Provider:   policy.Provider,
			LimitIndex: limitIndex,
		}

		counter := manager.fixedWindowCounters[key]

		if !counter.WindowStart.Equal(windowStart) {
			counter = fixedWindowCounter{
				WindowStart: windowStart,
			}

			manager.fixedWindowCounters[key] = counter
		}

		if counter.Count >= limit.MaxRequests {
			if retryAt.IsZero() || windowEnd.After(retryAt) {
				retryAt = windowEnd
			}
		}
	}

	if !retryAt.IsZero() {
		return Decision{
			Provider: policy.Provider,
			Allowed:  false,
			Reason:   DecisionReasonFixedWindowExhausted,
			RetryAt:  retryAt,
		}, nil
	}

	for limitIndex := range policy.RequestLimits {
		key := fixedWindowKey{
			Provider:   policy.Provider,
			LimitIndex: limitIndex,
		}

		counter := manager.fixedWindowCounters[key]
		counter.Count++

		manager.fixedWindowCounters[key] = counter
	}

	return Decision{
		Provider: policy.Provider,
		Allowed:  true,
		Reason:   DecisionReasonAllowed,
	}, nil
}

func (manager *Manager) acquireProviderReported(
	policy providerpolicy.Policy,
) Decision {
	now := manager.now().UTC()

	state := manager.providerReportedStates[policy.Provider]

	if !state.CooldownUntil.IsZero() {
		if now.Before(state.CooldownUntil) {
			return Decision{
				Provider: policy.Provider,
				Allowed:  false,
				Reason:   DecisionReasonProviderCooldown,
				RetryAt:  state.CooldownUntil,
			}
		}

		state.CooldownUntil = time.Time{}
		state.RemainingKnown = false
		state.Remaining = 0

		manager.providerReportedStates[policy.Provider] = state
	}

	if !state.RemainingKnown {
		return Decision{
			Provider: policy.Provider,
			Allowed:  true,
			Reason:   DecisionReasonAllowed,
		}
	}

	if state.Remaining <= 0 {
		return Decision{
			Provider: policy.Provider,
			Allowed:  false,
			Reason:   DecisionReasonProviderBudgetExhausted,
		}
	}

	state.Remaining--

	manager.providerReportedStates[policy.Provider] = state

	return Decision{
		Provider: policy.Provider,
		Allowed:  true,
		Reason:   DecisionReasonAllowed,
	}
}

func (manager *Manager) policy(
	provider providerpolicy.Provider,
) (providerpolicy.Policy, error) {
	policy, exists := manager.policies[provider]
	if !exists {
		return providerpolicy.Policy{}, fmt.Errorf(
			"%w: %s",
			providerpolicy.ErrUnknownProvider,
			provider,
		)
	}

	return policy, nil
}

func windowBounds(
	now time.Time,
	window providerpolicy.Window,
) (time.Time, time.Time, error) {
	now = now.UTC()

	switch window {
	case providerpolicy.WindowSecond:
		start := now.Truncate(
			time.Second,
		)

		return start, start.Add(time.Second), nil

	case providerpolicy.WindowMinute:
		start := now.Truncate(
			time.Minute,
		)

		return start, start.Add(time.Minute), nil

	case providerpolicy.WindowHour:
		start := now.Truncate(
			time.Hour,
		)

		return start, start.Add(time.Hour), nil

	case providerpolicy.WindowDay:
		start := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0,
			0,
			0,
			0,
			time.UTC,
		)

		return start, start.AddDate(0, 0, 1), nil

	case providerpolicy.WindowMonth:
		start := time.Date(
			now.Year(),
			now.Month(),
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		)

		return start, start.AddDate(0, 1, 0), nil

	default:
		return time.Time{}, time.Time{}, fmt.Errorf(
			"unsupported provider budget window: %s",
			window,
		)
	}
}
