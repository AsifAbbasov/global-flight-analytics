package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/sharedsnapshot"
)

type sharedSnapshotTestTrafficSource struct{}

func (
	source *sharedSnapshotTestTrafficSource,
) LoadByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	return []flightstate.FlightState{
		{
			ID:     "source-state",
			ICAO24: "abc123",
		},
	}, nil
}

type sharedSnapshotTestExecutor struct {
	values    map[providerpolicy.Provider]any
	errors    map[providerpolicy.Provider]error
	providers []providerpolicy.Provider
}

func (
	executor *sharedSnapshotTestExecutor,
) Execute(
	ctx context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	function ingestionorchestrator.Function,
) (ingestionorchestrator.ExecuteResult, error) {
	if executor != nil {
		executor.providers = append(
			executor.providers,
			provider,
		)

		if err := executor.errors[provider]; err != nil {
			return ingestionorchestrator.ExecuteResult{}, err
		}

		if value, exists := executor.values[provider]; exists {
			return ingestionorchestrator.ExecuteResult{
				Provider:   provider,
				RequestKey: requestKey,
				Value:      value,
			}, nil
		}
	}

	return ingestionorchestrator.ExecuteResult{
		Provider:   provider,
		RequestKey: requestKey,
	}, nil
}

func TestRunSharedSnapshotRejectsMissingTrafficSource(
	t *testing.T,
) {
	_, err := runSharedSnapshot(
		context.Background(),
		sharedSnapshotRunConfig{
			Executor: &sharedSnapshotTestExecutor{},
		},
	)
	if err == nil {
		t.Fatal(
			"expected missing traffic source to be rejected",
		)
	}

	if !errors.Is(
		err,
		sharedsnapshot.ErrRegionalTrafficSourceRequired,
	) {
		t.Fatalf(
			"expected ErrRegionalTrafficSourceRequired, got %v",
			err,
		)
	}
}

func TestRunSharedSnapshotRejectsMissingExecutor(
	t *testing.T,
) {
	_, err := runSharedSnapshot(
		context.Background(),
		sharedSnapshotRunConfig{
			TrafficSource: &sharedSnapshotTestTrafficSource{},
		},
	)
	if err == nil {
		t.Fatal(
			"expected missing executor to be rejected",
		)
	}

	if !strings.Contains(
		err.Error(),
		"create shared snapshot runtime",
	) {
		t.Fatalf(
			"expected shared snapshot runtime creation error, got %v",
			err,
		)
	}
}

func TestRunSharedSnapshotPublishesTypedTrafficResult(
	t *testing.T,
) {
	trafficStates := []flightstate.FlightState{
		{
			ID:     "state-1",
			ICAO24: "abc123",
		},
	}

	executor := &sharedSnapshotTestExecutor{
		values: map[providerpolicy.Provider]any{
			providerpolicy.ProviderAirplanesLive: trafficStates,
		},
	}

	snapshot, err := runSharedSnapshot(
		context.Background(),
		sharedSnapshotRunConfig{
			Executor:      executor,
			TrafficSource: &sharedSnapshotTestTrafficSource{},
			Latitude:      40.4093,
			Longitude:     49.8671,
			Radius:        100,
		},
	)
	if err != nil {
		t.Fatalf(
			"run shared snapshot: %v",
			err,
		)
	}

	if snapshot.Status != providerfanin.BatchStatusSucceeded {
		t.Fatalf(
			"unexpected snapshot status: got %q, want %q",
			snapshot.Status,
			providerfanin.BatchStatusSucceeded,
		)
	}

	if snapshot.TotalCount != 1 {
		t.Fatalf(
			"unexpected total count: got %d, want 1",
			snapshot.TotalCount,
		)
	}

	if snapshot.SuccessCount != 1 {
		t.Fatalf(
			"unexpected success count: got %d, want 1",
			snapshot.SuccessCount,
		)
	}

	if snapshot.FailureCount != 0 {
		t.Fatalf(
			"unexpected failure count: got %d, want 0",
			snapshot.FailureCount,
		)
	}

	if len(snapshot.Successes) != 1 {
		t.Fatalf(
			"unexpected snapshot success length: got %d, want 1",
			len(snapshot.Successes),
		)
	}

	success := snapshot.Successes[0]

	if success.TaskID != sharedsnapshot.TaskIDRegionalTraffic {
		t.Fatalf(
			"unexpected task identifier: got %q, want %q",
			success.TaskID,
			sharedsnapshot.TaskIDRegionalTraffic,
		)
	}

	trafficPayload, ok := success.Payload.(sharedsnapshot.RegionalTrafficPayload)
	if !ok {
		t.Fatalf(
			"unexpected regional traffic payload type: %T",
			success.Payload,
		)
	}

	if len(trafficPayload.States) != 1 {
		t.Fatalf(
			"unexpected traffic state count: got %d, want 1",
			len(trafficPayload.States),
		)
	}

	if trafficPayload.States[0].ID != "state-1" {
		t.Fatalf(
			"unexpected traffic state identifier: %q",
			trafficPayload.States[0].ID,
		)
	}

	if trafficPayload.States[0].ICAO24 != "abc123" {
		t.Fatalf(
			"unexpected traffic ICAO24: %q",
			trafficPayload.States[0].ICAO24,
		)
	}
}

