package aircraftprovider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/jackc/pgx/v5"
)

type lookupStub struct {
	mutex   sync.Mutex
	item    aircraft.Aircraft
	err     error
	calls   int
	icao24s []string
	started chan struct{}
	release chan struct{}
}

func (stub *lookupStub) GetByICAO24(
	ctx context.Context,
	icao24 string,
) (aircraft.Aircraft, error) {
	stub.mutex.Lock()
	stub.calls++
	stub.icao24s = append(stub.icao24s, icao24)
	started := stub.started
	release := stub.release
	item := stub.item
	err := stub.err
	stub.mutex.Unlock()

	if started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
	}
	if release != nil {
		select {
		case <-ctx.Done():
			return aircraft.Aircraft{}, ctx.Err()
		case <-release:
		}
	}

	return item, err
}

func (stub *lookupStub) callCount() int {
	stub.mutex.Lock()
	defer stub.mutex.Unlock()

	return stub.calls
}

func (stub *lookupStub) requestedICAO24s() []string {
	stub.mutex.Lock()
	defer stub.mutex.Unlock()

	return append([]string(nil), stub.icao24s...)
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	validLookup := &lookupStub{}

	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name:    "lookup",
			config:  Config{},
			wantErr: ErrLookupRequired,
		},
		{
			name: "positive cache ttl",
			config: Config{
				Lookup:           validLookup,
				PositiveCacheTTL: -time.Second,
			},
			wantErr: ErrInvalidPositiveCacheTTL,
		},
		{
			name: "negative cache ttl",
			config: Config{
				Lookup:           validLookup,
				NegativeCacheTTL: -time.Second,
			},
			wantErr: ErrInvalidNegativeCacheTTL,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.config)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"New() error = %v, want %v",
					err,
					test.wantErr,
				)
			}
		})
	}
}

func TestProviderReturnsAvailableNormalizedFeatures(t *testing.T) {
	lookup := &lookupStub{
		item: aircraft.Aircraft{
			ICAO24:       "abc123",
			Registration: " 4K-AZ01 ",
			Manufacturer: " Example Manufacturer ",
			Model:        " Example Model ",
			AircraftType: " Example Type ",
			Airline:      " Example Airline ",
			Country:      " Azerbaijan ",
		},
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})

	features, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{
			ICAO24: " abc123 ",
		},
	)
	if err != nil {
		t.Fatalf("Provide() error = %v", err)
	}

	want := flightfeatures.AircraftFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:              flightfeatures.AvailabilityStatusAvailable,
			AvailableFieldCount: AircraftFeatureFieldCount,
			TotalFieldCount:     AircraftFeatureFieldCount,
		},
		Registration: "4K-AZ01",
		Manufacturer: "Example Manufacturer",
		Model:        "Example Model",
		AircraftType: "Example Type",
		Airline:      "Example Airline",
		Country:      "Azerbaijan",
	}
	if !reflect.DeepEqual(features, want) {
		t.Fatalf(
			"features = %#v, want %#v",
			features,
			want,
		)
	}
	if got := lookup.requestedICAO24s(); !reflect.DeepEqual(
		got,
		[]string{"ABC123"},
	) {
		t.Fatalf("requested ICAO24s = %#v", got)
	}
}

