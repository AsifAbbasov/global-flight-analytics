package sharedsnapshot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

type runtimeExecutor struct {
	values map[string]Payload
	errors map[string]error
}

func (
	executor *runtimeExecutor,
) Execute(
	_ context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	_ ingestionorchestrator.Function[Payload],
) (ingestionorchestrator.ExecuteResult[Payload], error) {
	if err := executor.errors[requestKey]; err != nil {
		return ingestionorchestrator.ExecuteResult[Payload]{},
			err
	}

	value := executor.values[requestKey]

	return ingestionorchestrator.ExecuteResult[Payload]{
		Provider:   provider,
		RequestKey: requestKey,
		Value:      value,
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
		values: map[string]Payload{
			"regional-traffic": NewRegionalTrafficPayload(
				[]flightstate.FlightState{
					{},
				},
			),
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

	tasks := []providerfanout.Task[Payload]{
		{
			ID:         TaskIDRegionalTraffic,
			Provider:   providerpolicy.ProviderAirplanesLive,
			RequestKey: "regional-traffic",
		},
		{
			ID:         TaskIDCurrentWeather,
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

	if currentSnapshot.Successes[0].TaskID !=
		TaskIDRegionalTraffic {
		t.Fatalf(
			"unexpected success task identifier: %q",
			currentSnapshot.Successes[0].TaskID,
		)
	}

	trafficPayload, ok := currentSnapshot.Successes[0].Payload.RegionalTraffic()
	if !ok {
		t.Fatalf(
			"expected regional traffic payload, got kind %q",
			currentSnapshot.Successes[0].Payload.Kind(),
		)
	}

	if len(trafficPayload.States) != 1 {
		t.Fatalf(
			"unexpected traffic state count: got %d, want 1",
			len(trafficPayload.States),
		)
	}

	if currentSnapshot.Failures[0].TaskID !=
		TaskIDCurrentWeather {
		t.Fatalf(
			"unexpected failure task identifier: %q",
			currentSnapshot.Failures[0].TaskID,
		)
	}
}
