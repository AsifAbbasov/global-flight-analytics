package providerfallback

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type executorTestValue string

func (
	executorTestValue,
) RequestCoalescingValue() {
}

type executorResult = ingestionorchestrator.ExecuteResult[executorTestValue]

type providerExecutorStub struct {
	calls   []providerpolicy.Provider
	results map[providerpolicy.Provider]executorResult
	errors  map[providerpolicy.Provider]error
}

func (
	stub *providerExecutorStub,
) Execute(
	_ context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	_ ingestionorchestrator.Function[executorTestValue],
) (executorResult, error) {
	stub.calls = append(
		stub.calls,
		provider,
	)

	if err := stub.errors[provider]; err != nil {
		return executorResult{}, err
	}

	result := stub.results[provider]
	result.Provider = provider
	result.RequestKey = requestKey

	return result, nil
}

type fallbackRecorderStub struct {
	decisions []Decision
}

func (
	stub *fallbackRecorderStub,
) RecordFallbackDecision(
	decision Decision,
) {
	stub.decisions = append(
		stub.decisions,
		decision,
	)
}

var _ ProviderExecutor[executorTestValue] = (*ingestionorchestrator.Orchestrator[executorTestValue])(nil)

func TestExecutorUsesPrimaryProvider(
	t *testing.T,
) {
	providerExecutor := &providerExecutorStub{
		results: map[providerpolicy.Provider]executorResult{
			providerpolicy.ProviderAirplanesLive: {
				Value: executorTestValue(
					"primary",
				),
			},
		},
		errors: make(
			map[providerpolicy.Provider]error,
		),
	}

	recorder := &fallbackRecorderStub{}

	executor, err := NewExecutor[executorTestValue](
		providerExecutor,
		New(nil),
		recorder,
	)
	if err != nil {
		t.Fatalf(
			"create fallback executor: %v",
			err,
		)
	}

	result, err := executor.Execute(
		context.Background(),
		testAttempts(),
	)
	if err != nil {
		t.Fatalf(
			"execute primary provider: %v",
			err,
		)
	}

	if result.Decision.Outcome !=
		OutcomePrimarySelected {
		t.Fatalf(
			"expected %s, got %s",
			OutcomePrimarySelected,
			result.Decision.Outcome,
		)
	}

	if len(providerExecutor.calls) != 1 {
		t.Fatalf(
			"expected one provider call, got %d",
			len(providerExecutor.calls),
		)
	}

	if len(recorder.decisions) != 1 {
		t.Fatalf(
			"expected one fallback decision, got %d",
			len(recorder.decisions),
		)
	}
}

func TestExecutorUsesFallbackAfterBudgetDenial(
	t *testing.T,
) {
	retryAt := time.Date(
		2026,
		time.July,
		12,
		19,
		0,
		1,
		0,
		time.UTC,
	)

	providerExecutor := &providerExecutorStub{
		results: map[providerpolicy.Provider]executorResult{
			providerpolicy.ProviderOpenSky: {
				Value: executorTestValue(
					"fallback",
				),
			},
		},
		errors: map[providerpolicy.Provider]error{
			providerpolicy.ProviderAirplanesLive: &ingestionorchestrator.AccessDeniedError{
				Provider: providerpolicy.
					ProviderAirplanesLive,
				Reason: providerbudget.
					DecisionReasonFixedWindowExhausted,
				RetryAt: retryAt,
			},
		},
	}

	recorder := &fallbackRecorderStub{}

	executor, err := NewExecutor[executorTestValue](
		providerExecutor,
		New(nil),
		recorder,
	)
	if err != nil {
		t.Fatalf(
			"create fallback executor: %v",
			err,
		)
	}

	result, err := executor.Execute(
		context.Background(),
		testAttempts(),
	)
	if err != nil {
		t.Fatalf(
			"execute fallback provider: %v",
			err,
		)
	}

	if result.Decision.Outcome !=
		OutcomeFallbackSelected {
		t.Fatalf(
			"expected %s, got %s",
			OutcomeFallbackSelected,
			result.Decision.Outcome,
		)
	}

	if result.Decision.SelectedProvider !=
		providerpolicy.ProviderOpenSky {
		t.Fatalf(
			"unexpected selected provider: %s",
			result.Decision.SelectedProvider,
		)
	}

	if result.Decision.TriggerReason !=
		providerbudget.
			DecisionReasonFixedWindowExhausted {
		t.Fatalf(
			"unexpected fallback trigger: %s",
			result.Decision.TriggerReason,
		)
	}

	if len(providerExecutor.calls) != 2 {
		t.Fatalf(
			"expected two provider calls, got %d",
			len(providerExecutor.calls),
		)
	}
}

