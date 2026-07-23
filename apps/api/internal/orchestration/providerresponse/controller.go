package providerresponse

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

const maxRetryAfterSeconds = int64(^uint64(0)>>1) / int64(time.Second)

var (
	ErrBudgetManagerRequired = errors.New(
		"provider response budget manager is required",
	)

	ErrInvalidRemainingBudgetHeader = errors.New(
		"invalid provider remaining budget header",
	)

	ErrInvalidRetryAfterHeader = errors.New(
		"invalid provider retry-after header",
	)
)

type BudgetManager interface {
	Acquire(
		provider providerpolicy.Provider,
	) (providerbudget.Decision, error)

	AcquirePublication(
		provider providerpolicy.Provider,
		publicationID string,
	) (providerbudget.Decision, error)

	ObserveProviderReportedBudget(
		provider providerpolicy.Provider,
		remaining int,
		retryAfter time.Duration,
	) error
}

type Config struct {
	BudgetManager BudgetManager
	Now           func() time.Time
}

type Controller struct {
	mu sync.Mutex

	budgetManager BudgetManager
	now           func() time.Time

	cooldowns map[providerpolicy.Provider]time.Time
}

type Observation struct {
	Provider   providerpolicy.Provider
	StatusCode int

	RemainingKnown bool
	Remaining      int

	RetryAfterKnown bool
	RetryAfter      time.Duration

	CooldownUntil time.Time
}

func New(
	config Config,
) (*Controller, error) {
	if config.BudgetManager == nil {
		return nil, ErrBudgetManagerRequired
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Controller{
		budgetManager: config.BudgetManager,
		now:           now,
		cooldowns: make(
			map[providerpolicy.Provider]time.Time,
		),
	}, nil
}

func (controller *Controller) Acquire(
	provider providerpolicy.Provider,
) (providerbudget.Decision, error) {
	if retryAt, active := controller.activeCooldown(
		provider,
	); active {
		return providerbudget.Decision{
			Provider: provider,
			Allowed:  false,
			Reason:   providerbudget.DecisionReasonProviderCooldown,
			RetryAt:  retryAt,
		}, nil
	}

	return controller.budgetManager.Acquire(
		provider,
	)
}

func (controller *Controller) AcquirePublication(
	provider providerpolicy.Provider,
	publicationID string,
) (providerbudget.Decision, error) {
	if retryAt, active := controller.activeCooldown(
		provider,
	); active {
		return providerbudget.Decision{
			Provider: provider,
			Allowed:  false,
			Reason:   providerbudget.DecisionReasonProviderCooldown,
			RetryAt:  retryAt,
		}, nil
	}

	return controller.budgetManager.AcquirePublication(
		provider,
		publicationID,
	)
}

func (controller *Controller) ObserveHTTPResponse(
	provider providerpolicy.Provider,
	statusCode int,
	headers http.Header,
) (Observation, error) {
	policy, err := providerpolicy.Get(
		provider,
	)
	if err != nil {
		return Observation{}, err
	}

	now := controller.now().UTC()

	observation := Observation{
		Provider:   provider,
		StatusCode: statusCode,
	}

	remaining, remainingKnown, err := readRemainingBudget(
		policy,
		headers,
	)
	if err != nil {
		return Observation{}, err
	}

	observation.Remaining = remaining
	observation.RemainingKnown = remainingKnown

	retryAfter, retryAfterKnown, err := readRetryAfter(
		policy,
		headers,
		now,
	)
	if err != nil {
		return Observation{}, err
	}

	observation.RetryAfter = retryAfter
	observation.RetryAfterKnown = retryAfterKnown

	if statusCode == http.StatusTooManyRequests &&
		retryAfterKnown &&
		retryAfter > 0 {
		cooldownUntil := now.Add(
			retryAfter,
		)

		controller.setCooldown(
			provider,
			cooldownUntil,
		)

		observation.CooldownUntil = cooldownUntil
	}

	if policy.BudgetMode == providerpolicy.BudgetModeProviderReported {
		shouldObserveBudget := remainingKnown ||
			statusCode == http.StatusTooManyRequests

		if shouldObserveBudget {
			observedRemaining := remaining

			if !remainingKnown {
				observedRemaining = 0
			}

			observedRetryAfter := time.Duration(0)

			if retryAfterKnown {
				observedRetryAfter = retryAfter
			}

			err := controller.budgetManager.ObserveProviderReportedBudget(
				provider,
				observedRemaining,
				observedRetryAfter,
			)
			if err != nil {
				return Observation{}, fmt.Errorf(
					"observe provider-reported budget: %w",
					err,
				)
			}
		}
	}

	return observation, nil
}

func (controller *Controller) activeCooldown(
	provider providerpolicy.Provider,
) (time.Time, bool) {
	now := controller.now().UTC()

	controller.mu.Lock()
	defer controller.mu.Unlock()

	retryAt, exists := controller.cooldowns[provider]
	if !exists {
		return time.Time{}, false
	}

	if !now.Before(retryAt) {
		delete(
			controller.cooldowns,
			provider,
		)

		return time.Time{}, false
	}

	return retryAt, true
}

func (controller *Controller) setCooldown(
	provider providerpolicy.Provider,
	retryAt time.Time,
) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	currentRetryAt := controller.cooldowns[provider]

	if currentRetryAt.After(retryAt) {
		return
	}

	controller.cooldowns[provider] = retryAt.UTC()
}

