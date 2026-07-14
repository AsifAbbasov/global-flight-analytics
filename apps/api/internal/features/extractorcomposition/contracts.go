package extractorcomposition

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/aircraftprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/extractor"
)

const Version = "flight-feature-extractor-composition-v1"

const (
	ComponentGeographicalBuilder = "geographical_builder"
	ComponentAircraftProvider    = "aircraft_provider"
	ComponentExtractor           = "extractor"
)

type Config struct {
	AircraftLookup aircraftprovider.AircraftLookup

	GeographicCellPrecision int

	AircraftPositiveCacheTTL time.Duration
	AircraftNegativeCacheTTL time.Duration
	IsAircraftNotFound       func(error) bool

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

type Composition struct {
	Extractor *extractor.Extractor
	Versions  Versions
}
