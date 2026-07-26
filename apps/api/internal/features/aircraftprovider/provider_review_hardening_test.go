package aircraftprovider

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestProviderRecognizesDomainAircraftNotFoundByDefault(t *testing.T) {
	provider := newTestProvider(t, Config{
		Lookup: aircraftLookupFunc(func(
			context.Context,
			string,
		) (aircraft.Aircraft, error) {
			return aircraft.Aircraft{}, fmt.Errorf(
				"repository read: %w",
				aircraft.ErrNotFound,
			)
		}),
	})

	features, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{ICAO24: "ABC123"},
	)
	if err != nil {
		t.Fatalf("Provide() error = %v", err)
	}
	if features.Evidence.Status != flightfeatures.AvailabilityStatusUnavailable ||
		!hasLimitation(features.Evidence.Limitations, "aircraft_metadata_not_found") {
		t.Fatalf("unexpected features: %#v", features)
	}
}

func TestLeaderCancellationDoesNotCancelActiveWaiter(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	lookup := aircraftLookupFunc(func(
		ctx context.Context,
		icao24 string,
	) (aircraft.Aircraft, error) {
		if calls.Add(1) == 1 {
			close(started)
		}
		select {
		case <-release:
			return aircraft.Aircraft{
				ICAO24:       icao24,
				Registration: "4K-AZ01",
			}, nil
		case <-ctx.Done():
			return aircraft.Aircraft{}, ctx.Err()
		}
	})
	provider := newTestProvider(t, Config{Lookup: lookup})
	reference := extractor.AircraftReference{ICAO24: "ABC123"}

	leaderContext, cancelLeader := context.WithCancel(context.Background())
	leaderDone := make(chan error, 1)
	go func() {
		_, err := provider.Provide(leaderContext, reference)
		leaderDone <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("shared lookup did not start")
	}

	waiterDone := make(chan struct {
		features flightfeatures.AircraftFeatures
		err      error
	}, 1)
	go func() {
		features, err := provider.Provide(context.Background(), reference)
		waiterDone <- struct {
			features flightfeatures.AircraftFeatures
			err      error
		}{features: features, err: err}
	}()

	cancelLeader()
	if err := <-leaderDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("leader error = %v, want context.Canceled", err)
	}
	close(release)

	result := <-waiterDone
	if result.err != nil {
		t.Fatalf("waiter error = %v", result.err)
	}
	if result.features.Registration != "4K-AZ01" {
		t.Fatalf("waiter features = %#v", result.features)
	}
	if calls.Load() != 1 {
		t.Fatalf("lookup calls = %d, want 1", calls.Load())
	}
}

func TestAcquireAtomicallyUsesCompletedCache(t *testing.T) {
	provider := newTestProvider(t, Config{
		Lookup: aircraftLookupFunc(func(
			context.Context,
			string,
		) (aircraft.Aircraft, error) {
			return aircraft.Aircraft{}, errors.New("unexpected lookup")
		}),
	})

	_, call, leader, cached := provider.acquire("ABC123")
	if !leader || cached || call == nil {
		t.Fatalf("first acquire = leader:%v cached:%v call:%p", leader, cached, call)
	}
	provider.completeCall(
		"ABC123",
		call,
		flightfeatures.AircraftFeatures{Registration: "4K-AZ01"},
		time.Minute,
		nil,
	)

	features, secondCall, secondLeader, secondCached := provider.acquire("ABC123")
	if !secondCached || secondLeader || secondCall != nil {
		t.Fatalf(
			"second acquire = leader:%v cached:%v call:%p",
			secondLeader,
			secondCached,
			secondCall,
		)
	}
	if features.Registration != "4K-AZ01" {
		t.Fatalf("cached features = %#v", features)
	}
}

