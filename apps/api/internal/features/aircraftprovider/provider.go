package aircraftprovider

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

var _ extractor.AircraftFeatureProvider = (*Provider)(nil)

type Provider struct {
	lookup           AircraftLookup
	cacheMode        CacheMode
	positiveCacheTTL time.Duration
	negativeCacheTTL time.Duration
	maxCacheEntries  int
	lookupTimeout    time.Duration
	now              func() time.Time
	isNotFound       func(error) bool

	mutex    sync.Mutex
	cache    map[string]cacheEntry
	inFlight map[string]*inFlightCall
}

type cacheEntry struct {
	features  flightfeatures.AircraftFeatures
	expiresAt time.Time
}

type inFlightCall struct {
	done     chan struct{}
	features flightfeatures.AircraftFeatures
	err      error
}

func New(config Config) (*Provider, error) {
	if config.Lookup == nil {
		return nil, ErrLookupRequired
	}

	cacheMode := config.CacheMode
	if cacheMode == CacheModeDefault {
		cacheMode = CacheModeEnabled
	}
	if cacheMode != CacheModeEnabled && cacheMode != CacheModeDisabled {
		return nil, ErrInvalidCacheMode
	}

	positiveCacheTTL := config.PositiveCacheTTL
	negativeCacheTTL := config.NegativeCacheTTL
	if cacheMode == CacheModeEnabled {
		if positiveCacheTTL == 0 {
			positiveCacheTTL = DefaultPositiveCacheTTL
		}
		if positiveCacheTTL < 0 {
			return nil, ErrInvalidPositiveCacheTTL
		}
		if negativeCacheTTL == 0 {
			negativeCacheTTL = DefaultNegativeCacheTTL
		}
		if negativeCacheTTL < 0 {
			return nil, ErrInvalidNegativeCacheTTL
		}
	} else {
		positiveCacheTTL = 0
		negativeCacheTTL = 0
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	isNotFound := config.IsNotFound
	if isNotFound == nil {
		isNotFound = func(err error) bool {
			return errors.Is(err, aircraft.ErrNotFound)
		}
	}

	return &Provider{
		lookup:           config.Lookup,
		cacheMode:        cacheMode,
		positiveCacheTTL: positiveCacheTTL,
		negativeCacheTTL: negativeCacheTTL,
		maxCacheEntries:  DefaultMaxCacheEntries,
		lookupTimeout:    DefaultLookupTimeout,
		now:              now,
		isNotFound:       isNotFound,
		cache:            make(map[string]cacheEntry),
		inFlight:         make(map[string]*inFlightCall),
	}, nil
}

func (provider *Provider) Provide(
	ctx context.Context,
	reference extractor.AircraftReference,
) (flightfeatures.AircraftFeatures, error) {
	if ctx == nil {
		return flightfeatures.AircraftFeatures{}, ErrContextRequired
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.AircraftFeatures{}, err
	}

	icao24, valid := aircraft.NormalizeICAO24(reference.ICAO24)
	if !valid {
		return flightfeatures.AircraftFeatures{}, ErrInvalidICAO24
	}

	cached, call, leader, found := provider.acquire(icao24)
	if found {
		return applyTemporalPolicy(cached, reference.AsOfTime), nil
	}
	if leader {
		provider.startLookup(context.WithoutCancel(ctx), icao24, call)
	}

	features, err := waitForCall(ctx, call)
	if err != nil {
		return flightfeatures.AircraftFeatures{}, err
	}
	return applyTemporalPolicy(features, reference.AsOfTime), nil
}

func (provider *Provider) acquire(
	icao24 string,
) (
	flightfeatures.AircraftFeatures,
	*inFlightCall,
	bool,
	bool,
) {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	if provider.cacheMode == CacheModeEnabled {
		now := provider.now().UTC()
		provider.pruneExpiredLocked(now)
		if entry, exists := provider.cache[icao24]; exists {
			return cloneFeatures(entry.features), nil, false, true
		}
	}

	if existing, exists := provider.inFlight[icao24]; exists {
		return flightfeatures.AircraftFeatures{}, existing, false, false
	}

	call := &inFlightCall{done: make(chan struct{})}
	provider.inFlight[icao24] = call
	return flightfeatures.AircraftFeatures{}, call, true, false
}

func (provider *Provider) startLookup(
	base context.Context,
	icao24 string,
	call *inFlightCall,
) {
	go func() {
		lookupContext, cancel := context.WithTimeout(
			base,
			provider.lookupTimeout,
		)
		defer cancel()

		features, ttl, err := provider.resolve(lookupContext, icao24)
		provider.completeCall(icao24, call, features, ttl, err)
	}()
}

func waitForCall(
	ctx context.Context,
	call *inFlightCall,
) (flightfeatures.AircraftFeatures, error) {
	select {
	case <-call.done:
		if call.err != nil {
			return flightfeatures.AircraftFeatures{}, call.err
		}
		return cloneFeatures(call.features), nil
	default:
	}

	select {
	case <-ctx.Done():
		return flightfeatures.AircraftFeatures{}, ctx.Err()
	case <-call.done:
		if call.err != nil {
			return flightfeatures.AircraftFeatures{}, call.err
		}
		return cloneFeatures(call.features), nil
	}
}

func (provider *Provider) resolve(
	ctx context.Context,
	icao24 string,
) (flightfeatures.AircraftFeatures, time.Duration, error) {
	item, err := provider.lookup.GetByICAO24(ctx, icao24)
	if err != nil {
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return flightfeatures.AircraftFeatures{}, 0, err
		}
		if provider.isNotFound(err) {
			features := unavailableFeatures(
				"aircraft_metadata_not_found",
				"Aircraft metadata was not found for the supplied ICAO24.",
			)
			return features, provider.negativeCacheTTL, nil
		}
		return flightfeatures.AircraftFeatures{}, 0, &LookupError{
			ICAO24: icao24,
			Err:    err,
		}
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.AircraftFeatures{}, 0, err
	}

	if strings.TrimSpace(item.ICAO24) == "" {
		return flightfeatures.AircraftFeatures{}, 0, ErrAircraftIdentityMissing
	}
	returnedICAO24, valid := aircraft.NormalizeICAO24(item.ICAO24)
	if !valid || returnedICAO24 != icao24 {
		return flightfeatures.AircraftFeatures{}, 0, ErrAircraftIdentityMismatch
	}

	return mapAircraft(item), provider.positiveCacheTTL, nil
}

