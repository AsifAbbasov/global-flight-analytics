package ingestionorchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/requestcoalescing"
)

type orchestrationTestValue string

func (orchestrationTestValue) RequestCoalescingValue() {}

type budgetManagerStub struct {
	acquireCallCount            int
	acquirePublicationCallCount int

	acquireFunction func(
		provider providerpolicy.Provider,
	) (providerbudget.Decision, error)

	acquirePublicationFunction func(
		provider providerpolicy.Provider,
		publicationID string,
	) (providerbudget.Decision, error)
}

func (
	stub *budgetManagerStub,
) Acquire(
	provider providerpolicy.Provider,
) (providerbudget.Decision, error) {
	stub.acquireCallCount++

	if stub.acquireFunction != nil {
		return stub.acquireFunction(
			provider,
		)
	}

	return providerbudget.Decision{
		Provider: provider,
		Allowed:  true,
		Reason:   providerbudget.DecisionReasonAllowed,
	}, nil
}

func (
	stub *budgetManagerStub,
) AcquirePublication(
	provider providerpolicy.Provider,
	publicationID string,
) (providerbudget.Decision, error) {
	stub.acquirePublicationCallCount++

	if stub.acquirePublicationFunction != nil {
		return stub.acquirePublicationFunction(
			provider,
			publicationID,
		)
	}

	return providerbudget.Decision{
		Provider: provider,
		Allowed:  true,
		Reason:   providerbudget.DecisionReasonAllowed,
	}, nil
}

type coalescerStub struct {
	callCount int
	lastKey   string
	shared    bool
}

func (
	stub *coalescerStub,
) Do(
	ctx context.Context,
	key string,
	function requestcoalescing.Function[orchestrationTestValue],
) (requestcoalescing.Result[orchestrationTestValue], error) {
	stub.callCount++
	stub.lastKey = key

	value, err := function(
		ctx,
	)
	if err != nil {
		return requestcoalescing.Result[orchestrationTestValue]{},
			err
	}

	return requestcoalescing.Result[orchestrationTestValue]{
		Value:  value,
		Shared: stub.shared,
	}, nil
}

func TestExecuteCombinesCoalescingBudgetAndProviderExecution(
	t *testing.T,
) {
	budgetManager := &budgetManagerStub{}

	coalescer := &coalescerStub{
		shared: true,
	}

	orchestrator, err := New(
		Config[orchestrationTestValue]{
			BudgetManager: budgetManager,
			Coalescer:     coalescer,
		},
	)
	if err != nil {
		t.Fatalf(
			"create orchestrator: %v",
			err,
		)
	}

	providerExecutionCount := 0

	result, err := orchestrator.Execute(
		context.Background(),
		providerpolicy.ProviderAirplanesLive,
		"traffic:regional-snapshot",
		func(
			_ context.Context,
		) (orchestrationTestValue, error) {
			providerExecutionCount++

			return orchestrationTestValue(
				"snapshot",
			), nil
		},
	)
	if err != nil {
		t.Fatalf(
			"execute orchestration: %v",
			err,
		)
	}

	if budgetManager.acquireCallCount != 1 {
		t.Fatalf(
			"expected one budget acquire, got %d",
			budgetManager.acquireCallCount,
		)
	}

	if providerExecutionCount != 1 {
		t.Fatalf(
			"expected one provider execution, got %d",
			providerExecutionCount,
		)
	}

	if coalescer.callCount != 1 {
		t.Fatalf(
			"expected one coalescer call, got %d",
			coalescer.callCount,
		)
	}

	if coalescer.lastKey !=
		"airplanes.live:traffic:regional-snapshot" {
		t.Fatalf(
			"unexpected coalescing key: %s",
			coalescer.lastKey,
		)
	}

	if result.Value != orchestrationTestValue("snapshot") {
		t.Fatalf(
			"unexpected result value: %s",
			result.Value,
		)
	}

	if !result.Shared {
		t.Fatal(
			"expected shared result marker",
		)
	}
}

