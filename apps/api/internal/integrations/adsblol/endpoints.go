package adsblol

const (
	BaseURL = "https://api.adsb.lol"

	EndpointByCallsign     = "/v2/callsign/%s"
	EndpointByICAO24       = "/v2/hex/%s"
	EndpointByPoint        = "/v2/point/%f/%f/%d"
	EndpointByRegistration = "/v2/reg/%s"
)