func readRemainingBudget(
	policy providerpolicy.Policy,
	headers http.Header,
) (int, bool, error) {
	if policy.BudgetMode != providerpolicy.BudgetModeProviderReported {
		return 0, false, nil
	}

	budget := policy.ProviderReportedBudget
	if budget == nil {
		return 0, false, nil
	}

	value := strings.TrimSpace(
		headers.Get(
			budget.RemainingHeader,
		),
	)

	if value == "" {
		return 0, false, nil
	}

	remaining, err := strconv.Atoi(
		value,
	)
	if err != nil || remaining < 0 {
		return 0, false, fmt.Errorf(
			"%w: provider=%s header=%s value=%q",
			ErrInvalidRemainingBudgetHeader,
			policy.Provider,
			budget.RemainingHeader,
			value,
		)
	}

	return remaining, true, nil
}

func readRetryAfter(
	policy providerpolicy.Policy,
	headers http.Header,
	now time.Time,
) (time.Duration, bool, error) {
	if policy.BudgetMode == providerpolicy.BudgetModeProviderReported &&
		policy.ProviderReportedBudget != nil {
		providerHeader := strings.TrimSpace(
			headers.Get(
				policy.ProviderReportedBudget.RetryAfterSecondsHeader,
			),
		)

		if providerHeader != "" {
			duration, err := parseRetryAfterSeconds(
				providerHeader,
			)
			if err != nil {
				return 0, false, fmt.Errorf(
					"%w: provider=%s header=%s value=%q",
					ErrInvalidRetryAfterHeader,
					policy.Provider,
					policy.ProviderReportedBudget.RetryAfterSecondsHeader,
					providerHeader,
				)
			}

			return duration, true, nil
		}
	}

	standardHeader := strings.TrimSpace(
		headers.Get(
			"Retry-After",
		),
	)

	if standardHeader == "" {
		return 0, false, nil
	}

	duration, err := parseStandardRetryAfter(
		standardHeader,
		now,
	)
	if err != nil {
		return 0, false, fmt.Errorf(
			"%w: provider=%s header=Retry-After value=%q",
			ErrInvalidRetryAfterHeader,
			policy.Provider,
			standardHeader,
		)
	}

	return duration, true, nil
}

func parseRetryAfterSeconds(
	value string,
) (time.Duration, error) {
	seconds, err := strconv.ParseInt(
		strings.TrimSpace(value),
		10,
		64,
	)
	if err != nil {
		return 0, err
	}

	if seconds < 0 || seconds > maxRetryAfterSeconds {
		return 0, ErrInvalidRetryAfterHeader
	}

	return time.Duration(seconds) * time.Second, nil
}

func parseStandardRetryAfter(
	value string,
	now time.Time,
) (time.Duration, error) {
	trimmedValue := strings.TrimSpace(
		value,
	)

	seconds, err := strconv.ParseInt(
		trimmedValue,
		10,
		64,
	)
	if err == nil {
		if seconds < 0 || seconds > maxRetryAfterSeconds {
			return 0, ErrInvalidRetryAfterHeader
		}

		return time.Duration(seconds) * time.Second, nil
	}

	retryAt, err := http.ParseTime(
		trimmedValue,
	)
	if err != nil {
		return 0, err
	}

	retryAfter := retryAt.UTC().Sub(
		now.UTC(),
	)

	if retryAfter < 0 {
		return 0, nil
	}

	return retryAfter, nil
}