func (provider *Provider) completeCall(
	icao24 string,
	call *inFlightCall,
	features flightfeatures.AircraftFeatures,
	ttl time.Duration,
	err error,
) {
	provider.mutex.Lock()
	if err == nil && provider.cacheMode == CacheModeEnabled && ttl > 0 {
		provider.storeCacheLocked(
			icao24,
			features,
			provider.now().UTC().Add(ttl),
		)
	}
	call.features = cloneFeatures(features)
	call.err = err
	delete(provider.inFlight, icao24)
	close(call.done)
	provider.mutex.Unlock()
}

func (provider *Provider) storeCacheLocked(
	icao24 string,
	features flightfeatures.AircraftFeatures,
	expiresAt time.Time,
) {
	provider.pruneExpiredLocked(provider.now().UTC())
	if _, exists := provider.cache[icao24]; !exists &&
		len(provider.cache) >= provider.maxCacheEntries {
		provider.evictOneLocked()
	}
	provider.cache[icao24] = cacheEntry{
		features:  cloneFeatures(features),
		expiresAt: expiresAt,
	}
}

func (provider *Provider) pruneExpiredLocked(now time.Time) {
	for key, entry := range provider.cache {
		if !now.Before(entry.expiresAt) {
			delete(provider.cache, key)
		}
	}
}

