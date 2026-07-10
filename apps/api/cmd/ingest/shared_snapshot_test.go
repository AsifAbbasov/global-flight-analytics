package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/sharedsnapshot"
)

type sharedSnapshotTestTrafficSource struct{}

func (
	source *sharedSnapshotTestTrafficSource,
) LoadByPoint(
	_ context.Context,
	_ float64,
	_ float64,
	_ int,
) ([]flightstate.FlightState, error) {
	return []flightstate.FlightState{
		{
			ID:     "source-state",
			ICAO24: "abc123",
		},
	}, nil
}

type sharedSnapshotTestExecutor struct {
	values    map[providerpolicy.Provider]sharedsnapshot.Payload
	errors    map[providerpolicy.Provider]error
	providers []providerpolicy.Provider
}

func (
	executor *sharedSnapshotTestExecutor,
) Execute(
	_ context.Context,
	provider providerpolicy.Provider,
	requestKey string,
	_ ingestionorchestrator.Function[sharedsnapshot.Payload],
) (ingestionorchestrator.ExecuteResult[sharedsnapshot.Payload], error) {
	if executor != nil {
		executor.providers = append(
			executor.providers,
			provider,
		)

		if err := executor.errors[provider]; err != nil {
			return ingestionorchestrator.ExecuteResult[sharedsnapshot.Payload]{},
				err
		}

		if value, exists := executor.values[provider]; exists {
			return ingestionorchestrator.ExecuteResult[sharedsnapshot.Payload]{
				Provider:   provider,
				RequestKey: requestKey,
				Value:      value,
			}, nil
		}
	}

	return ingestionorchestrator.ExecuteResult[sharedsnapshot.Payload]{
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
	executor := &sharedSnapshotTestExecutor{
		values: map[providerpolicy.Provider]sharedsnapshot.Payload{
			providerpolicy.ProviderAirplanesLive: sharedsnapshot.NewRegionalTrafficPayload(
				[]flightstate.FlightState{
					{
						ID:     "state-1",
						ICAO24: "abc123",
					},
				},
			),
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

	if len(snapshot.Successes) != 1 {
		t.Fatalf(
			"unexpected snapshot success length: got %d, want 1",
			len(snapshot.Successes),
		)
	}

	trafficPayload, ok := snapshot.Successes[0].Payload.RegionalTraffic()
	if !ok {
		t.Fatalf(
			"expected regional traffic payload, got kind %q",
			snapshot.Successes[0].Payload.Kind(),
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
}

func TestRunSharedSnapshotDoesNotExecuteWeatherProvider(
	t *testing.T,
) {
	weatherFailure := errors.New(
		"weather provider unavailable",
	)

	executor := &sharedSnapshotTestExecutor{
		values: map[providerpolicy.Provider]sharedsnapshot.Payload{
			providerpolicy.ProviderAirplanesLive: sharedsnapshot.NewRegionalTrafficPayload(
				[]flightstate.FlightState{
					{
						ID:     "state-1",
						ICAO24: "abc123",
					},
				},
			),
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
			"expected traffic-only snapshot to ignore weather failure, got %v",
			err,
		)
	}

	if snapshot.Status != providerfanin.BatchStatusSucceeded {
		t.Fatalf(
			"unexpected snapshot status: %q",
			snapshot.Status,
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
			"unexpected executed provider: %q",
			executor.providers[0],
		)
	}
}

func TestRunSharedSnapshotWrapsPayloadKindMismatch(
	t *testing.T,
) {
	executor := &sharedSnapshotTestExecutor{
		values: map[providerpolicy.Provider]sharedsnapshot.Payload{
			providerpolicy.ProviderAirplanesLive: sharedsnapshot.NewCurrentWeatherPayload(
				domainweather.CurrentSnapshot{},
			),
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

	if !errors.Is(
		err,
		sharedsnapshot.ErrSuccessPayloadKindMismatch,
	) {
		t.Fatalf(
			"expected ErrSuccessPayloadKindMismatch, got %v",
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
