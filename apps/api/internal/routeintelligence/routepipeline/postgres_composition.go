package routepipeline

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/endpointevidence"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routeresolver"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
	trafficquery "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/query"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PostgresCompositionVersion = "route-intelligence-postgres-pipeline-composition-v1"

	ComponentTrajectoryRepository = "trajectory_repository"
	ComponentTrajectoryService    = "trajectory_service"
	ComponentAirportRepository    = "airport_repository"
	ComponentAirportService       = "airport_service"
	ComponentStore                = "route_store"
	ComponentPipeline             = "route_pipeline"
)

type PostgresConfig struct {
	Pool          *pgxpool.Pool
	StoreExecutor routestore.PostgresExecutor

	MaximumCandidateDistanceKM float64
	MaximumCandidates          int
	AirportCatalogTTL          time.Duration
	AirportSourceName          string

	EndpointEvidence endpointevidence.Config
	RouteResolver    routeresolver.Config
	Now              func() time.Time
}

type PostgresVersions struct {
	Composition string
	Pipeline    Versions
	Store       string
}

type PostgresComposition struct {
	Pipeline *Pipeline
	Store    *routestore.PostgresStore

	TrajectoryRepository *postgres.TrajectoryRepository
	TrajectoryService    *trafficquery.Service
	AirportRepository    *postgres.AirportRepository
	AirportService       *airport.Service

	Versions PostgresVersions
}

func NewPostgres(
	config PostgresConfig,
) (*PostgresComposition, error) {
	if config.Pool == nil {
		return nil, ErrPostgresPoolRequired
	}

	now := config.Now
	if now == nil {
		now = time.Now
	}

	trajectoryRepository :=
		postgres.NewTrajectoryRepository(
			config.Pool,
		)
	trajectoryService := trafficquery.New(
		trafficquery.Config{
			TrajectoryRepository: trajectoryRepository,
		},
	)

	airportRepository :=
		postgres.NewAirportRepository(
			config.Pool,
		)
	airportService := airport.MustNewService(
		airportRepository,
	)

	var store *routestore.PostgresStore
	var err error

	if config.StoreExecutor != nil {
		store, err =
			routestore.NewPostgresWithExecutor(
				config.StoreExecutor,
				now,
			)
	} else {
		store, err = routestore.NewPostgres(
			routestore.PostgresConfig{
				Pool: config.Pool,
				Now:  now,
			},
		)
	}
	if err != nil {
		return nil, &ConstructionError{
			Component: ComponentStore,
			Err:       err,
		}
	}

	pipeline, err := New(
		Config{
			TrajectoryReader: trajectoryService,
			AirportLister:    airportService,
			Store:            store,

			MaximumCandidateDistanceKM: config.
				MaximumCandidateDistanceKM,
			MaximumCandidates: config.MaximumCandidates,
			AirportCatalogTTL: config.AirportCatalogTTL,
			AirportSourceName: config.AirportSourceName,

			EndpointEvidence: config.EndpointEvidence,
			RouteResolver:    config.RouteResolver,
			Now:              now,
		},
	)
	if err != nil {
		return nil, &ConstructionError{
			Component: ComponentPipeline,
			Err:       err,
		}
	}

	return &PostgresComposition{
		Pipeline: pipeline,
		Store:    store,

		TrajectoryRepository: trajectoryRepository,
		TrajectoryService:    trajectoryService,
		AirportRepository:    airportRepository,
		AirportService:       airportService,

		Versions: CurrentPostgresVersions(),
	}, nil
}

func CurrentPostgresVersions() PostgresVersions {
	return PostgresVersions{
		Composition: PostgresCompositionVersion,
		Pipeline:    CurrentVersions(),
		Store:       routestore.PostgresVersion,
	}
}
