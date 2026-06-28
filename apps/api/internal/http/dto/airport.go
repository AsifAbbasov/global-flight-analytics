package dto

type AirportListItem struct {
	ICAOCode  string  `json:"icao_code"`
	IATACode  string  `json:"iata_code"`
	Name      string  `json:"name"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type AirportProfile struct {
	ICAOCode    string  `json:"icao_code"`
	IATACode    string  `json:"iata_code"`
	Name        string  `json:"name"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ElevationM  float64 `json:"elevation_m"`
	Timezone    string  `json:"timezone"`
	Description string  `json:"description"`
}
