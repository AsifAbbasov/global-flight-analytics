package flightsplitter

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

type Observation struct {
	State        flightstate.FlightState
	QualityScore float64
}

type Group struct {
	ICAO24        string
	IdentityKey   string
	IdentityBasis trajectory.FlightIdentityBasis
	SplitReason   trajectory.FlightSplitReason
	Observations  []Observation
}
