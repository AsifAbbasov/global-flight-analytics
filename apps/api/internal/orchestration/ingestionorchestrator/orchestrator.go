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

type Function[T requestcoalescing.Value] func(
	ctx context.Context,
) (T, error)

type DecisionRecorder interface {
	RecordBudgetDecision(
		provider providerpolicy.Provider,
		requestKey string,
		publicationID string,
		decision providerbudget.Decision,
	)
}

type BudgetManager interface {
	Acquire(
		provider providerpolicy.Provider,
	) (providerbudget.Decision, error)

	AcquirePublication(
		provider providerpolicy.Provider,
		publicationID string,
	) (providerbudget.Decision, error)
}

type Coalescer[T requestcoalescing.Value] interface {
	Do(
		ctx context.Context,
		key string,
		function requestcoalescing.Function[T],
	) (requestcoalescing.Result[T], error)
}

type Config[T requestcoalescing.Value] struct {
	BudgetManager    BudgetManager
	Coalescer        Coalescer[T]
	DecisionRecorder DecisionRecorder
}

type Orchestrator[T requestcoalescing.Value] struct {
	budgetManager    BudgetManager
	coalescer        Coalescer[T]
	decisionRecorder DecisionRecorder
}

type ExecuteResult[T requestcoalescing.Value] struct {
	Provider   providerpolicy.Provider
	RequestKey string
	Value      T
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
		err.RetryAt.UTC().Format(
			time.RFC3339Nano,
		),
	)
}

func New[T requestcoalescing.Value](
	config Config[T],
) (*Orchestrator[T], error) {
	if config.BudgetManager == nil {
		return nil, ErrBudgetManagerRequired
	}

	if config.Coalescer == nil {
		return nil, ErrCoalescerRequired
	}

	return &Orchestrator[T]{
		budgetManager:    config.BudgetManager,
		coalescer:        config.Coalescer,
		decisionRecorder: config.DecisionRecorder,
	}, nil
}

func NewDefault[T requestcoalescing.Value](
	budgetManager BudgetManager,
) (*Orchestrator[T], error) {
	return NewDefaultWithDecisionRecorder[T](
		budgetManager,
		nil,
	)
}

func NewDefaultWithDecisionRecorder[
	T requestcoalescing.Value,
](
	budgetManager BudgetManager,
	decisionRecorder DecisionRecorder,
) (*Orchestrator[T], error) {
	return New(
		Config[T]{
			BudgetManager:    budgetManager,
			Coalescer:        requestcoalescing.New[T](),
			DecisionRecorder: decisionRecorder,
		},
	)
}

func (
	orchestrator *Orchestrator[T],
) Execute(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	function Function[T],
) (ExecuteResult[T], error) {
	normalizedRequestKey := strings.TrimSpace(
		requestKey,
	)

	if normalizedRequestKey == "" {
		return ExecuteResult[T]{},
			ErrRequestKeyRequired
	}

	if function == nil {
		return ExecuteResult[T]{},
			ErrFunctionRequired
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
		) (T, error) {
			decision, err := orchestrator.budgetManager.Acquire(
				provider,
			)
			if err != nil {
				var zero T

				return zero, fmt.Errorf(
					"acquire provider budget: %w",
					err,
				)
			}

			orchestrator.recordBudgetDecision(
				provider,
				normalizedRequestKey,
				"",
				decision,
			)

			if !decision.Allowed {
				var zero T

				return zero, &AccessDeniedError{
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
		return ExecuteResult[T]{},
			err
	}

	return ExecuteResult[T]{
		Provider:   provider,
		RequestKey: normalizedRequestKey,
		Value:      result.Value,
		Shared:     result.Shared,
	}, nil
}

func (
	orchestrator *Orchestrator[T],
) ExecutePublication(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	publicationID string,
	function Function[T],
) (ExecuteResult[T], error) {
	normalizedRequestKey := strings.TrimSpace(
		requestKey,
	)

	if normalizedRequestKey == "" {
		return ExecuteResult[T]{},
			ErrRequestKeyRequired
	}

	normalizedPublicationID := strings.TrimSpace(
		publicationID,
	)

	if normalizedPublicationID == "" {
		return ExecuteResult[T]{},
			ErrPublicationIDRequired
	}

	if function == nil {
		return ExecuteResult[T]{},
			ErrFunctionRequired
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
		) (T, error) {
			decision, err := orchestrator.budgetManager.AcquirePublication(
				provider,
				normalizedPublicationID,
			)
			if err != nil {
				var zero T

				return zero, fmt.Errorf(
					"acquire publication provider budget: %w",
					err,
				)
			}

			orchestrator.recordBudgetDecision(
				provider,
				normalizedRequestKey,
				normalizedPublicationID,
				decision,
			)

			if !decision.Allowed {
				var zero T

				return zero, &AccessDeniedError{
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
		return ExecuteResult[T]{},
			err
	}

	return ExecuteResult[T]{
		Provider:   provider,
		RequestKey: normalizedRequestKey,
		Value:      result.Value,
		Shared:     result.Shared,
	}, nil
}

func (
	orchestrator *Orchestrator[T],
) recordBudgetDecision(
	provider providerpolicy.Provider,
	requestKey string,
	publicationID string,
	decision providerbudget.Decision,
) {
	if orchestrator.decisionRecorder == nil {
		return
	}

	orchestrator.decisionRecorder.RecordBudgetDecision(
		provider,
		requestKey,
		publicationID,
		decision,
	)
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
