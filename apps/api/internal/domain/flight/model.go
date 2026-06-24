package flight

import "time"

type Flight struct {
	ID            string
	AircraftID    string
	ICAO24        string
	Callsign      string
	Status        string
	FirstSeenAt   time.Time
	LastSeenAt    time.Time
	AircraftModel string
	Airline       string
	Country       string
}
