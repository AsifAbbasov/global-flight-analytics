package aircraft

import "time"

type Aircraft struct {
	ICAO24            string
	Registration      string
	Model             string
	Manufacturer      string
	AircraftType      string
	Airline           string
	Country           string
	MetadataUpdatedAt time.Time
}
