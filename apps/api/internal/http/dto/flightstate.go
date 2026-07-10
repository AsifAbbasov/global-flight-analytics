package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type FlightStateItem struct {
	ID                       string                     `json:"id"`
	FlightID                 string                     `json:"flight_id"`
	AircraftID               string                     `json:"aircraft_id"`
	ICAO24                   string                     `json:"icao24"`
	Callsign                 string                     `json:"callsign"`
	Latitude                 float64                    `json:"latitude"`
	Longitude                float64                    `json:"longitude"`
	BarometricAltitudeM      *float64                   `json:"barometric_altitude_m"`
	BarometricAltitudeStatus flightstate.AltitudeStatus `json:"barometric_altitude_status"`
	GeometricAltitudeM       *float64                   `json:"geometric_altitude_m"`
	GeometricAltitudeStatus  flightstate.AltitudeStatus `json:"geometric_altitude_status"`
	VelocityMPS              float64                    `json:"velocity_mps"`
	HeadingDegrees           float64                    `json:"heading_degrees"`
	VerticalRateMPS          float64                    `json:"vertical_rate_mps"`
	OnGround                 bool                       `json:"on_ground"`
	OriginCountry            string                     `json:"origin_country"`
	ObservedAt               time.Time                  `json:"observed_at"`
	SourceName               string                     `json:"source_name"`
}
