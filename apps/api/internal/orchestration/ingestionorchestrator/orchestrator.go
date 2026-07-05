package ingestionorchestrator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/requestcoalescing"
)

var (
	ErrBudgetManagerRequired = errors.New(
		"provider budget manager is required",
	)

	ErrCoalescerRequired = errors.New(
		"request coalescer is required",
	)

	ErrRequestKeyRequired = errors.New(
		"orchestration request key is required",
	)

	ErrFunctionRequired = errors.New(
		"orchestration function is required",
	)

	ErrPublicationIDRequired = errors.New(
		"publication identifier is required",
	)
)

type Function func(
	ctx context.Context,
) (any, error)

type BudgetManager interface {
	Acquire(
		provider providerpolicy.Provider,
	) (providerbudget.Decision, error)

	AcquirePublication(
		provider providerpolicy.Provider,
		publicationID string,
	) (providerbudget.Decision, error)
}

type Coalescer interface {
	Do(
		ctx context.Context,
		key string,
		function requestcoalescing.Function[any],
	) (requestcoalescing.Result[any], error)
}

type Config struct {
	BudgetManager BudgetManager
	Coalescer     Coalescer
}

type Orchestrator struct {
	budgetManager BudgetManager
	coalescer     Coalescer
}

type ExecuteResult struct {
	Provider   providerpolicy.Provider
	RequestKey string
	Value      any
	Shared     bool
}

type AccessDeniedError struct {
	Provider providerpolicy.Provider
	Reason   providerbudget.DecisionReason
	RetryAt  time.Time
}

func (err *AccessDeniedError) Error() string {
	if err == nil {
		return "provider access denied"
	}

	if err.RetryAt.IsZero() {
		return fmt.Sprintf(
			"provider access denied: provider=%s reason=%s",
			err.Provider,
			err.Reason,
		)
	}

	return fmt.Sprintf(
		"provider access denied: provider=%s reason=%s retry_at=%s",
		err.Provider,
		err.Reason,
		err.RetryAt.UTC().Format(time.RFC3339Nano),
	)
}

func New(
	config Config,
) (*Orchestrator, error) {
	if config.BudgetManager == nil {
		return nil, ErrBudgetManagerRequired
	}

	if config.Coalescer == nil {
		return nil, ErrCoalescerRequired
	}

	return &Orchestrator{
		budgetManager: config.BudgetManager,
		coalescer:     config.Coalescer,
	}, nil
}

func NewDefault(
	budgetManager BudgetManager,
) (*Orchestrator, error) {
	return New(
		Config{
			BudgetManager: budgetManager,
			Coalescer:     requestcoalescing.New[any](),
		},
	)
}

func (orchestrator *Orchestrator) Execute(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	function Function,
) (ExecuteResult, error) {
	normalizedRequestKey := strings.TrimSpace(
		requestKey,
	)

	if normalizedRequestKey == "" {
		return ExecuteResult{}, ErrRequestKeyRequired
	}

	if function == nil {
		return ExecuteResult{}, ErrFunctionRequired
	}

	if ctx == nil {
		ctx = context.Background()
	}

	result, err := orchestrator.coalescer.Do(
		ctx,
		coalescingKey(
			provider,
			normalizedRequestKey,
		),
		func(
			operationContext context.Context,
		) (any, error) {
			decision, err := orchestrator.budgetManager.Acquire(
				provider,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"acquire provider budget: %w",
					err,
				)
			}

			if !decision.Allowed {
				return nil, &AccessDeniedError{
					Provider: decision.Provider,
					Reason:   decision.Reason,
					RetryAt:  decision.RetryAt,
				}
			}

			return function(
				operationContext,
			)
		},
	)
	if err != nil {
		return ExecuteResult{}, err
	}

	return ExecuteResult{
		Provider:   provider,
		RequestKey: normalizedRequestKey,
		Value:      result.Value,
		Shared:     result.Shared,
	}, nil
}

func (orchestrator *Orchestrator) ExecutePublication(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	publicationID string,
	function Function,
) (ExecuteResult, error) {
	normalizedRequestKey := strings.TrimSpace(
		requestKey,
	)

	if normalizedRequestKey == "" {
		return ExecuteResult{}, ErrRequestKeyRequired
	}

	normalizedPublicationID := strings.TrimSpace(
		publicationID,
	)

	if normalizedPublicationID == "" {
		return ExecuteResult{}, ErrPublicationIDRequired
	}

	if function == nil {
		return ExecuteResult{}, ErrFunctionRequired
	}

	if ctx == nil {
		ctx = context.Background()
	}

	result, err := orchestrator.coalescer.Do(
		ctx,
		publicationCoalescingKey(
			provider,
			normalizedRequestKey,
			normalizedPublicationID,
		),
		func(
			operationContext context.Context,
		) (any, error) {
			decision, err := orchestrator.budgetManager.AcquirePublication(
				provider,
				normalizedPublicationID,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"acquire publication provider budget: %w",
					err,
				)
			}

			if !decision.Allowed {
				return nil, &AccessDeniedError{
					Provider: decision.Provider,
					Reason:   decision.Reason,
					RetryAt:  decision.RetryAt,
				}
			}

			return function(
				operationContext,
			)
		},
	)
	if err != nil {
		return ExecuteResult{}, err
	}

	return ExecuteResult{
		Provider:   provider,
		RequestKey: normalizedRequestKey,
		Value:      result.Value,
		Shared:     result.Shared,
	}, nil
}

func coalescingKey(
	provider providerpolicy.Provider,
	requestKey string,
) string {
	return fmt.Sprintf(
		"%s:%s",
		provider,
		requestKey,
	)
}

func publicationCoalescingKey(
	provider providerpolicy.Provider,
	requestKey string,
	publicationID string,
) string {
	return fmt.Sprintf(
		"%s:%s:%s",
		provider,
		requestKey,
		publicationID,
	)
}
