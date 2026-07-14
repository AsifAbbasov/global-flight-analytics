package extractor

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

const Version = "flight-feature-extractor-v1"

type Request struct {
	Trajectory trajectory.FlightTrajectory
	AsOfTime   time.Time
}

type AircraftReference struct {
	AircraftID string
	ICAO24     string
	Callsign   string
}

type TemporalBuilder interface {
	Build(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (flightfeatures.TemporalFeatures, error)
}

type GeographicalBuilder interface {
	Build(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (flightfeatures.GeographicalFeatures, error)
}

type OperationalBuilder interface {
	Build(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (flightfeatures.OperationalFeatures, error)
}

type TrajectoryBuilder interface {
	Build(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) (flightfeatures.TrajectoryFeatures, error)
}

type AircraftFeatureProvider interface {
	Provide(
		ctx context.Context,
		reference AircraftReference,
	) (flightfeatures.AircraftFeatures, error)
}

type Config struct {
	TemporalBuilder         TemporalBuilder
	GeographicalBuilder     GeographicalBuilder
	OperationalBuilder      OperationalBuilder
	TrajectoryBuilder       TrajectoryBuilder
	AircraftFeatureProvider AircraftFeatureProvider
	Now                     func() time.Time
}
