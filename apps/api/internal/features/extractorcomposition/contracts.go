package extractorcomposition

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/geographicalbuilder"
)

const Version = "flight-feature-extractor-composition-v3"

type Component string

const (
	ComponentGeographicalBuilder Component = "geographical_builder"
	ComponentAircraftProvider    Component = "aircraft_provider"
	ComponentExtractor           Component = "extractor"
)

type Config struct {
	aircraftLookup aircraftprovider.AircraftLookup

	geographicCellPrecision int

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
		geographicCellPrecision:       geographicalbuilder.DefaultGeographicCellPrecision,
		aircraftPositiveCacheTTL:      aircraftprovider.DefaultPositiveCacheTTL,
		aircraftNegativeCacheTTL:      aircraftprovider.DefaultNegativeCacheTTL,
		aircraftNotFoundPolicyVersion: aircraftprovider.DefaultNotFoundPolicyVersion,
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
	config.aircraftPositiveCacheTTL = positive
	config.aircraftNegativeCacheTTL = negative
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

type Versions struct {
	Composition         string
	Extractor           string
	AircraftProvider    string
	TemporalBuilder     string
	GeographicalBuilder string
	OperationalBuilder  string
	TrajectoryBuilder   string
}

type ProcessingIdentity struct {
	Versions                      Versions
	GeographicCellPrecision       int
	AircraftPositiveCacheTTL      time.Duration
	AircraftNegativeCacheTTL      time.Duration
	AircraftNotFoundPolicyVersion string
}

type Composition struct {
	Extractor           *extractor.Extractor
	Versions            Versions
	ProcessingIdentity  ProcessingIdentity
	FingerprintIdentity string
}
