package extractorcomposition

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/geographicalbuilder"
)

const Version = "flight-feature-extractor-composition-v6"
const disabledAircraftNotFoundPolicyVersion = "aircraft-enrichment-disabled-v1"

type Component string

const (
	ComponentGeographicalBuilder Component = "geographical_builder"
	ComponentAircraftProvider    Component = "aircraft_provider"
	ComponentExtractor           Component = "extractor"
)

type FeatureExtractor interface {
	Extract(
		ctx context.Context,
		request extractor.Request,
	) (flightfeatures.FlightFeatures, error)
}

type Config struct {
	aircraftLookup         aircraftprovider.AircraftLookup
	aircraftEnrichmentMode flightfeatures.AircraftEnrichmentMode

	geographicCellPrecision int

	aircraftCacheMode             aircraftprovider.CacheMode
	aircraftPositiveCacheTTL      time.Duration
	aircraftNegativeCacheTTL      time.Duration
	isAircraftNotFound            func(error) bool
	aircraftNotFoundPolicyVersion string

	now func() time.Time
}

func DefaultConfig(
	lookup aircraftprovider.AircraftLookup,
) Config {
	return Config{
		aircraftLookup:                lookup,
		aircraftEnrichmentMode:        flightfeatures.AircraftEnrichmentModeEnabled,
		geographicCellPrecision:       geographicalbuilder.DefaultGeographicCellPrecision,
		aircraftCacheMode:             aircraftprovider.CacheModeEnabled,
		aircraftPositiveCacheTTL:      aircraftprovider.DefaultPositiveCacheTTL,
		aircraftNegativeCacheTTL:      aircraftprovider.DefaultNegativeCacheTTL,
		aircraftNotFoundPolicyVersion: aircraftprovider.DefaultNotFoundPolicyVersion,
	}
}

func DefaultConfigWithoutAircraftEnrichment() Config {
	return Config{
		aircraftEnrichmentMode:        flightfeatures.AircraftEnrichmentModeDisabled,
		geographicCellPrecision:       geographicalbuilder.DefaultGeographicCellPrecision,
		aircraftCacheMode:             aircraftprovider.CacheModeDisabled,
		aircraftNotFoundPolicyVersion: disabledAircraftNotFoundPolicyVersion,
	}
}

func (config Config) WithGeographicCellPrecision(
	precision int,
) Config {
	config.geographicCellPrecision = precision
	return config
}

func (config Config) WithAircraftCacheDurations(
	positive time.Duration,
	negative time.Duration,
) Config {
	config.aircraftCacheMode = aircraftprovider.CacheModeEnabled
	config.aircraftPositiveCacheTTL = positive
	config.aircraftNegativeCacheTTL = negative
	return config
}

func (config Config) WithoutAircraftCache() Config {
	config.aircraftCacheMode = aircraftprovider.CacheModeDisabled
	config.aircraftPositiveCacheTTL = 0
	config.aircraftNegativeCacheTTL = 0
	return config
}

func (config Config) WithAircraftNotFoundPolicy(
	version string,
	classifier func(error) bool,
) Config {
	config.aircraftNotFoundPolicyVersion = version
	config.isAircraftNotFound = classifier
	return config
}

func (config Config) WithClock(
	now func() time.Time,
) Config {
	config.now = now
	return config
}

type Versions = flightfeatures.ProcessingComponentVersions
type ProcessingIdentity = flightfeatures.ProcessingIdentity

type Composition struct {
	Extractor           FeatureExtractor
	Versions            Versions
	ProcessingIdentity  ProcessingIdentity
	FingerprintIdentity string
}
