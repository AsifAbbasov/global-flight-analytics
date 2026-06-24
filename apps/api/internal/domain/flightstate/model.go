package flightstate

import "time"

type FlightState struct {
	ID                  string
	FlightID            string
	AircraftID          string
	ICAO24              string
	Callsign            string
	Latitude            float64
	Longitude           float64
	BarometricAltitudeM int
	GeometricAltitudeM  int
	VelocityMPS         float64
	HeadingDegrees      float64
	VerticalRateMPS     float64
	OnGround            bool
	OriginCountry       string
	ObservedAt          time.Time
	SourceName          string
}