func TestProviderReturnsPartialEvidence(t *testing.T) {
	lookup := &lookupStub{
		item: aircraft.Aircraft{
			ICAO24:       "ABC123",
			Registration: "4K-AZ01",
			Model:        "Example Model",
			Country:      "Azerbaijan",
		},
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})

	features, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{
			ICAO24: "ABC123",
		},
	)
	if err != nil {
		t.Fatalf("Provide() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusPartial ||
		features.Evidence.AvailableFieldCount != 3 ||
		features.Evidence.TotalFieldCount !=
			AircraftFeatureFieldCount {
		t.Fatalf(
			"unexpected partial evidence: %#v",
			features.Evidence,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"aircraft_metadata_partial",
	) {
		t.Fatalf(
			"missing partial limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestProviderReturnsUnavailableEvidenceForEmptyRecord(
	t *testing.T,
) {
	lookup := &lookupStub{
		item: aircraft.Aircraft{
			ICAO24: "ABC123",
		},
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})

	features, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{
			ICAO24: "ABC123",
		},
	)
	if err != nil {
		t.Fatalf("Provide() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusUnavailable ||
		features.Evidence.AvailableFieldCount != 0 {
		t.Fatalf(
			"unexpected unavailable evidence: %#v",
			features.Evidence,
		)
	}
	if !hasLimitation(
		features.Evidence.Limitations,
		"aircraft_metadata_empty",
	) {
		t.Fatalf(
			"missing empty-record limitation: %#v",
			features.Evidence.Limitations,
		)
	}
}

func TestProviderTreatsNotFoundAsUnavailableEvidence(
	t *testing.T,
) {
	lookup := &lookupStub{
		err: fmt.Errorf(
			"read aircraft: %w",
			pgx.ErrNoRows,
		),
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})

	features, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{
			ICAO24: "ABC123",
		},
	)
	if err != nil {
		t.Fatalf("Provide() error = %v", err)
	}

	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusUnavailable ||
		!hasLimitation(
			features.Evidence.Limitations,
			"aircraft_metadata_not_found",
		) {
		t.Fatalf(
			"unexpected not-found features: %#v",
			features,
		)
	}
}

func TestProviderWrapsRepositoryFailures(t *testing.T) {
	sourceErr := errors.New("database unavailable")
	lookup := &lookupStub{
		err: sourceErr,
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})

	_, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{
			ICAO24: "ABC123",
		},
	)
	if !errors.Is(err, sourceErr) {
		t.Fatalf(
			"Provide() error = %v, want wrapped source error",
			err,
		)
	}

	var lookupErr *LookupError
	if !errors.As(err, &lookupErr) {
		t.Fatalf(
			"Provide() error = %T, want *LookupError",
			err,
		)
	}
	if lookupErr.ICAO24 != "ABC123" {
		t.Fatalf(
			"LookupError ICAO24 = %q",
			lookupErr.ICAO24,
		)
	}

	lookup.err = nil
	lookup.item = aircraft.Aircraft{
		ICAO24:       "ABC123",
		Registration: "4K-AZ01",
	}
	if _, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{
			ICAO24: "ABC123",
		},
	); err != nil {
		t.Fatalf("second Provide() error = %v", err)
	}
	if lookup.callCount() != 2 {
		t.Fatalf(
			"lookup calls = %d, want 2 because errors are not cached",
			lookup.callCount(),
		)
	}
}

func TestProviderRejectsInvalidReferenceBeforeLookup(
	t *testing.T,
) {
	lookup := &lookupStub{}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})

	for _, value := range []string{
		"",
		"ABC",
		"ABC12Z",
		"ABC1234",
	} {
		_, err := provider.Provide(
			context.Background(),
			extractor.AircraftReference{
				ICAO24: value,
			},
		)
		if !errors.Is(err, ErrInvalidICAO24) {
			t.Fatalf(
				"Provide(%q) error = %v, want %v",
				value,
				err,
				ErrInvalidICAO24,
			)
		}
	}

	if lookup.callCount() != 0 {
		t.Fatalf(
			"lookup calls = %d, want 0",
			lookup.callCount(),
		)
	}
}

func TestProviderRejectsMismatchedReturnedICAO24(
	t *testing.T,
) {
	lookup := &lookupStub{
		item: aircraft.Aircraft{
			ICAO24: "DEF456",
		},
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})

	_, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{
			ICAO24: "ABC123",
		},
	)
	if !errors.Is(
		err,
		ErrAircraftIdentityMismatch,
	) {
		t.Fatalf(
			"Provide() error = %v, want %v",
			err,
			ErrAircraftIdentityMismatch,
		)
	}
}

