package main

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanin"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/regionalprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/sharedsnapshot"
)

func TestNewSnapshotTrafficProviderRejectsBlankSourceName(
	t *testing.T,
) {
	_, err := newSnapshotTrafficProvider(
		sharedsnapshot.Snapshot{},
		"   ",
	)
	if err == nil {
		t.Fatal(
			"expected blank source name to be rejected",
		)
	}

	if !errors.Is(
		err,
		errSnapshotTrafficSourceNameRequired,
	) {
		t.Fatalf(
			"expected errSnapshotTrafficSourceNameRequired, got %v",
			err,
		)
	}
}

func TestNewSnapshotTrafficProviderRejectsMissingTrafficResult(
	t *testing.T,
) {
	_, err := newSnapshotTrafficProvider(
		sharedsnapshot.Snapshot{},
		"airplanes_live",
	)
	if err == nil {
		t.Fatal(
			"expected missing traffic result to be rejected",
		)
	}

	if !errors.Is(
		err,
		errSnapshotTrafficResultMissing,
	) {
		t.Fatalf(
			"expected errSnapshotTrafficResultMissing, got %v",
			err,
		)
	}
}

func TestNewSnapshotTrafficProviderPropagatesTrafficFailure(
	t *testing.T,
) {
	providerFailure := errors.New(
		"traffic provider unavailable",
	)

	_, err := newSnapshotTrafficProvider(
		sharedsnapshot.Snapshot{
			Failures: []providerfanin.Failure{
				{
					TaskID: sharedsnapshot.TaskIDRegionalTraffic,
					Err:    providerFailure,
				},
			},
		},
		"airplanes_live",
	)
	if err == nil {
		t.Fatal(
			"expected traffic failure",
		)
	}

	if !errors.Is(
		err,
		providerFailure,
	) {
		t.Fatalf(
			"expected wrapped traffic provider failure, got %v",
			err,
		)
	}
}

func TestNewSnapshotTrafficProviderRejectsTrafficFailureWithoutError(
	t *testing.T,
) {
	_, err := newSnapshotTrafficProvider(
		sharedsnapshot.Snapshot{
			Failures: []providerfanin.Failure{
				{
					TaskID: sharedsnapshot.TaskIDRegionalTraffic,
				},
			},
		},
		"airplanes_live",
	)
	if err == nil {
		t.Fatal(
			"expected missing traffic result error",
		)
	}

	if !errors.Is(
		err,
		errSnapshotTrafficResultMissing,
	) {
		t.Fatalf(
			"expected errSnapshotTrafficResultMissing, got %v",
			err,
		)
	}
}

func TestNewSnapshotTrafficProviderRejectsUnexpectedTrafficPayload(
	t *testing.T,
) {
	_, err := newSnapshotTrafficProvider(
		sharedsnapshot.Snapshot{
			Successes: []sharedsnapshot.Success{
				{
					TaskID:  sharedsnapshot.TaskIDRegionalTraffic,
					Payload: sharedsnapshot.CurrentWeatherPayload{},
				},
			},
		},
		"airplanes_live",
	)
	if err == nil {
		t.Fatal(
			"expected unexpected traffic payload type to be rejected",
		)
	}

	if !errors.Is(
		err,
		errSnapshotTrafficResultType,
	) {
		t.Fatalf(
			"expected errSnapshotTrafficResultType, got %v",
			err,
		)
	}
}

func TestNewSnapshotTrafficProviderRejectsMissingTrafficRequestKey(
	t *testing.T,
) {
	_, err := newSnapshotTrafficProvider(
		sharedsnapshot.Snapshot{
			Successes: []sharedsnapshot.Success{
				{
					TaskID: sharedsnapshot.TaskIDRegionalTraffic,
					Payload: sharedsnapshot.RegionalTrafficPayload{
						States: []flightstate.FlightState{
							{
								ID:     "state-1",
								ICAO24: "abc123",
							},
						},
					},
				},
			},
		},
		"airplanes_live",
	)
	if err == nil {
		t.Fatal(
			"expected missing traffic request key to be rejected",
		)
	}

	if !errors.Is(
		err,
		errSnapshotTrafficRequestKeyMissing,
	) {
		t.Fatalf(
			"expected errSnapshotTrafficRequestKeyMissing, got %v",
			err,
		)
	}
}

