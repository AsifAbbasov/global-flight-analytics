package researchdataset

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/sourceconstraints"
)

type ID string

const (
	IDEmergencyReference ID = "opensky-emergency-reference"
	IDClimbingAircraft   ID = "opensky-climbing-aircraft-2017"
	IDTrinoSnapshot2026  ID = "opensky-trino-snapshot-2026-03-01"
	IDWeeklyStateVectors ID = "opensky-weekly-state-vectors-2017-2022"
	IDPRCTakeoffWeight   ID = "opensky-prc-takeoff-weight-2024"

	IDRawPhysicalLayer ID = "opensky-raw-physical-layer"
	IDLocaRDS          ID = "opensky-locards"
	IDCOVID19          ID = "opensky-covid-2019-2022"
	IDAircraftMetadata ID = "opensky-aircraft-metadata"
	IDGICB             ID = "opensky-gicb-capabilities"
	IDADSC             ID = "opensky-adsc"
)

type Selection string

const (
	SelectionAdopted  Selection = "adopted_for_bounded_offline_research"
	SelectionDeferred Selection = "deferred"
	SelectionBlocked  Selection = "blocked"
)

type Profile struct {
	ID        ID
	Name      string
	Selection Selection

	Source sourceconstraints.SourceProfile

	Purposes        []string
	RequiredLabels  []string
	BlockedPurposes []string

	AllowedTables []string
	BlockedTables []string

	MaximumDownloadBytes        int64
	MaximumRecords              int64
	RequiresRegionFilter        bool
	RequiresLicenseReview       bool
	ProductionDependencyAllowed bool
}

type File struct {
	Name      string
	Format    string
	SizeBytes int64
	SHA256    string
}

type Manifest struct {
	DatasetID ID
	Version   string

	Files          []File
	SelectedTables []string

	TotalBytes     int64
	MaximumRecords int64
	RegionFilter   string

	OfflineOnly          bool
	ProductionDependency bool
	LicenseReviewed      bool
	AttributionProvided  bool

	PreparedAt time.Time
}

type Decision struct {
	DatasetID ID
	Allowed   bool
	Reasons   []string
	Labels    []string
}
