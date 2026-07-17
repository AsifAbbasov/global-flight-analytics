package opensky

const (
	ProviderName     = "The OpenSky Network"
	ProviderURL      = "https://opensky-network.org"
	RequiredCitation = "Matthias Schäfer, Martin Strohmeier, Vincent Lenders, Ivan Martinovic and Matthias Wilhelm. Bringing Up OpenSky: A Large-scale ADS-B Sensor Network for Research. In Proceedings of the 13th IEEE/ACM International Symposium on Information Processing in Sensor Networks, pages 83-94, April 2014."
)

type UsagePolicy struct {
	ResearchAndNonCommercialUseOnly bool     `json:"research_and_non_commercial_use_only"`
	AttributionRequired             bool     `json:"attribution_required"`
	RequiredCitation                string   `json:"required_citation"`
	ProviderURL                     string   `json:"provider_url"`
	CommercialFlightDataUnavailable bool     `json:"commercial_flight_data_unavailable"`
	CloudAccessNotGuaranteed        bool     `json:"cloud_access_not_guaranteed"`
	Limitations                     []string `json:"limitations"`
}

func OfficialUsagePolicy() UsagePolicy {
	return UsagePolicy{
		ResearchAndNonCommercialUseOnly: true,
		AttributionRequired:             true,
		RequiredCitation:                RequiredCitation,
		ProviderURL:                     ProviderURL,
		CommercialFlightDataUnavailable: true,
		CloudAccessNotGuaranteed:        true,
		Limitations: []string{
			"The free real-time API does not provide official airport schedules, delays, gates, or causes that cannot be derived from ADS-B observations.",
			"Access from large cloud-provider IP ranges may be blocked because of abuse and therefore cannot be the sole production ingestion dependency.",
			"Public web pages, articles, presentations, and other publications using OpenSky data must preserve the required attribution.",
		},
	}
}
