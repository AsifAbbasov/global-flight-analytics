package aircraftprovider

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/jackc/pgx/v5"
)

var icao24Pattern = regexp.MustCompile(`^[A-F0-9]{6}$`)

var _ extractor.AircraftFeatureProvider = (*Provider)(nil)

type Provider struct {
	lookup           AircraftLookup
	positiveCacheTTL time.Duration
	negativeCacheTTL time.Duration
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

	positiveCacheTTL := config.PositiveCacheTTL
	if positiveCacheTTL == 0 {
		positiveCacheTTL = DefaultPositiveCacheTTL
	}
	if positiveCacheTTL < 0 {
		return nil, ErrInvalidPositiveCacheTTL
	}

	negativeCacheTTL := config.NegativeCacheTTL
	if negativeCacheTTL == 0 {
		negativeCacheTTL = DefaultNegativeCacheTTL
	}
	if negativeCacheTTL < 0 {
		return nil, ErrInvalidNegativeCacheTTL
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	isNotFound := config.IsNotFound
	if isNotFound == nil {
		isNotFound = func(err error) bool {
			return errors.Is(err, pgx.ErrNoRows)
		}
	}

	return &Provider{
		lookup:           config.Lookup,
		positiveCacheTTL: positiveCacheTTL,
		negativeCacheTTL: negativeCacheTTL,
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
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.AircraftFeatures{}, err
	}

	icao24, err := normalizeICAO24(reference.ICAO24)
	if err != nil {
		return flightfeatures.AircraftFeatures{}, err
	}

	if cached, found := provider.cached(icao24); found {
		return cloneFeatures(cached), nil
	}

	call, leader := provider.beginCall(icao24)
	if !leader {
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

	features, lookupErr := provider.resolve(ctx, icao24)
	provider.completeCall(icao24, call, features, lookupErr)

	if lookupErr != nil {
		return flightfeatures.AircraftFeatures{}, lookupErr
	}

	return cloneFeatures(features), nil
}

func (provider *Provider) resolve(
	ctx context.Context,
	icao24 string,
) (flightfeatures.AircraftFeatures, error) {
	item, err := provider.lookup.GetByICAO24(ctx, icao24)
	if err != nil {
		if provider.isNotFound(err) {
			features := unavailableFeatures(
				"aircraft_metadata_not_found",
				"Aircraft metadata was not found for the supplied ICAO24.",
			)
			provider.storeCache(
				icao24,
				features,
				provider.negativeCacheTTL,
			)

			return features, nil
		}

		return flightfeatures.AircraftFeatures{},
			&LookupError{
				ICAO24: icao24,
				Err:    err,
			}
	}
	if err := ctx.Err(); err != nil {
		return flightfeatures.AircraftFeatures{}, err
	}

	returnedICAO24 := strings.ToUpper(
		strings.TrimSpace(item.ICAO24),
	)
	if returnedICAO24 != "" && returnedICAO24 != icao24 {
		return flightfeatures.AircraftFeatures{},
			ErrAircraftIdentityMismatch
	}

	features := mapAircraft(item)
	provider.storeCache(
		icao24,
		features,
		provider.positiveCacheTTL,
	)

	return features, nil
}

func (provider *Provider) cached(
	icao24 string,
) (flightfeatures.AircraftFeatures, bool) {
	now := provider.now().UTC()

	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	entry, exists := provider.cache[icao24]
	if !exists {
		return flightfeatures.AircraftFeatures{}, false
	}
	if !now.Before(entry.expiresAt) {
		delete(provider.cache, icao24)
		return flightfeatures.AircraftFeatures{}, false
	}

	return cloneFeatures(entry.features), true
}

func (provider *Provider) storeCache(
	icao24 string,
	features flightfeatures.AircraftFeatures,
	ttl time.Duration,
) {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	provider.cache[icao24] = cacheEntry{
		features:  cloneFeatures(features),
		expiresAt: provider.now().UTC().Add(ttl),
	}
}

func (provider *Provider) beginCall(
	icao24 string,
) (*inFlightCall, bool) {
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	if existing, exists := provider.inFlight[icao24]; exists {
		return existing, false
	}

	call := &inFlightCall{
		done: make(chan struct{}),
	}
	provider.inFlight[icao24] = call

	return call, true
}

func (provider *Provider) completeCall(
	icao24 string,
	call *inFlightCall,
	features flightfeatures.AircraftFeatures,
	err error,
) {
	provider.mutex.Lock()
	call.features = cloneFeatures(features)
	call.err = err
	delete(provider.inFlight, icao24)
	close(call.done)
	provider.mutex.Unlock()
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
	}

	availableFieldCount := countAvailableFields(features)
	features.Evidence = flightfeatures.GroupEvidence{
		AvailableFieldCount: availableFieldCount,
		TotalFieldCount:     AircraftFeatureFieldCount,
	}

	switch {
	case availableFieldCount == AircraftFeatureFieldCount:
		features.Evidence.Status =
			flightfeatures.AvailabilityStatusAvailable
	case availableFieldCount == 0:
		features.Evidence.Status =
			flightfeatures.AvailabilityStatusUnavailable
		features.Evidence.Limitations =
			[]flightfeatures.FeatureLimitation{
				{
					Code:    "aircraft_metadata_empty",
					Message: "Aircraft lookup succeeded but returned no usable metadata fields.",
				},
			}
	default:
		features.Evidence.Status =
			flightfeatures.AvailabilityStatusPartial
		features.Evidence.Limitations =
			[]flightfeatures.FeatureLimitation{
				{
					Code:    "aircraft_metadata_partial",
					Message: "Only part of the aircraft metadata is available.",
				},
			}
	}

	return features
}

func unavailableFeatures(
	code string,
	message string,
) flightfeatures.AircraftFeatures {
	return flightfeatures.AircraftFeatures{
		Evidence: flightfeatures.GroupEvidence{
			Status:          flightfeatures.AvailabilityStatusUnavailable,
			TotalFieldCount: AircraftFeatureFieldCount,
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

func normalizeICAO24(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if !icao24Pattern.MatchString(normalized) {
		return "", ErrInvalidICAO24
	}

	return normalized, nil
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