func TestProviderUsesPositiveCacheUntilExpiry(t *testing.T) {
	currentTime := time.Date(
		2026,
		time.July,
		14,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	lookup := &lookupStub{
		item: aircraft.Aircraft{
			ICAO24:       "ABC123",
			Registration: "4K-AZ01",
		},
	}
	provider := newTestProvider(t, Config{
		Lookup:           lookup,
		PositiveCacheTTL: time.Minute,
		Now: func() time.Time {
			return currentTime
		},
	})
	reference := extractor.AircraftReference{
		ICAO24: "ABC123",
	}

	if _, err := provider.Provide(
		context.Background(),
		reference,
	); err != nil {
		t.Fatalf("first Provide() error = %v", err)
	}
	if _, err := provider.Provide(
		context.Background(),
		reference,
	); err != nil {
		t.Fatalf("second Provide() error = %v", err)
	}
	if lookup.callCount() != 1 {
		t.Fatalf(
			"lookup calls before expiry = %d, want 1",
			lookup.callCount(),
		)
	}

	currentTime = currentTime.Add(time.Minute)
	if _, err := provider.Provide(
		context.Background(),
		reference,
	); err != nil {
		t.Fatalf("expired Provide() error = %v", err)
	}
	if lookup.callCount() != 2 {
		t.Fatalf(
			"lookup calls after expiry = %d, want 2",
			lookup.callCount(),
		)
	}
}

func TestProviderUsesNegativeCacheUntilExpiry(t *testing.T) {
	currentTime := time.Date(
		2026,
		time.July,
		14,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	lookup := &lookupStub{
		err: pgx.ErrNoRows,
	}
	provider := newTestProvider(t, Config{
		Lookup:           lookup,
		NegativeCacheTTL: time.Minute,
		Now: func() time.Time {
			return currentTime
		},
	})
	reference := extractor.AircraftReference{
		ICAO24: "ABC123",
	}

	for call := 0; call < 2; call++ {
		if _, err := provider.Provide(
			context.Background(),
			reference,
		); err != nil {
			t.Fatalf(
				"Provide(%d) error = %v",
				call,
				err,
			)
		}
	}
	if lookup.callCount() != 1 {
		t.Fatalf(
			"lookup calls before expiry = %d, want 1",
			lookup.callCount(),
		)
	}

	currentTime = currentTime.Add(time.Minute)
	if _, err := provider.Provide(
		context.Background(),
		reference,
	); err != nil {
		t.Fatalf("expired Provide() error = %v", err)
	}
	if lookup.callCount() != 2 {
		t.Fatalf(
			"lookup calls after expiry = %d, want 2",
			lookup.callCount(),
		)
	}
}

func TestProviderCoalescesConcurrentRequests(t *testing.T) {
	lookup := &lookupStub{
		item: aircraft.Aircraft{
			ICAO24:       "ABC123",
			Registration: "4K-AZ01",
		},
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})

	const workerCount = 32
	results := make(chan flightfeatures.AircraftFeatures, workerCount)
	errorsChannel := make(chan error, workerCount)
	start := make(chan struct{})
	var waitGroup sync.WaitGroup

	for worker := 0; worker < workerCount; worker++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			features, err := provider.Provide(
				context.Background(),
				extractor.AircraftReference{
					ICAO24: "ABC123",
				},
			)
			if err != nil {
				errorsChannel <- err
				return
			}
			results <- features
		}()
	}

	close(start)
	select {
	case <-lookup.started:
	case <-time.After(time.Second):
		t.Fatal("lookup did not start")
	}

	time.Sleep(20 * time.Millisecond)
	close(lookup.release)
	waitGroup.Wait()
	close(results)
	close(errorsChannel)

	for err := range errorsChannel {
		t.Fatalf("concurrent Provide() error = %v", err)
	}
	if lookup.callCount() != 1 {
		t.Fatalf(
			"lookup calls = %d, want 1",
			lookup.callCount(),
		)
	}

	resultCount := 0
	for features := range results {
		resultCount++
		if features.Registration != "4K-AZ01" {
			t.Fatalf(
				"unexpected concurrent features: %#v",
				features,
			)
		}
	}
	if resultCount != workerCount {
		t.Fatalf(
			"result count = %d, want %d",
			resultCount,
			workerCount,
		)
	}
}

func TestWaitingCallerCanCancelWithoutCancelingLeader(
	t *testing.T,
) {
	lookup := &lookupStub{
		item: aircraft.Aircraft{
			ICAO24:       "ABC123",
			Registration: "4K-AZ01",
		},
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})
	reference := extractor.AircraftReference{
		ICAO24: "ABC123",
	}

	leaderDone := make(chan error, 1)
	go func() {
		_, err := provider.Provide(
			context.Background(),
			reference,
		)
		leaderDone <- err
	}()

	select {
	case <-lookup.started:
	case <-time.After(time.Second):
		t.Fatal("leader lookup did not start")
	}

	waiterContext, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := provider.Provide(
		waiterContext,
		reference,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"waiting Provide() error = %v, want context.Canceled",
			err,
		)
	}

	close(lookup.release)
	if err := <-leaderDone; err != nil {
		t.Fatalf("leader Provide() error = %v", err)
	}
	if lookup.callCount() != 1 {
		t.Fatalf(
			"lookup calls = %d, want 1",
			lookup.callCount(),
		)
	}
}

