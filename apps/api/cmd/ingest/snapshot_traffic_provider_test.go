package main

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
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

func TestNewSnapshotTrafficProviderRejectsUnexpectedTrafficPayloadKind(
	t *testing.T,
) {
	_, err := newSnapshotTrafficProvider(
		sharedsnapshot.Snapshot{
			Successes: []sharedsnapshot.Success{
				{
					TaskID: sharedsnapshot.TaskIDRegionalTraffic,
					Payload: sharedsnapshot.NewCurrentWeatherPayload(
						domainweather.CurrentSnapshot{},
					),
				},
			},
		},
		"airplanes_live",
	)

	if !errors.Is(
		err,
		errSnapshotTrafficPayloadKind,
	) {
		t.Fatalf(
			"expected errSnapshotTrafficPayloadKind, got %v",
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
					Payload: sharedsnapshot.NewRegionalTrafficPayload(
						[]flightstate.FlightState{
							{
								ID:     "state-1",
								ICAO24: "abc123",
							},
						},
					),
				},
			},
		},
		"airplanes_live",
	)

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
					Payload: sharedsnapshot.NewRegionalTrafficPayload(
						sourceStates,
					),
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
			"unexpected source name: %q",
			provider.SourceName(),
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
	}

	_, err := provider.LoadByPoint(
		ctx,
		40.4093,
		49.8671,
		100,
	)

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

func TestNilSnapshotTrafficProviderContracts(
	t *testing.T,
) {
	var provider *snapshotTrafficProvider

	if provider.SourceName() != "" {
		t.Fatalf(
			"expected empty source name for nil provider, got %q",
			provider.SourceName(),
		)
	}

	_, err := provider.LoadByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)

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
