package sourceconstraints

type DecisionLevel string

const (
	DecisionLevelAllowed DecisionLevel = "allowed"
	DecisionLevelLimited DecisionLevel = "limited"
	DecisionLevelBlocked DecisionLevel = "blocked"
)

type ClaimStrength string

const (
	ClaimStrengthObserved  ClaimStrength = "observed"
	ClaimStrengthDerived   ClaimStrength = "derived"
	ClaimStrengthEstimated ClaimStrength = "estimated"
	ClaimStrengthUnknown   ClaimStrength = "unknown"
	ClaimStrengthBlocked   ClaimStrength = "blocked"
)

type SourceClass string

const (
	SourceClassPublicCommunitySurveillance SourceClass = "public_community_surveillance"
	SourceClassPublicStaticDataset         SourceClass = "public_static_dataset"
	SourceClassPublicWeatherService        SourceClass = "public_weather_service"
	SourceClassFirstPartyReceiverNetwork   SourceClass = "first_party_receiver_network"
	SourceClassSatelliteSurveillance       SourceClass = "satellite_surveillance"
	SourceClassCommercialAviation          SourceClass = "commercial_aviation"
	SourceClassOfficialOperational         SourceClass = "official_operational"
)

type Capability string

const (
	CapabilityRegionalLiveObservation     Capability = "regional_live_observation"
	CapabilityHistoricalFlightObservation Capability = "historical_flight_observation"
	CapabilityEstimatedAirportContext     Capability = "estimated_airport_context"
	CapabilityExperimentalTrackContext    Capability = "experimental_track_context"
	CapabilityGlobalContinuousTracking    Capability = "global_continuous_tracking"
	CapabilityOceanicContinuousTracking   Capability = "oceanic_continuous_tracking"
	CapabilityOwnReceiverObservation      Capability = "own_receiver_observation"
	CapabilityOfficialSchedule            Capability = "official_schedule"
	CapabilityOfficialDelayCause          Capability = "official_delay_cause"
	CapabilityPilotIntent                 Capability = "pilot_intent"
	CapabilityATCInstruction              Capability = "atc_instruction"
	CapabilityCertifiedSeparation         Capability = "certified_separation"
	CapabilityOperationalWeather          Capability = "operational_weather"
	CapabilityCommercialFleetData         Capability = "commercial_fleet_data"
)

type ProjectConstraints struct {
	FreeSourcesOnly                 bool `json:"free_sources_only"`
	HasOwnCollectionInfrastructure  bool `json:"has_own_collection_infrastructure"`
	HasSatelliteAccess              bool `json:"has_satellite_access"`
	HasCommercialAviationDataAccess bool `json:"has_commercial_aviation_data_access"`
	ResearchOnly                    bool `json:"research_only"`
}

func FixedProjectConstraints() ProjectConstraints {
	return ProjectConstraints{
		FreeSourcesOnly:                 true,
		HasOwnCollectionInfrastructure:  false,
		HasSatelliteAccess:              false,
		HasCommercialAviationDataAccess: false,
		ResearchOnly:                    true,
	}
}

type SourceProfile struct {
	ID                                    string      `json:"id"`
	Class                                 SourceClass `json:"class"`
	FreeAccess                            bool        `json:"free_access"`
	ExternallyCollected                   bool        `json:"externally_collected"`
	RequiresOwnInfrastructure             bool        `json:"requires_own_infrastructure"`
	SatelliteDerived                      bool        `json:"satellite_derived"`
	Commercial                            bool        `json:"commercial"`
	OfficialOperational                   bool        `json:"official_operational"`
	SupportsRegionalObservation           bool        `json:"supports_regional_observation"`
	SupportsHistoricalFlights             bool        `json:"supports_historical_flights"`
	SupportsEstimatedAirports             bool        `json:"supports_estimated_airports"`
	SupportsExperimentalTracks            bool        `json:"supports_experimental_tracks"`
	ContinuousCoverage                    bool        `json:"continuous_coverage"`
	OceanicCoverage                       bool        `json:"oceanic_coverage"`
	AttributionRequired                   bool        `json:"attribution_required"`
	AttributionText                       string      `json:"attribution_text"`
	NonCommercialUseOnly                  bool        `json:"non_commercial_use_only"`
	CloudHostingAvailabilityNotGuaranteed bool        `json:"cloud_hosting_availability_not_guaranteed"`
}

type Request struct {
	Constraints ProjectConstraints `json:"constraints"`
	Source      SourceProfile      `json:"source"`
	Capability  Capability         `json:"capability"`
}

type Decision struct {
	ContractVersion      string        `json:"contract_version"`
	SourceID             string        `json:"source_id"`
	Capability           Capability    `json:"capability"`
	Level                DecisionLevel `json:"level"`
	MaximumClaimStrength ClaimStrength `json:"maximum_claim_strength"`
	Reasons              []string      `json:"reasons"`
	RequiredLabels       []string      `json:"required_labels"`
	ScopeGuards          []string      `json:"scope_guards"`
}

func (decision Decision) Usable() bool {
	return decision.Level == DecisionLevelAllowed ||
		decision.Level == DecisionLevelLimited
}
