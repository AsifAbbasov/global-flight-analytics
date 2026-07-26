package flightfeatures

import "time"

type AircraftEnrichmentMode string

const (
	AircraftEnrichmentModeEnabled  AircraftEnrichmentMode = "enabled"
	AircraftEnrichmentModeDisabled AircraftEnrichmentMode = "disabled"
)

type ProcessingComponentVersions struct {
	Composition         string
	Extractor           string
	AircraftProvider    string
	TemporalBuilder     string
	GeographicalBuilder string
	OperationalBuilder  string
	TrajectoryBuilder   string
}

type ProcessingIdentity struct {
	Versions                      ProcessingComponentVersions
	GeographicCellPrecision       int
	AircraftEnrichmentMode        AircraftEnrichmentMode
	AircraftCacheMode             string
	AircraftPositiveCacheTTL      time.Duration
	AircraftNegativeCacheTTL      time.Duration
	AircraftNotFoundPolicyVersion string
	AircraftMetadataSourceName    string
}
