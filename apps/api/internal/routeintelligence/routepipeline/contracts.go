package routepipeline

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/endpointevidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routeresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
)

const (
	Version = "route-intelligence-pipeline-v1"

	DefaultAirportCatalogTTL = 30 * time.Minute
	DefaultAirportSourceName = "ourairports"
)

type Stage string

const (
	StageTrajectoryLoad        Stage = "trajectory_load"
	StageAirportCatalog        Stage = "airport_catalog"
	StageOriginCandidates      Stage = "origin_candidates"
	StageOriginEvidence        Stage = "origin_evidence"
	StageDestinationCandidates Stage = "destination_candidates"
	StageDestinationEvidence   Stage = "destination_evidence"
	StageRouteResolution       Stage = "route_resolution"
	StageStorage               Stage = "storage"
)

type TrajectoryReader interface {
	GetTrajectoryByID(
		ctx context.Context,
		trajectoryID string,
	) (trajectory.FlightTrajectory, error)
}

type AirportLister interface {
	List(
		ctx context.Context,
	) ([]airport.Airport, error)
}

type Request struct {
	TrajectoryID string
}

type Config struct {
	TrajectoryReader TrajectoryReader
	AirportLister    AirportLister
	Store            routestore.Store

	MaximumCandidateDistanceKM float64
	MaximumCandidates          int
	AirportCatalogTTL          time.Duration
	AirportSourceName          string

	EndpointEvidence endpointevidence.Config
	RouteResolver    routeresolver.Config
	Now              func() time.Time
}

type Versions struct {
	Pipeline         string
	AirportCatalog   string
	AirportResolver  string
	EndpointEvidence string
	RouteResolver    string
	RouteStore       string
}

type Result struct {
	PipelineVersion string
	TrajectoryID    string
	CatalogReport   airportresolver.CatalogBuildReport
	Origin          endpointevidence.Result
	Destination     endpointevidence.Result
	Resolution      routeresolver.Resolution
	Record          routestore.Record
	Versions        Versions
}

func (result Result) Clone() Result {
	return Result{
		PipelineVersion: result.PipelineVersion,
		TrajectoryID:    result.TrajectoryID,
		CatalogReport:   result.CatalogReport.Clone(),
		Origin:          result.Origin.Clone(),
		Destination:     result.Destination.Clone(),
		Resolution:      result.Resolution.Clone(),
		Record:          result.Record.Clone(),
		Versions:        result.Versions,
	}
}

func CurrentVersions() Versions {
	return Versions{
		Pipeline:         Version,
		AirportCatalog:   airportresolver.CatalogVersion,
		AirportResolver:  airportresolver.ResolverVersion,
		EndpointEvidence: endpointevidence.Version,
		RouteResolver:    routeresolver.Version,
		RouteStore:       routestore.Version,
	}
}