func TestRunSharedSnapshotDoesNotExecuteWeatherProvider(
	t *testing.T,
) {
	weatherFailure := errors.New(
		"weather provider unavailable",
	)

	executor := &sharedSnapshotTestExecutor{
		values: map[providerpolicy.Provider]any{
			providerpolicy.ProviderAirplanesLive: []flightstate.FlightState{
				{
					ID:     "state-1",
					ICAO24: "abc123",
				},
			},
		},
		errors: map[providerpolicy.Provider]error{
			providerpolicy.ProviderOpenMeteo: weatherFailure,
		},
	}

	snapshot, err := runSharedSnapshot(
		context.Background(),
		sharedSnapshotRunConfig{
			Executor:      executor,
			TrafficSource: &sharedSnapshotTestTrafficSource{},
			Latitude:      40.4093,
			Longitude:     49.8671,
			Radius:        100,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected traffic-only shared snapshot to ignore weather provider failure, got %v",
			err,
		)
	}

	if snapshot.Status != providerfanin.BatchStatusSucceeded {
		t.Fatalf(
			"unexpected snapshot status: got %q, want %q",
			snapshot.Status,
			providerfanin.BatchStatusSucceeded,
		)
	}

	if snapshot.TotalCount != 1 {
		t.Fatalf(
			"unexpected total count: got %d, want 1",
			snapshot.TotalCount,
		)
	}

	if snapshot.SuccessCount != 1 {
		t.Fatalf(
			"unexpected success count: got %d, want 1",
			snapshot.SuccessCount,
		)
	}

	if snapshot.FailureCount != 0 {
		t.Fatalf(
			"unexpected failure count: got %d, want 0",
			snapshot.FailureCount,
		)
	}

	if len(executor.providers) != 1 {
		t.Fatalf(
			"unexpected executed provider count: got %d, want 1",
			len(executor.providers),
		)
	}

	if executor.providers[0] != providerpolicy.ProviderAirplanesLive {
		t.Fatalf(
			"unexpected executed provider: got %q, want %q",
			executor.providers[0],
			providerpolicy.ProviderAirplanesLive,
		)
	}

	for _, provider := range executor.providers {
		if provider == providerpolicy.ProviderOpenMeteo {
			t.Fatal(
				"expected weather provider not to be executed by traffic-only ingest snapshot",
			)
		}
	}
}

func TestRunSharedSnapshotWrapsRuntimeError(
	t *testing.T,
) {
	executor := &sharedSnapshotTestExecutor{
		values: map[providerpolicy.Provider]any{
			providerpolicy.ProviderAirplanesLive: "invalid-traffic-payload",
		},
	}

	_, err := runSharedSnapshot(
		context.Background(),
		sharedSnapshotRunConfig{
			Executor:      executor,
			TrafficSource: &sharedSnapshotTestTrafficSource{},
			Latitude:      40.4093,
			Longitude:     49.8671,
			Radius:        100,
		},
	)
	if err == nil {
		t.Fatal(
			"expected shared snapshot runtime error",
		)
	}

	if !errors.Is(
		err,
		sharedsnapshot.ErrSuccessValueTypeMismatch,
	) {
		t.Fatalf(
			"expected ErrSuccessValueTypeMismatch, got %v",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"run shared snapshot runtime",
	) {
		t.Fatalf(
			"expected shared snapshot runtime error context, got %v",
			err,
		)
	}
}
