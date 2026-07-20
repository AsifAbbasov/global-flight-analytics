package dto

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
)

type CurrentTrafficItem struct {
	ICAO24         string                     `json:"icao24"`
	Callsign       string                     `json:"callsign"`
	Latitude       float64                    `json:"latitude"`
	Longitude      float64                    `json:"longitude"`
	AltitudeM      *float64                   `json:"altitude_m"`
	AltitudeStatus flightstate.AltitudeStatus `json:"altitude_status"`
	AltitudeSource traffic.AltitudeSource     `json:"altitude_source"`
	VelocityMPS    float64                    `json:"velocity_mps"`
	HeadingDegrees float64                    `json:"heading_degrees"`
	OnGround       bool                       `json:"on_ground"`
	ObservedAt     time.Time                  `json:"observed_at"`
	AircraftModel  string                     `json:"aircraft_model"`
	Airline        string                     `json:"airline"`
	OriginCountry  string                     `json:"origin_country"`
}
