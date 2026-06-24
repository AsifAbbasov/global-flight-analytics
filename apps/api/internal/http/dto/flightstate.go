package dto

import "time"

type FlightStateItem struct {
	ID                  string    `json:"id"`
	FlightID            string    `json:"flight_id"`
	AircraftID          string    `json:"aircraft_id"`
	ICAO24              string    `json:"icao24"`
	Callsign            string    `json:"callsign"`
	Latitude            float64   `json:"latitude"`
	Longitude           float64   `json:"longitude"`
	BarometricAltitudeM int       `json:"barometric_altitude_m"`
	GeometricAltitudeM  int       `json:"geometric_altitude_m"`
	VelocityMPS         float64   `json:"velocity_mps"`
	HeadingDegrees      float64   `json:"heading_degrees"`
	VerticalRateMPS     float64   `json:"vertical_rate_mps"`
	OnGround            bool      `json:"on_ground"`
	OriginCountry       string    `json:"origin_country"`
	ObservedAt          time.Time `json:"observed_at"`
	SourceName          string    `json:"source_name"`
}
