package dto

type AircraftListItem struct {
	ICAO24       string `json:"icao24"`
	Registration string `json:"registration"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	Airline      string `json:"airline"`
	Country      string `json:"country"`
}

type AircraftProfile struct {
	ICAO24       string `json:"icao24"`
	Registration string `json:"registration"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	AircraftType string `json:"aircraft_type"`
	Airline      string `json:"airline"`
	Country      string `json:"country"`
}