func TestProviderReturnsDefensiveCopies(t *testing.T) {
	lookup := &lookupStub{
		item: aircraft.Aircraft{
			ICAO24:       "ABC123",
			Registration: "4K-AZ01",
		},
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})
	reference := extractor.AircraftReference{
		ICAO24: "ABC123",
	}

	first, err := provider.Provide(
		context.Background(),
		reference,
	)
	if err != nil {
		t.Fatalf("first Provide() error = %v", err)
	}
	first.Evidence.Limitations[0].Code = "changed"

	second, err := provider.Provide(
		context.Background(),
		reference,
	)
	if err != nil {
		t.Fatalf("second Provide() error = %v", err)
	}
	if second.Evidence.Limitations[0].Code !=
		"aircraft_metadata_partial" {
		t.Fatalf(
			"cached limitation changed: %#v",
			second.Evidence.Limitations,
		)
	}
}

func TestProviderHonorsCustomNotFoundClassifier(
	t *testing.T,
) {
	customNotFound := errors.New("custom not found")
	lookup := &lookupStub{
		err: customNotFound,
	}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
		IsNotFound: func(err error) bool {
			return errors.Is(err, customNotFound)
		},
	})

	features, err := provider.Provide(
		context.Background(),
		extractor.AircraftReference{
			ICAO24: "ABC123",
		},
	)
	if err != nil {
		t.Fatalf("Provide() error = %v", err)
	}
	if features.Evidence.Status !=
		flightfeatures.AvailabilityStatusUnavailable {
		t.Fatalf(
			"status = %q",
			features.Evidence.Status,
		)
	}
}

func TestProviderPreservesAlreadyCanceledContext(
	t *testing.T,
) {
	lookup := &lookupStub{}
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.Provide(
		ctx,
		extractor.AircraftReference{
			ICAO24: "ABC123",
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Provide() error = %v, want context.Canceled",
			err,
		)
	}
	if lookup.callCount() != 0 {
		t.Fatalf(
			"lookup calls = %d, want 0",
			lookup.callCount(),
		)
	}
}

func TestProviderContractConstantsRemainStable(t *testing.T) {
	if Version != "aircraft-feature-provider-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if AircraftFeatureFieldCount != 6 {
		t.Fatalf(
			"AircraftFeatureFieldCount = %d",
			AircraftFeatureFieldCount,
		)
	}
	if DefaultPositiveCacheTTL != 30*time.Minute ||
		DefaultNegativeCacheTTL != 5*time.Minute {
		t.Fatalf(
			"unexpected cache defaults: positive=%v negative=%v",
			DefaultPositiveCacheTTL,
			DefaultNegativeCacheTTL,
		)
	}
}

func TestConcurrentDifferentICAO24RequestsRemainIndependent(
	t *testing.T,
) {
	var calls atomic.Int32
	lookup := aircraftLookupFunc(
		func(
			ctx context.Context,
			icao24 string,
		) (aircraft.Aircraft, error) {
			calls.Add(1)
			return aircraft.Aircraft{
				ICAO24:       icao24,
				Registration: "REG-" + icao24,
			}, nil
		},
	)
	provider := newTestProvider(t, Config{
		Lookup: lookup,
	})

	var waitGroup sync.WaitGroup
	for _, icao24 := range []string{"ABC123", "DEF456"} {
		icao24 := icao24
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			features, err := provider.Provide(
				context.Background(),
				extractor.AircraftReference{
					ICAO24: icao24,
				},
			)
			if err != nil {
				t.Errorf(
					"Provide(%s) error = %v",
					icao24,
					err,
				)
				return
			}
			if features.Registration !=
				"REG-"+icao24 {
				t.Errorf(
					"registration = %q",
					features.Registration,
				)
			}
		}()
	}
	waitGroup.Wait()

	if calls.Load() != 2 {
		t.Fatalf(
			"lookup calls = %d, want 2",
			calls.Load(),
		)
	}
}

type aircraftLookupFunc func(
	ctx context.Context,
	icao24 string,
) (aircraft.Aircraft, error)

func (function aircraftLookupFunc) GetByICAO24(
	ctx context.Context,
	icao24 string,
) (aircraft.Aircraft, error) {
	return function(ctx, icao24)
}

func newTestProvider(
	t *testing.T,
	config Config,
) *Provider {
	t.Helper()

	provider, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return provider
}

func hasLimitation(
	limitations []flightfeatures.FeatureLimitation,
	code string,
) bool {
	for _, limitation := range limitations {
		if limitation.Code == code {
			return true
		}
	}

	return false
}
