package requestcoalescing

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestSameKeySharesOneInFlightExecution(
	t *testing.T,
) {
	group := New[string]()

	operationStarted := make(
		chan struct{},
	)

	releaseOperation := make(
		chan struct{},
	)

	var executionCount atomic.Int32

	function := func(
		_ context.Context,
	) (string, error) {
		executionCount.Add(1)

		close(operationStarted)

		<-releaseOperation

		return "shared-snapshot", nil
	}

	firstResultChannel := make(
		chan Result[string],
		1,
	)

	firstErrorChannel := make(
		chan error,
		1,
	)

	go func() {
		result, err := group.Do(
			context.Background(),
			"traffic:region",
			function,
		)

		firstResultChannel <- result
		firstErrorChannel <- err
	}()

	<-operationStarted

	secondResultChannel := make(
		chan Result[string],
		1,
	)

	secondErrorChannel := make(
		chan error,
		1,
	)

	go func() {
		result, err := group.Do(
			context.Background(),
			"traffic:region",
			function,
		)

		secondResultChannel <- result
		secondErrorChannel <- err
	}()

	waitForWaiterCount(
		t,
		group,
		"traffic:region",
		2,
	)

	close(releaseOperation)

	firstResult := <-firstResultChannel
	firstErr := <-firstErrorChannel

	secondResult := <-secondResultChannel
	secondErr := <-secondErrorChannel

	if firstErr != nil {
		t.Fatalf(
			"first request failed: %v",
			firstErr,
		)
	}

	if secondErr != nil {
		t.Fatalf(
			"second request failed: %v",
			secondErr,
		)
	}

	if executionCount.Load() != 1 {
		t.Fatalf(
			"expected one execution, got %d",
			executionCount.Load(),
		)
	}

	if firstResult.Value != "shared-snapshot" {
		t.Fatalf(
			"unexpected first value: %s",
			firstResult.Value,
		)
	}

	if secondResult.Value != "shared-snapshot" {
		t.Fatalf(
			"unexpected second value: %s",
			secondResult.Value,
		)
	}

	if firstResult.Shared {
		t.Fatal(
			"expected first request to lead execution",
		)
	}

	if !secondResult.Shared {
		t.Fatal(
			"expected second request to share execution",
		)
	}
}

func TestDifferentKeysExecuteIndependently(
	t *testing.T,
) {
	group := New[string]()

	var executionCount atomic.Int32

	function := func(
		_ context.Context,
	) (string, error) {
		executionCount.Add(1)

		return "snapshot", nil
	}

	_, err := group.Do(
		context.Background(),
		"traffic:region-a",
		function,
	)
	if err != nil {
		t.Fatalf(
			"first key failed: %v",
			err,
		)
	}

	_, err = group.Do(
		context.Background(),
		"traffic:region-b",
		function,
	)
	if err != nil {
		t.Fatalf(
			"second key failed: %v",
			err,
		)
	}

	if executionCount.Load() != 2 {
		t.Fatalf(
			"expected two executions, got %d",
			executionCount.Load(),
		)
	}
}

func TestLastCanceledWaiterCancelsSharedOperation(
	t *testing.T,
) {
	group := New[string]()

	requestContext, cancelRequest := context.WithCancel(
		context.Background(),
	)

	operationStarted := make(
		chan struct{},
	)

	operationCanceled := make(
		chan struct{},
	)

	resultChannel := make(
		chan error,
		1,
	)

	go func() {
		_, err := group.Do(
			requestContext,
			"weather:point",
			func(
				ctx context.Context,
			) (string, error) {
				close(operationStarted)

				<-ctx.Done()

				close(operationCanceled)

				return "", ctx.Err()
			},
		)

		resultChannel <- err
	}()

	<-operationStarted

	cancelRequest()

	err := <-resultChannel

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected canceled waiter, got %v",
			err,
		)
	}

	select {
	case <-operationCanceled:
	case <-time.After(time.Second):
		t.Fatal(
			"shared operation was not canceled after last waiter left",
		)
	}
}

func TestDoRejectsEmptyKey(
	t *testing.T,
) {
	group := New[string]()

	_, err := group.Do(
		context.Background(),
		"   ",
		func(
			_ context.Context,
		) (string, error) {
			return "snapshot", nil
		},
	)

	if !errors.Is(
		err,
		ErrKeyRequired,
	) {
		t.Fatalf(
			"expected ErrKeyRequired, got %v",
			err,
		)
	}
}

func waitForWaiterCount[T any](
	t *testing.T,
	group *Group[T],
	key string,
	expectedCount int,
) {
	t.Helper()

	deadline := time.After(
		time.Second,
	)

	for {
		group.mu.Lock()

		currentCall := group.calls[key]
		currentCount := 0

		if currentCall != nil {
			currentCount = currentCall.waiters
		}

		group.mu.Unlock()

		if currentCount == expectedCount {
			return
		}

		select {
		case <-deadline:
			t.Fatalf(
				"expected %d waiters, got %d",
				expectedCount,
				currentCount,
			)

		default:
			runtime.Gosched()
		}
	}
}
