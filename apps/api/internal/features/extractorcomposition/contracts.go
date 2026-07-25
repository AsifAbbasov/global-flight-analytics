package extractorcomposition

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
)

const Version = "flight-feature-extractor-composition-v2"

type Component string

const (
	ComponentGeographicalBuilder Component = "geographical_builder"
	ComponentAircraftProvider    Component = "aircraft_provider"
	ComponentExtractor           Component = "extractor"
)

type Config struct {
	AircraftLookup aircraftprovider.AircraftLookup

	GeographicCellPrecision int

	AircraftPositiveCacheTTL      time.Duration
	AircraftNegativeCacheTTL      time.Duration
	IsAircraftNotFound            func(error) bool
	AircraftNotFoundPolicyVersion string

	Now func() time.Time
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
