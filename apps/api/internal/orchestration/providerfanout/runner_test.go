package providerfanout

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

const testWatchdog = time.Second

type executorStub struct {
	executeFunction func(
		ctx context.Context,
		provider providerpolicy.Provider,
		requestKey string,
		function ingestionorchestrator.Function,
	) (ingestionorchestrator.ExecuteResult, error)
}

func (stub *executorStub) Execute(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	function ingestionorchestrator.Function,
) (ingestionorchestrator.ExecuteResult, error) {
	return stub.executeFunction(
		ctx,
		provider,
		requestKey,
		function,
	)
}

func TestRunExecutesIndependentTasksConcurrently(
	t *testing.T,
) {
	startedTasks := make(
		chan string,
		2,
	)

	releaseTasks := make(
		chan struct{},
	)

	executor := &executorStub{
		executeFunction: func(
			_ context.Context,
			provider providerpolicy.Provider,
			requestKey string,
			_ ingestionorchestrator.Function,
		) (ingestionorchestrator.ExecuteResult, error) {
			startedTasks <- requestKey

			<-releaseTasks

			return ingestionorchestrator.ExecuteResult{
				Provider:   provider,
				RequestKey: requestKey,
				Value:      requestKey,
			}, nil
		},
	}

	runner, err := New(
		executor,
	)
	if err != nil {
		t.Fatalf(
			"create provider fan-out runner: %v",
			err,
		)
	}

	resultChannel := make(
		chan []Result,
		1,
	)

	errorChannel := make(
		chan error,
		1,
	)

	go func() {
		results, runErr := runner.Run(
			context.Background(),
			[]Task{
				{
					ID:         "traffic",
					Provider:   providerpolicy.ProviderAirplanesLive,
					RequestKey: "traffic:regional-snapshot",
				},
				{
					ID:         "weather",
					Provider:   providerpolicy.ProviderOpenMeteo,
					RequestKey: "weather:regional-context",
				},
			},
		)

		resultChannel <- results
		errorChannel <- runErr
	}()

	firstStartedTask := waitForStartedTask(
		t,
		startedTasks,
	)

	secondStartedTask := waitForStartedTask(
		t,
		startedTasks,
	)

	if firstStartedTask == secondStartedTask {
		t.Fatalf(
			"expected two independent tasks, got duplicate start signal %s",
			firstStartedTask,
		)
	}

	close(releaseTasks)

	results := <-resultChannel
	runErr := <-errorChannel

	if runErr != nil {
		t.Fatalf(
			"run provider fan-out: %v",
			runErr,
		)
	}

	if len(results) != 2 {
		t.Fatalf(
			"expected two results, got %d",
			len(results),
		)
	}

	if results[0].TaskID != "traffic" {
		t.Fatalf(
			"expected first result to preserve traffic task order, got %s",
			results[0].TaskID,
		)
	}

	if results[1].TaskID != "weather" {
		t.Fatalf(
			"expected second result to preserve weather task order, got %s",
			results[1].TaskID,
		)
	}
}

func TestRunPreservesSuccessfulResultsWhenOneTaskFails(
	t *testing.T,
) {
	providerFailure := errors.New(
		"provider failure",
	)

	executor := &executorStub{
		executeFunction: func(
			_ context.Context,
			provider providerpolicy.Provider,
			requestKey string,
			_ ingestionorchestrator.Function,
		) (ingestionorchestrator.ExecuteResult, error) {
			if requestKey == "traffic:regional-snapshot" {
				return ingestionorchestrator.ExecuteResult{},
					providerFailure
			}

			return ingestionorchestrator.ExecuteResult{
				Provider:   provider,
				RequestKey: requestKey,
				Value:      "weather-snapshot",
			}, nil
		},
	}

	runner, err := New(
		executor,
	)
	if err != nil {
		t.Fatalf(
			"create provider fan-out runner: %v",
			err,
		)
	}

	results, err := runner.Run(
		context.Background(),
		[]Task{
			{
				ID:         "traffic",
				Provider:   providerpolicy.ProviderAirplanesLive,
				RequestKey: "traffic:regional-snapshot",
			},
			{
				ID:         "weather",
				Provider:   providerpolicy.ProviderOpenMeteo,
				RequestKey: "weather:regional-context",
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"run provider fan-out: %v",
			err,
		)
	}

	if !errors.Is(
		results[0].Err,
		providerFailure,
	) {
		t.Fatalf(
			"expected traffic provider failure, got %v",
			results[0].Err,
		)
	}

	if results[1].Err != nil {
		t.Fatalf(
			"expected weather task success, got %v",
			results[1].Err,
		)
	}

	if results[1].Value != "weather-snapshot" {
		t.Fatalf(
			"unexpected weather value: %v",
			results[1].Value,
		)
	}
}

func TestRunRejectsDuplicateTaskIdentifiers(
	t *testing.T,
) {
	executor := &executorStub{
		executeFunction: func(
			_ context.Context,
			_ providerpolicy.Provider,
			_ string,
			_ ingestionorchestrator.Function,
		) (ingestionorchestrator.ExecuteResult, error) {
			return ingestionorchestrator.ExecuteResult{}, nil
		},
	}

	runner, err := New(
		executor,
	)
	if err != nil {
		t.Fatalf(
			"create provider fan-out runner: %v",
			err,
		)
	}

	_, err = runner.Run(
		context.Background(),
		[]Task{
			{
				ID:       "traffic",
				Provider: providerpolicy.ProviderAirplanesLive,
			},
			{
				ID:       "traffic",
				Provider: providerpolicy.ProviderOpenMeteo,
			},
		},
	)

	if !errors.Is(
		err,
		ErrDuplicateTaskID,
	) {
		t.Fatalf(
			"expected ErrDuplicateTaskID, got %v",
			err,
		)
	}
}

func waitForStartedTask(
	t *testing.T,
	startedTasks <-chan string,
) string {
	t.Helper()

	select {
	case taskID := <-startedTasks:
		return taskID

	case <-time.After(testWatchdog):
		t.Fatal(
			"parallel provider task did not start before test watchdog expired",
		)

		return ""
	}
}
