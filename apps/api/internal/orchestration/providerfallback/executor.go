package providerfallback

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/requestcoalescing"
)

var (
	ErrProviderExecutorRequired = errors.New(
		"provider fallback executor dependency is required",
	)
	ErrDecisionRecorderRequired = errors.New(
		"provider fallback decision recorder is required",
	)
	ErrAttemptFunctionRequired = errors.New(
		"provider fallback attempt function is required",
	)
	ErrAttemptRequestKeyRequired = errors.New(
		"provider fallback attempt request key is required",
	)
	ErrAccessDeniedProviderMismatch = errors.New(
		"provider fallback access denial provider mismatch",
	)
)

type ProviderExecutor[
	T requestcoalescing.Value,
] interface {
	Execute(
		ctx context.Context,
		provider providerpolicy.Provider,
		requestKey string,
		function ingestionorchestrator.Function[T],
	) (ingestionorchestrator.ExecuteResult[T], error)
}

type DecisionSelector interface {
	Select(
		candidates []Candidate,
	) (Decision, error)
}

type DecisionRecorder interface {
	RecordFallbackDecision(
		decision Decision,
	)
}

type Attempt[
	T requestcoalescing.Value,
] struct {
	Provider   providerpolicy.Provider
	RequestKey string
	Function   ingestionorchestrator.Function[T]
}

type ExecuteResult[
	T requestcoalescing.Value,
] struct {
	ProviderResult ingestionorchestrator.ExecuteResult[T]
	Decision       Decision
}

type Executor[
	T requestcoalescing.Value,
] struct {
	providerExecutor ProviderExecutor[T]
	selector         DecisionSelector
	recorder         DecisionRecorder
}

type NoProviderAvailableError struct {
	Decision Decision
}

func (
	err *NoProviderAvailableError,
) Error() string {
	if err == nil {
		return "no provider is available"
	}

	if err.Decision.RetryAt.IsZero() {
		return fmt.Sprintf(
			"no provider is available: primary=%s reason=%s",
			err.Decision.PrimaryProvider,
			err.Decision.TriggerReason,
		)
	}

	return fmt.Sprintf(
		"no provider is available: primary=%s reason=%s retry_at=%s",
		err.Decision.PrimaryProvider,
		err.Decision.TriggerReason,
		err.Decision.RetryAt.UTC().Format(
			time.RFC3339Nano,
		),
	)
}

func NewExecutor[
	T requestcoalescing.Value,
](
	providerExecutor ProviderExecutor[T],
	selector DecisionSelector,
	recorder DecisionRecorder,
) (*Executor[T], error) {
	if providerExecutor == nil {
		return nil, ErrProviderExecutorRequired
	}

	if selector == nil {
		return nil, ErrSelectorRequired
	}

	if recorder == nil {
		return nil, ErrDecisionRecorderRequired
	}

	return &Executor[T]{
		providerExecutor: providerExecutor,
		selector:         selector,
		recorder:         recorder,
	}, nil
}

func (
	executor *Executor[T],
) Execute(
	ctx context.Context,
	attempts []Attempt[T],
) (ExecuteResult[T], error) {
	if executor == nil {
		return ExecuteResult[T]{},
			ErrProviderExecutorRequired
	}

	normalizedAttempts, err := validateAttempts(
		attempts,
	)
	if err != nil {
		return ExecuteResult[T]{}, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	candidates := make(
		[]Candidate,
		0,
		len(normalizedAttempts),
	)

	for _, attempt := range normalizedAttempts {
		providerResult, executionErr :=
			executor.providerExecutor.Execute(
				ctx,
				attempt.Provider,
				attempt.RequestKey,
				attempt.Function,
			)
		if executionErr == nil {
			candidates = append(
				candidates,
				Candidate{
					Provider: attempt.Provider,
					Allowed:  true,
				},
			)

			decision, selectErr :=
				executor.selector.Select(
					candidates,
				)
			if selectErr != nil {
				return ExecuteResult[T]{},
					fmt.Errorf(
						"select successful provider: %w",
						selectErr,
					)
			}

			executor.recorder.RecordFallbackDecision(
				decision,
			)

			return ExecuteResult[T]{
				ProviderResult: providerResult,
				Decision:       decision,
			}, nil
		}

		var accessDeniedError *ingestionorchestrator.AccessDeniedError
		if !errors.As(
			executionErr,
			&accessDeniedError,
		) {
			return ExecuteResult[T]{},
				fmt.Errorf(
					"execute provider %s: %w",
					attempt.Provider,
					executionErr,
				)
		}

		if accessDeniedError.Provider !=
			attempt.Provider {
			return ExecuteResult[T]{},
				fmt.Errorf(
					"%w: attempted=%s denied=%s",
					ErrAccessDeniedProviderMismatch,
					attempt.Provider,
					accessDeniedError.Provider,
				)
		}

		candidates = append(
			candidates,
			Candidate{
				Provider: attempt.Provider,
				Allowed:  false,
				DenialReason: accessDeniedError.
					Reason,
				RetryAt: accessDeniedError.
					RetryAt,
			},
		)
	}

	decision, selectErr := executor.selector.Select(
		candidates,
	)
	if selectErr != nil {
		return ExecuteResult[T]{},
			fmt.Errorf(
				"select unavailable providers: %w",
				selectErr,
			)
	}

	executor.recorder.RecordFallbackDecision(
		decision,
	)

	return ExecuteResult[T]{},
		&NoProviderAvailableError{
			Decision: decision,
		}
}

func validateAttempts[
	T requestcoalescing.Value,
](
	attempts []Attempt[T],
) ([]Attempt[T], error) {
	if len(attempts) == 0 {
		return nil, ErrCandidatesRequired
	}

	normalizedAttempts := make(
		[]Attempt[T],
		len(attempts),
	)
	providers := make(
		map[providerpolicy.Provider]struct{},
		len(attempts),
	)

	for index, attempt := range attempts {
		if attempt.Provider == "" {
			return nil, fmt.Errorf(
				"%w: index=%d",
				ErrProviderRequired,
				index,
			)
		}

		if _, exists := providers[attempt.Provider]; exists {
			return nil, fmt.Errorf(
				"%w: %s",
				ErrDuplicateProvider,
				attempt.Provider,
			)
		}
		providers[attempt.Provider] = struct{}{}

		requestKey := strings.TrimSpace(
			attempt.RequestKey,
		)
		if requestKey == "" {
			return nil, fmt.Errorf(
				"%w: provider=%s",
				ErrAttemptRequestKeyRequired,
				attempt.Provider,
			)
		}

		if attempt.Function == nil {
			return nil, fmt.Errorf(
				"%w: provider=%s",
				ErrAttemptFunctionRequired,
				attempt.Provider,
			)
		}

		attempt.RequestKey = requestKey
		normalizedAttempts[index] = attempt
	}

	return normalizedAttempts, nil
}