func TestNewSnapshotTrafficProviderCreatesProviderFromTrafficPayload(
	t *testing.T,
) {
	sourceStates := []flightstate.FlightState{
		{
			ID:     "state-1",
			ICAO24: "abc123",
		},
	}

	requestKey := regionalprovider.PointRequestKey(
		40.4093,
		49.8671,
		100,
	)

	provider, err := newSnapshotTrafficProvider(
		sharedsnapshot.Snapshot{
			Successes: []sharedsnapshot.Success{
				{
					TaskID:     sharedsnapshot.TaskIDRegionalTraffic,
					RequestKey: requestKey,
					Payload: sharedsnapshot.RegionalTrafficPayload{
						States: sourceStates,
					},
				},
			},
		},
		"  airplanes_live  ",
	)
	if err != nil {
		t.Fatalf(
			"create snapshot traffic provider: %v",
			err,
		)
	}

	if provider.SourceName() != "airplanes_live" {
		t.Fatalf(
			"unexpected source name: got %q, want %q",
			provider.SourceName(),
			"airplanes_live",
		)
	}

	sourceStates[0].ID = "caller-mutated"

	loadedStates, err := provider.LoadByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf(
			"load snapshot traffic states: %v",
			err,
		)
	}

	if len(loadedStates) != 1 {
		t.Fatalf(
			"unexpected loaded state count: got %d, want 1",
			len(loadedStates),
		)
	}

	if loadedStates[0].ID != "state-1" {
		t.Fatalf(
			"expected provider state to be protected from caller mutation, got %q",
			loadedStates[0].ID,
		)
	}

	loadedStates[0].ID = "reader-mutated"

	loadedAgain, err := provider.LoadByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf(
			"load snapshot traffic states again: %v",
			err,
		)
	}

	if len(loadedAgain) != 1 {
		t.Fatalf(
			"unexpected repeated loaded state count: got %d, want 1",
			len(loadedAgain),
		)
	}

	if loadedAgain[0].ID != "state-1" {
		t.Fatalf(
			"expected provider state to be protected from reader mutation, got %q",
			loadedAgain[0].ID,
		)
	}
}

func TestSnapshotTrafficProviderLoadByPointRejectsMismatchedRequest(
	t *testing.T,
) {
	provider := &snapshotTrafficProvider{
		sourceName: "airplanes_live",
		requestKey: regionalprovider.PointRequestKey(
			40.4093,
			49.8671,
			100,
		),
		states: []flightstate.FlightState{
			{
				ID: "state-1",
			},
		},
	}

	_, err := provider.LoadByPoint(
		context.Background(),
		41.7151,
		44.8271,
		100,
	)
	if err == nil {
		t.Fatal(
			"expected mismatched point request to be rejected",
		)
	}

	if !errors.Is(
		err,
		errSnapshotTrafficRequestMismatch,
	) {
		t.Fatalf(
			"expected errSnapshotTrafficRequestMismatch, got %v",
			err,
		)
	}
}

func TestSnapshotTrafficProviderLoadByPointReturnsCancelledContextError(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	provider := &snapshotTrafficProvider{
		sourceName: "airplanes_live",
		requestKey: regionalprovider.PointRequestKey(
			40.4093,
			49.8671,
			100,
		),
		states: []flightstate.FlightState{
			{
				ID: "state-1",
			},
		},
	}

	_, err := provider.LoadByPoint(
		ctx,
		40.4093,
		49.8671,
		100,
	)
	if err == nil {
		t.Fatal(
			"expected cancelled context error",
		)
	}

	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"expected context.Canceled, got %v",
			err,
		)
	}
}

func TestNilSnapshotTrafficProviderLoadByPointReturnsMissingResult(
	t *testing.T,
) {
	var provider *snapshotTrafficProvider

	_, err := provider.LoadByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err == nil {
		t.Fatal(
			"expected nil provider error",
		)
	}

	if !errors.Is(
		err,
		errSnapshotTrafficResultMissing,
	) {
		t.Fatalf(
			"expected errSnapshotTrafficResultMissing, got %v",
			err,
		)
	}
}

func TestNilSnapshotTrafficProviderSourceNameReturnsEmptyString(
	t *testing.T,
) {
	var provider *snapshotTrafficProvider

	if provider.SourceName() != "" {
		t.Fatalf(
			"expected empty source name for nil provider, got %q",
			provider.SourceName(),
		)
	}
}
