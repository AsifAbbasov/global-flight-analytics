package flightstate

import "time"

type AltitudeStatus string

const (
	AltitudeStatusObserved    AltitudeStatus = "observed"
	AltitudeStatusGround      AltitudeStatus = "ground"
	AltitudeStatusUnknown     AltitudeStatus = "unknown"
	AltitudeStatusUnavailable AltitudeStatus = "unavailable"
	AltitudeStatusInvalid     AltitudeStatus = "invalid"
)

type FlightState struct {
	ID                       string
	FlightID                 string
	AircraftID               string
	IngestionRunID           string
	ICAO24                   string
	Callsign                 string
	Latitude                 float64
	Longitude                float64
	BarometricAltitudeM      float64
	BarometricAltitudeStatus AltitudeStatus
	GeometricAltitudeM       float64
	GeometricAltitudeStatus  AltitudeStatus
	VelocityMPS              float64
	HeadingDegrees           float64
	VerticalRateMPS          float64
	OnGround                 bool
	OriginCountry            string
	ObservedAt               time.Time
	SourceName               string
}