func TestExecutorReturnsNoProviderAvailable(
	t *testing.T,
) {
	primaryRetryAt := time.Date(
		2026,
		time.July,
		12,
		19,
		0,
		5,
		0,
		time.UTC,
	)
	fallbackRetryAt := primaryRetryAt.Add(
		-3 * time.Second,
	)

	providerExecutor := &providerExecutorStub{
		results: make(
			map[providerpolicy.Provider]executorResult,
		),
		errors: map[providerpolicy.Provider]error{
			providerpolicy.ProviderAirplanesLive: &ingestionorchestrator.AccessDeniedError{
				Provider: providerpolicy.
					ProviderAirplanesLive,
				Reason: providerbudget.
					DecisionReasonFixedWindowExhausted,
				RetryAt: primaryRetryAt,
			},
			providerpolicy.ProviderOpenSky: &ingestionorchestrator.AccessDeniedError{
				Provider: providerpolicy.
					ProviderOpenSky,
				Reason: providerbudget.
					DecisionReasonProviderCooldown,
				RetryAt: fallbackRetryAt,
			},
		},
	}

	recorder := &fallbackRecorderStub{}

	executor, err := NewExecutor[executorTestValue](
		providerExecutor,
		New(nil),
		recorder,
	)
	if err != nil {
		t.Fatalf(
			"create fallback executor: %v",
			err,
		)
	}

	_, err = executor.Execute(
		context.Background(),
		testAttempts(),
	)
	if err == nil {
		t.Fatal(
			"expected no provider available error",
		)
	}

	var noProviderError *NoProviderAvailableError
	if !errors.As(
		err,
		&noProviderError,
	) {
		t.Fatalf(
			"expected NoProviderAvailableError, got %v",
			err,
		)
	}

	if !noProviderError.Decision.RetryAt.Equal(
		fallbackRetryAt,
	) {
		t.Fatalf(
			"expected retry at %s, got %s",
			fallbackRetryAt,
			noProviderError.Decision.RetryAt,
		)
	}

	if len(recorder.decisions) != 1 {
		t.Fatalf(
			"expected one recorded decision, got %d",
			len(recorder.decisions),
		)
	}
}

func TestExecutorDoesNotFallbackAfterRuntimeFailure(
	t *testing.T,
) {
	runtimeError := errors.New(
		"provider response failed",
	)

	providerExecutor := &providerExecutorStub{
		results: make(
			map[providerpolicy.Provider]executorResult,
		),
		errors: map[providerpolicy.Provider]error{
			providerpolicy.ProviderAirplanesLive: runtimeError,
		},
	}

	recorder := &fallbackRecorderStub{}

	executor, err := NewExecutor[executorTestValue](
		providerExecutor,
		New(nil),
		recorder,
	)
	if err != nil {
		t.Fatalf(
			"create fallback executor: %v",
			err,
		)
	}

	_, err = executor.Execute(
		context.Background(),
		testAttempts(),
	)

	if !errors.Is(
		err,
		runtimeError,
	) {
		t.Fatalf(
			"expected runtime error, got %v",
			err,
		)
	}

	if len(providerExecutor.calls) != 1 {
		t.Fatalf(
			"expected one provider call, got %d",
			len(providerExecutor.calls),
		)
	}

	if len(recorder.decisions) != 0 {
		t.Fatalf(
			"expected no completed fallback decision, got %d",
			len(recorder.decisions),
		)
	}
}

func TestExecutorRejectsDuplicateAttemptsBeforeCallingProvider(
	t *testing.T,
) {
	providerExecutor := &providerExecutorStub{
		results: make(
			map[providerpolicy.Provider]executorResult,
		),
		errors: make(
			map[providerpolicy.Provider]error,
		),
	}

	executor, err := NewExecutor[executorTestValue](
		providerExecutor,
		New(nil),
		&fallbackRecorderStub{},
	)
	if err != nil {
		t.Fatalf(
			"create fallback executor: %v",
			err,
		)
	}

	attempts := testAttempts()
	attempts[1].Provider =
		providerpolicy.ProviderAirplanesLive

	_, err = executor.Execute(
		context.Background(),
		attempts,
	)

	if !errors.Is(
		err,
		ErrDuplicateProvider,
	) {
		t.Fatalf(
			"expected ErrDuplicateProvider, got %v",
			err,
		)
	}

	if len(providerExecutor.calls) != 0 {
		t.Fatalf(
			"expected no provider calls, got %d",
			len(providerExecutor.calls),
		)
	}
}

func testAttempts() []Attempt[executorTestValue] {
	return []Attempt[executorTestValue]{
		{
			Provider: providerpolicy.
				ProviderAirplanesLive,
			RequestKey: "traffic:primary",
			Function: func(
				context.Context,
			) (
				executorTestValue,
				error,
			) {
				return "primary", nil
			},
		},
		{
			Provider: providerpolicy.
				ProviderOpenSky,
			RequestKey: "traffic:fallback",
			Function: func(
				context.Context,
			) (
				executorTestValue,
				error,
			) {
				return "fallback", nil
			},
		},
	}
}
