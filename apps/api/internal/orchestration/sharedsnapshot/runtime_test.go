package sharedsnapshot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type runtimeExecutor struct {
	values map[string]any
	errors map[string]error
}

func (executor *runtimeExecutor) Execute(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	function ingestionorchestrator.Function,
) (ingestionorchestrator.ExecuteResult, error) {
	if err := executor.errors[requestKey]; err != nil {
		return ingestionorchestrator.ExecuteResult{}, err
	}

	value, exists := executor.values[requestKey]
	if exists {
		return ingestionorchestrator.ExecuteResult{
			Provider:   provider,
			RequestKey: requestKey,
			Value:      value,
		}, nil
	}

	return ingestionorchestrator.ExecuteResult{
		Provider:   provider,
		RequestKey: requestKey,
	}, nil
}

func TestRuntimeRunsTasksAndPublishesCurrentSnapshot(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		5,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	executor := &runtimeExecutor{
		values: map[string]any{
			"regional-traffic": "traffic-value",
		},
		errors: map[string]error{
			"current-weather": errors.New(
				"weather unavailable",
			),
		},
	}

	runtime, err := NewRuntime(
		RuntimeConfig{
			Executor: executor,
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create shared snapshot runtime: %v",
			err,
		)
	}

	tasks := []providerfanout.Task{
		{
			ID:         "traffic",
			Provider:   providerpolicy.ProviderAirplanesLive,
			RequestKey: "regional-traffic",
		},
		{
			ID:         "weather",
			Provider:   providerpolicy.ProviderOpenMeteo,
			RequestKey: "current-weather",
		},
	}

	publishedSnapshot, err := runtime.Run(
		context.Background(),
		tasks,
	)
	if err != nil {
		t.Fatalf(
			"run shared snapshot runtime: %v",
			err,
		)
	}

	if publishedSnapshot.TotalCount != 2 {
		t.Fatalf(
			"unexpected total count: %d",
			publishedSnapshot.TotalCount,
		)
	}

	if publishedSnapshot.SuccessCount != 1 {
		t.Fatalf(
			"unexpected success count: %d",
			publishedSnapshot.SuccessCount,
		)
	}

	if publishedSnapshot.FailureCount != 1 {
		t.Fatalf(
			"unexpected failure count: %d",
			publishedSnapshot.FailureCount,
		)
	}

	currentSnapshot, exists := runtime.Current()
	if !exists {
		t.Fatal(
			"expected current shared snapshot",
		)
	}

	if currentSnapshot.Successes[0].TaskID != "traffic" {
		t.Fatalf(
			"unexpected success task identifier: %q",
			currentSnapshot.Successes[0].TaskID,
		)
	}

	if currentSnapshot.Failures[0].TaskID != "weather" {
		t.Fatalf(
			"unexpected failure task identifier: %q",
			currentSnapshot.Failures[0].TaskID,
		)
	}
}
