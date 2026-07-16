package airspaceproduction

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceregionanalytics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactiongraph"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/interactionradius"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/localtrafficscene"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/proximityscanner"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/separationrisk"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

const Version = "airspace-production-composition-v1"

type Request struct {
	RegionCode string
	AsOfTime   time.Time
	Window     time.Duration
}

type ObservationQuery struct {
	Bounds      region.Bounds
	WindowStart time.Time
	WindowEnd   time.Time
	Limit       int
}

type Observation struct {
	StateID    string
	FlightID   string
	AircraftID string
	ICAO24     string
	Callsign   string

	Latitude  float64
	Longitude float64

	AltitudeMeters    *float64
	AltitudeReference interactiongraph.AltitudeReference

	VelocityMetersPerSecond     float64
	HeadingDegrees              float64
	VerticalRateMetersPerSecond float64
	OnGround                    bool

	ObservedAt time.Time
	SourceName string
}

func (observation Observation) Clone() Observation {
	cloned := observation
	if observation.AltitudeMeters != nil {
		value := *observation.AltitudeMeters
		cloned.AltitudeMeters = &value
	}
	return cloned
}

type ObservationReader interface {
	ListAirspaceObservations(
		context.Context,
		ObservationQuery,
	) ([]Observation, error)
}

type RegionResolver interface {
	GetByCode(string) (region.Region, error)
}

type Config struct {
	ObservationReader ObservationReader
	RegionResolver    RegionResolver
	Now               func() time.Time

	DefaultWindow       time.Duration
	MinimumWindow       time.Duration
	MaximumWindow       time.Duration
	MaximumObservations int

	ScenePolicy   localtrafficscene.Policy
	RadiusPolicy  interactionradius.Policy
	ScannerPolicy proximityscanner.Policy
	RiskPolicy    separationrisk.Policy
	RegionPolicy  airspaceregionanalytics.Policy
}

type Service struct {
	observationReader ObservationReader
	regionResolver    RegionResolver
	now               func() time.Time

	defaultWindow       time.Duration
	minimumWindow       time.Duration
	maximumWindow       time.Duration
	maximumObservations int

	scenePolicy   localtrafficscene.Policy
	radiusPolicy  interactionradius.Policy
	scannerPolicy proximityscanner.Policy
	riskPolicy    separationrisk.Policy
	regionPolicy  airspaceregionanalytics.Policy
}
