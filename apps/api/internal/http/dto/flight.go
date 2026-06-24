package dto

import "time"

type FlightListItem struct {
	ID            string    `json:"id"`
	AircraftID    string    `json:"aircraft_id"`
	ICAO24        string    `json:"icao24"`
	Callsign      string    `json:"callsign"`
	Status        string    `json:"status"`
	FirstSeenAt   time.Time `json:"first_seen_at"`
	LastSeenAt    time.Time `json:"last_seen_at"`
	AircraftModel string    `json:"aircraft_model"`
	Airline       string    `json:"airline"`
	Country       string    `json:"country"`
}

type FlightProfile struct {
	ID            string    `json:"id"`
	AircraftID    string    `json:"aircraft_id"`
	ICAO24        string    `json:"icao24"`
	Callsign      string    `json:"callsign"`
	Status        string    `json:"status"`
	FirstSeenAt   time.Time `json:"first_seen_at"`
	LastSeenAt    time.Time `json:"last_seen_at"`
	AircraftModel string    `json:"aircraft_model"`
	Airline       string    `json:"airline"`
	Country       string    `json:"country"`
}