func (provider *Provider) evictOneLocked() {
	victimKey := ""
	var victimExpiry time.Time
	for key, entry := range provider.cache {
		if victimKey == "" ||
			entry.expiresAt.Before(victimExpiry) ||
			(entry.expiresAt.Equal(victimExpiry) && key < victimKey) {
			victimKey = key
			victimExpiry = entry.expiresAt
		}
	}
	if victimKey != "" {
		delete(provider.cache, victimKey)
	}
}

func mapAircraft(
	item aircraft.Aircraft,
) flightfeatures.AircraftFeatures {
	features := flightfeatures.AircraftFeatures{
		Registration: strings.TrimSpace(item.Registration),
		Manufacturer: strings.TrimSpace(item.Manufacturer),
		Model:        strings.TrimSpace(item.Model),
		AircraftType: strings.TrimSpace(item.AircraftType),
		Airline:      strings.TrimSpace(item.Airline),
		Country:      strings.TrimSpace(item.Country),
		MetadataUpdatedAt: normalizedMetadataUpdatedAt(
			item.MetadataUpdatedAt,
		),
	}

	availableFieldCount := countAvailableFields(features)
	features.Evidence = flightfeatures.GroupEvidence{
		AvailableFieldCount: availableFieldCount,
		TotalFieldCount:     flightfeatures.CurrentGroupFieldCount(flightfeatures.FeatureGroupAircraft),
	}

	switch {
	case availableFieldCount == flightfeatures.CurrentGroupFieldCount(flightfeatures.FeatureGroupAircraft):
		features.Evidence.Status = flightfeatures.AvailabilityStatusAvailable
	case availableFieldCount == 0:
		features.Evidence.Status = flightfeatures.AvailabilityStatusUnavailable
		features.Evidence.Limitations = []flightfeatures.FeatureLimitation{
			{
				Code:    "aircraft_metadata_empty",
				Message: "Aircraft lookup succeeded but returned no usable metadata fields.",
			},
		}
	default:
		features.Evidence.Status = flightfeatures.AvailabilityStatusPartial
		features.Evidence.Limitations = []flightfeatures.FeatureLimitation{
			{
				Code:    "aircraft_metadata_partial",
				Message: "Only part of the aircraft metadata is available.",
			},
		}
	}

	return features
}

func applyTemporalPolicy(
	features flightfeatures.AircraftFeatures,
	asOfTime time.Time,
) flightfeatures.AircraftFeatures {
	normalized := cloneFeatures(features)
	if asOfTime.IsZero() ||
		normalized.MetadataUpdatedAt.IsZero() ||
		!normalized.MetadataUpdatedAt.After(asOfTime.UTC()) {
		return normalized
	}

	blocked := unavailableFeatures(
		"aircraft_metadata_newer_than_feature_as_of",
		"Current aircraft metadata was updated after the feature as-of time and cannot be used as historical evidence.",
	)
	blocked.MetadataUpdatedAt = normalized.MetadataUpdatedAt
	return blocked
}

func normalizedMetadataUpdatedAt(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value.UTC()
}

func unavailableFeatures(
	code string,
	message string,
) flightfeatures.AircraftFeatures {
	return flightfeatures.AircraftFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:          flightfeatures.AvailabilityStatusUnavailable,
			TotalFieldCount: flightfeatures.CurrentGroupFieldCount(flightfeatures.FeatureGroupAircraft),
			Limitations: []flightfeatures.FeatureLimitation{
				{
					Code:    code,
					Message: message,
				},
			},
		},
	}
}

func countAvailableFields(
	features flightfeatures.AircraftFeatures,
) int {
	values := []string{
		features.Registration,
		features.Manufacturer,
		features.Model,
		features.AircraftType,
		features.Airline,
		features.Country,
	}

	count := 0
	for _, value := range values {
		if value != "" {
			count++
		}
	}
	return count
}

func cloneFeatures(
	features flightfeatures.AircraftFeatures,
) flightfeatures.AircraftFeatures {
	cloned := features
	cloned.Evidence.Limitations = append(
		[]flightfeatures.FeatureLimitation(nil),
		features.Evidence.Limitations...,
	)
	return cloned
}
