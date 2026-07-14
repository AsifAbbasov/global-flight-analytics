package routeresolver

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/endpointevidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

const (
	Version = "route-resolver-v1"

	DefaultPartialConfidenceFactor     = 0.50
	DefaultSameAirportConfidenceFactor = 0.75
)

type Config struct {
	Now                         func() time.Time
	PartialConfidenceFactor     float64
	SameAirportConfidenceFactor float64
}

type Input struct {
	TrajectoryID string
	IdentityKey  string
	FlightID     string
	AircraftID   string
	ICAO24       string
	Callsign     string

	Window              routecontract.RouteWindow
	TrajectoryUpdatedAt time.Time
	Origin              endpointevidence.Result
	Destination         endpointevidence.Result
	SourceNames         []string
}

type Resolution struct {
	Version    string
	Result     routecontract.Result
	Validation routecontract.ValidationReport
}

func (resolution Resolution) Clone() Resolution {
	return Resolution{
		Version:    resolution.Version,
		Result:     resolution.Result.Clone(),
		Validation: resolution.Validation.Clone(),
	}
}
