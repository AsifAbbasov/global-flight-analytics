package sourceconstraints

const OpenSkyAttribution = "Matthias Schäfer, Martin Strohmeier, Vincent Lenders, Ivan Martinovic and Matthias Wilhelm. Bringing Up OpenSky: A Large-scale ADS-B Sensor Network for Research. IPSN 2014, pages 83-94."

func OpenSkyProfile() SourceProfile {
	return SourceProfile{
		ID:                                    "opensky-network",
		Class:                                 SourceClassPublicCommunitySurveillance,
		FreeAccess:                            true,
		ExternallyCollected:                   true,
		RequiresOwnInfrastructure:             false,
		SatelliteDerived:                      false,
		Commercial:                            false,
		OfficialOperational:                   false,
		SupportsRegionalObservation:           true,
		SupportsHistoricalFlights:             true,
		SupportsEstimatedAirports:             true,
		SupportsExperimentalTracks:            true,
		ContinuousCoverage:                    false,
		OceanicCoverage:                       false,
		AttributionRequired:                   true,
		AttributionText:                       OpenSkyAttribution,
		NonCommercialUseOnly:                  true,
		CloudHostingAvailabilityNotGuaranteed: true,
	}
}
