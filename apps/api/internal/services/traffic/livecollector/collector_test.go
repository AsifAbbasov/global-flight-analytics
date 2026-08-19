package livecollector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	livetraffic "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/live"
)

type fakeLoader struct {
	source string
	states []flightstate.FlightState
	err    error
	calls  int
}

func (loader *fakeLoader) SourceName() string {
	return loader.source
}

func (loader *fakeLoader) LoadByPoint(
	context.Context,
	float64,
	float64,
	int,
) ([]flightstate.FlightState, error) {
	loader.calls++
	if loader.err != nil {
		return nil, loader.err
	}
	return append([]flightstate.FlightState(nil), loader.states...), nil
}

type fakeStore struct {
	calls      int
	lastStates []flightstate.FlightState
	result     livetraffic.UpsertResult
}

func (store *fakeStore) UpsertFlightStates(
	states []flightstate.FlightState,
	_ time.Time,
) livetraffic.UpsertResult {
	store.calls++
	store.lastStates = append([]flightstate.FlightState(nil), states...)
	return store.result
}

type retryAtError struct {
	at time.Time
}

func (err retryAtError) Error() string {
	return "retry later"
}

func (err retryAtError) RetryAtTime() time.Time {
	return err.at
}

func validConfig(loader Loader, store Store) Config {
	return Config{
		Loader:         loader,
		Store:          store,
		Targets:        []Target{{Name: "baku", Latitude: 40.4093, Longitude: 49.8671, RadiusNM: 250}},
		PollInterval:   10 * time.Second,
		RequestTimeout: 5 * time.Second,
		MaxBackoff:     time.Minute,
		JitterRatio:    0.01,
		Now: func() time.Time {
			return time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
		},
		RandomFloat64: func() float64 { return 0 },
	}
}

func TestNewRejectsMissingDependencies(t *testing.T) {
	_, err := New(Config{})
	if !errors.Is(err, ErrLoaderRequired) {
		t.Fatalf("expected ErrLoaderRequired, got %v", err)
	}

	_, err = New(Config{Loader: &fakeLoader{source: "fixture"}})
	if !errors.Is(err, ErrStoreRequired) {
		t.Fatalf("expected ErrStoreRequired, got %v", err)
	}
}

func TestTargetRejectsOutOfRangeRadius(t *testing.T) {
	err := (Target{
		Name:      "baku",
		Latitude:  40.4093,
		Longitude: 49.8671,
		RadiusNM:  251,
	}).Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestCollectOnceLoadsAndUpsertsCanonicalStates(t *testing.T) {
	loader := &fakeLoader{
		source: "fixture",
		states: []flightstate.FlightState{
			{
				ICAO24:     "4b1801",
				Latitude:   40.4,
				Longitude:  49.8,
				ObservedAt: time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC),
				SourceName: "fixture",
			},
		},
	}
	store := &fakeStore{
		result: livetraffic.UpsertResult{
			Accepted: 1,
			Sequence: 7,
		},
	}

	collector, err := New(validConfig(loader, store))
	if err != nil {
		t.Fatal(err)
	}

	result, err := collector.CollectOnce(context.Background())
	if err != nil {
		t.Fatalf("collect once: %v", err)
	}
	if loader.calls != 1 {
		t.Fatalf("loader calls = %d, want 1", loader.calls)
	}
	if store.calls != 1 {
		t.Fatalf("store calls = %d, want 1", store.calls)
	}
	if result.StatesLoaded != 1 || result.Upsert.Accepted != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	status := collector.SnapshotStatus()
	if status.SuccessfulCycles != 1 || status.FailedCycles != 0 {
		t.Fatalf("unexpected status: %+v", status)
	}
	if status.Accepted != 1 || status.StatesLoaded != 1 {
		t.Fatalf("unexpected totals: %+v", status)
	}
}

func TestCollectOnceStopsAfterProviderFailure(t *testing.T) {
	loader := &fakeLoader{
		source: "fixture",
		err:    errors.New("provider unavailable"),
	}
	store := &fakeStore{}
	config := validConfig(loader, store)
	config.Targets = []Target{
		{Name: "baku", Latitude: 40.4093, Longitude: 49.8671, RadiusNM: 250},
		{Name: "tbilisi", Latitude: 41.7151, Longitude: 44.8271, RadiusNM: 250},
	}

	collector, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	result, err := collector.CollectOnce(context.Background())
	if err == nil {
		t.Fatal("expected provider error")
	}
	if loader.calls != 1 {
		t.Fatalf("loader calls = %d, want fail-fast 1", loader.calls)
	}
	if store.calls != 0 {
		t.Fatalf("store calls = %d, want 0", store.calls)
	}
	if result.TargetsAttempted != 1 {
		t.Fatalf("targets attempted = %d, want 1", result.TargetsAttempted)
	}

	status := collector.SnapshotStatus()
	if status.FailedCycles != 1 || status.ConsecutiveFailures != 1 {
		t.Fatalf("unexpected failure status: %+v", status)
	}
}

func TestFailureDelayHonorsProviderRetryAt(t *testing.T) {
	now := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	loader := &fakeLoader{source: "fixture"}
	store := &fakeStore{}
	config := validConfig(loader, store)
	config.Now = func() time.Time { return now }
	config.RandomFloat64 = func() float64 { return 0 }

	collector, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	delay := collector.failureDelay(
		retryAtError{at: now.Add(3 * time.Minute)},
		2,
	)
	if delay != 3*time.Minute {
		t.Fatalf("delay = %s, want 3m", delay)
	}
}

func TestFailureDelayUsesCappedExponentialBackoff(t *testing.T) {
	loader := &fakeLoader{source: "fixture"}
	store := &fakeStore{}
	config := validConfig(loader, store)
	config.PollInterval = 10 * time.Second
	config.MaxBackoff = 40 * time.Second
	config.RandomFloat64 = func() float64 { return 0 }

	collector, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	if got := collector.failureDelay(errors.New("x"), 1); got != 10*time.Second {
		t.Fatalf("first delay = %s", got)
	}
	if got := collector.failureDelay(errors.New("x"), 2); got != 20*time.Second {
		t.Fatalf("second delay = %s", got)
	}
	if got := collector.failureDelay(errors.New("x"), 3); got != 40*time.Second {
		t.Fatalf("third delay = %s", got)
	}
	if got := collector.failureDelay(errors.New("x"), 8); got != 40*time.Second {
		t.Fatalf("capped delay = %s", got)
	}
}
