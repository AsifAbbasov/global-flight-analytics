package traffic

import "time"

type CurrentTrafficItem struct {
	ICAO24         string
	Callsign       string
	Latitude       float64
	Longitude      float64
	AltitudeM      int
	VelocityMPS    float64
	HeadingDegrees float64
	OnGround       bool
	ObservedAt     time.Time
	AircraftModel  string
	Airline        string
	OriginCountry  string
}