func TestProviderBoundsCacheAndEvictsOldestExpiry(t *testing.T) {
	fixedNow := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
	var calls atomic.Int32
	provider := newTestProvider(t, Config{
		Lookup: aircraftLookupFunc(func(
			_ context.Context,
			icao24 string,
		) (aircraft.Aircraft, error) {
			calls.Add(1)
			return aircraft.Aircraft{ICAO24: icao24, Registration: "REG-" + icao24}, nil
		}),
		Now: func() time.Time { return fixedNow },
	})
	provider.maxCacheEntries = 2

	for _, icao24 := range []string{"AAA111", "BBB222", "CCC333"} {
		if _, err := provider.Provide(
			context.Background(),
			extractor.AircraftReference{ICAO24: icao24},
		); err != nil {
			t.Fatalf("Provide(%s) error = %v", icao24, err)
		}
	}
	provider.mutex.Lock()
	cacheSize := len(provider.cache)
	_, oldestStillPresent := provider.cache["AAA111"]
	provider.mutex.Unlock()
	if cacheSize != 2 || oldestStillPresent {
		t.Fatalf("cache size = %d, oldest present = %v", cacheSize, oldestStillPresent)
	}

	if _, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{ICAO24: "AAA111"},
	); err != nil {
		t.Fatalf("reloaded Provide() error = %v", err)
	}
	if calls.Load() != 4 {
		t.Fatalf("lookup calls = %d, want 4", calls.Load())
	}
}

func TestProviderPrunesExpiredUniqueEntries(t *testing.T) {
	currentTime := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
	provider := newTestProvider(t, Config{
		Lookup: aircraftLookupFunc(func(
			_ context.Context,
			icao24 string,
		) (aircraft.Aircraft, error) {
			return aircraft.Aircraft{ICAO24: icao24, Registration: "REG-" + icao24}, nil
		}),
		PositiveCacheTTL: time.Minute,
		Now:              func() time.Time { return currentTime },
	})
	provider.maxCacheEntries = 10

	for _, icao24 := range []string{"AAA111", "BBB222", "CCC333"} {
		if _, err := provider.Provide(
			context.Background(),
			extractor.AircraftReference{ICAO24: icao24},
		); err != nil {
			t.Fatalf("Provide(%s) error = %v", icao24, err)
		}
	}
	currentTime = currentTime.Add(2 * time.Minute)
	if _, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{ICAO24: "DDD444"},
	); err != nil {
		t.Fatalf("Provide(DDD444) error = %v", err)
	}

	provider.mutex.Lock()
	cacheSize := len(provider.cache)
	_, latestPresent := provider.cache["DDD444"]
	provider.mutex.Unlock()
	if cacheSize != 1 || !latestPresent {
		t.Fatalf("cache size = %d, latest present = %v", cacheSize, latestPresent)
	}
}

func TestProviderRejectsLookupWithoutAircraftIdentity(t *testing.T) {
	provider := newTestProvider(t, Config{
		Lookup: aircraftLookupFunc(func(
			context.Context,
			string,
		) (aircraft.Aircraft, error) {
			return aircraft.Aircraft{Registration: "4K-AZ01"}, nil
		}),
	})

	_, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{ICAO24: "ABC123"},
	)
	if !errors.Is(err, ErrAircraftIdentityMissing) {
		t.Fatalf("Provide() error = %v, want %v", err, ErrAircraftIdentityMissing)
	}
}

func TestProviderRejectsNilContext(t *testing.T) {
	provider := newTestProvider(t, Config{
		Lookup: aircraftLookupFunc(func(
			context.Context,
			string,
		) (aircraft.Aircraft, error) {
			return aircraft.Aircraft{}, nil
		}),
	})
	_, err := provider.Provide(nil, extractor.AircraftReference{ICAO24: "ABC123"})
	if !errors.Is(err, ErrContextRequired) {
		t.Fatalf("Provide() error = %v, want %v", err, ErrContextRequired)
	}
}

func TestSharedLookupHasBoundedLifetime(t *testing.T) {
	provider := newTestProvider(t, Config{
		Lookup: aircraftLookupFunc(func(
			ctx context.Context,
			_ string,
		) (aircraft.Aircraft, error) {
			<-ctx.Done()
			return aircraft.Aircraft{}, ctx.Err()
		}),
	})
	provider.lookupTimeout = 20 * time.Millisecond

	_, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{ICAO24: "ABC123"},
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Provide() error = %v, want context.DeadlineExceeded", err)
	}
}