func TestExecuteStopsBeforeProviderCallWhenBudgetDenies(
	t *testing.T,
) {
	retryAt := time.Date(
		2026,
		time.July,
		4,
		20,
		0,
		0,
		0,
		time.UTC,
	)

	budgetManager := &budgetManagerStub{
		acquireFunction: func(
			provider providerpolicy.Provider,
		) (providerbudget.Decision, error) {
			return providerbudget.Decision{
				Provider: provider,
				Allowed:  false,
				Reason:   providerbudget.DecisionReasonFixedWindowExhausted,
				RetryAt:  retryAt,
			}, nil
		},
	}

	orchestrator, err := New(
		Config[orchestrationTestValue]{
			BudgetManager: budgetManager,
			Coalescer:     &coalescerStub{},
		},
	)
	if err != nil {
		t.Fatalf(
			"create orchestrator: %v",
			err,
		)
	}

	providerExecutionCount := 0

	_, err = orchestrator.Execute(
		context.Background(),
		providerpolicy.ProviderAirplanesLive,
		"traffic:regional-snapshot",
		func(
			_ context.Context,
		) (orchestrationTestValue, error) {
			providerExecutionCount++

			return orchestrationTestValue(
				"snapshot",
			), nil
		},
	)

	if err == nil {
		t.Fatal(
			"expected provider access denial",
		)
	}

	var accessDeniedError *AccessDeniedError

	if !errors.As(
		err,
		&accessDeniedError,
	) {
		t.Fatalf(
			"expected AccessDeniedError, got %v",
			err,
		)
	}

	if providerExecutionCount != 0 {
		t.Fatalf(
			"expected no provider execution, got %d",
			providerExecutionCount,
		)
	}

	if !accessDeniedError.RetryAt.Equal(
		retryAt,
	) {
		t.Fatalf(
			"expected retry at %s, got %s",
			retryAt,
			accessDeniedError.RetryAt,
		)
	}
}

func TestExecutePublicationUsesPublicationBudget(
	t *testing.T,
) {
	budgetManager := &budgetManagerStub{}

	orchestrator, err := New(
		Config[orchestrationTestValue]{
			BudgetManager: budgetManager,
			Coalescer:     &coalescerStub{},
		},
	)
	if err != nil {
		t.Fatalf(
			"create orchestrator: %v",
			err,
		)
	}

	result, err := orchestrator.ExecutePublication(
		context.Background(),
		providerpolicy.ProviderOurAirports,
		"airports:regional-import",
		"publication-a",
		func(
			_ context.Context,
		) (orchestrationTestValue, error) {
			return orchestrationTestValue(
				"import-result",
			), nil
		},
	)
	if err != nil {
		t.Fatalf(
			"execute publication orchestration: %v",
			err,
		)
	}

	if budgetManager.acquirePublicationCallCount != 1 {
		t.Fatalf(
			"expected one publication budget acquire, got %d",
			budgetManager.acquirePublicationCallCount,
		)
	}

	if result.Value != orchestrationTestValue("import-result") {
		t.Fatalf(
			"unexpected publication result: %s",
			result.Value,
		)
	}
}

func TestNewRejectsMissingDependencies(
	t *testing.T,
) {
	_, err := New(
		Config[orchestrationTestValue]{},
	)

	if !errors.Is(
		err,
		ErrBudgetManagerRequired,
	) {
		t.Fatalf(
			"expected ErrBudgetManagerRequired, got %v",
			err,
		)
	}

	_, err = New(
		Config[orchestrationTestValue]{
			BudgetManager: &budgetManagerStub{},
		},
	)

	if !errors.Is(
		err,
		ErrCoalescerRequired,
	) {
		t.Fatalf(
			"expected ErrCoalescerRequired, got %v",
			err,
		)
	}
}
