package traffic

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

type CurrentTrafficItem struct {
	ICAO24         string
	Callsign       string
	Latitude       float64
	Longitude      float64
	AltitudeM      *float64
	AltitudeStatus flightstate.AltitudeStatus
	AltitudeSource AltitudeSource
	VelocityMPS    float64
	HeadingDegrees float64
	OnGround       bool
	ObservedAt     time.Time
	AircraftModel  string
	Airline        string
	OriginCountry  string
}
