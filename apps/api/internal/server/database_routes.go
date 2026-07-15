package server

import (
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flight"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routepipeline"
	trafficquery "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/query"
	trafficroutecontext "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/routecontext"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseRouteHandlers struct {
	airport           *handlers.AirportHandler
	aircraft          *handlers.AircraftHandler
	flight            *handlers.FlightHandler
	flightState       *handlers.FlightStateHandler
	metrics           *handlers.MetricsHandler
	region            *handlers.RegionHandler
	routeContext      *handlers.RouteContextHandler
	routeIntelligence *handlers.RouteIntelligenceHandler
	traffic           *handlers.TrafficHandler
	trajectory        *handlers.TrajectoryHandler
}

func registerDatabaseRoutes(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
	openMeteoTimeout time.Duration,
) error {
	h := buildDatabaseRouteHandlers(dbPool)

	routeComposition, err :=
		routepipeline.NewPostgres(
			routepipeline.PostgresConfig{
				Pool: dbPool,
			},
		)
	if err != nil {
		return fmt.Errorf(
			"compose production Route Intelligence pipeline: %w",
			err,
		)
	}
	h.routeIntelligence =
		handlers.NewRouteIntelligenceHandler(
			routeComposition.Pipeline,
			routeComposition.Store,
		)

	projectionReader, err :=
		newProjectionIntelligencePostgresReader(
			dbPool,
		)
	if err != nil {
		return fmt.Errorf(
			"compose production Projection Intelligence reader: %w",
			err,
		)
	}

	if err := registerWeatherRoute(
		v1,
		dbPool,
		openMeteoTimeout,
	); err != nil {
		return fmt.Errorf(
			"register weather route: %w",
			err,
		)
	}
	if err := registerAnalyticalMetricRoutes(
		v1,
		dbPool,
	); err != nil {
		return fmt.Errorf(
			"register analytical metric routes: %w",
			err,
		)
	}
	if err := registerHistoricalIntelligenceRoutes(
		v1,
		dbPool,
	); err != nil {
		return fmt.Errorf(
			"register Historical Intelligence routes: %w",
			err,
		)
	}
	if err := RegisterProjectionIntelligenceReadRoute(
		v1,
		projectionReader,
	); err != nil {
		return fmt.Errorf(
			"register Projection Intelligence route: %w",
			err,
		)
	}

	v1.Get(
		"/regions",
		h.region.List,
	)
	v1.Get(
		"/regions/:code",
		h.region.GetByCode,
	)
	v1.Get(
		"/metrics/active-aircraft",
		h.metrics.GetActiveAircraft,
	)
	v1.Get(
		"/traffic/current",
		h.traffic.GetCurrent,
	)
	v1.Get(
		"/aircraft/:icao24/trajectory",
		h.trajectory.GetLatestByICAO24,
	)
	v1.Get(
		"/aircraft/:icao24/route-context",
		h.routeContext.GetByICAO24,
	)
	v1.Get(
		"/trajectories/:id",
		h.trajectory.GetByID,
	)
	v1.Post(
		"/trajectories/:id/route-intelligence",
		h.routeIntelligence.ProcessByTrajectoryID,
	)
	v1.Get(
		"/trajectories/:id/route-intelligence/latest",
		h.routeIntelligence.GetLatestByTrajectoryID,
	)
	v1.Get(
		"/trajectories/:id/route-intelligence/history",
		h.routeIntelligence.ListHistoryByTrajectoryID,
	)
	v1.Get(
		"/flights/:flightID/states",
		h.flightState.ListByFlightID,
	)
	v1.Get(
		"/aircraft/:icao24/latest-state",
		h.flightState.GetLatestByICAO24,
	)
	v1.Get(
		"/flights",
		h.flight.List,
	)
	v1.Get(
		"/flights/:id",
		h.flight.GetByID,
	)
	v1.Get(
		"/aircraft",
		h.aircraft.List,
	)
	v1.Get(
		"/aircraft/:icao24",
		h.aircraft.GetByICAO24,
	)
	v1.Get(
		"/airports",
		h.airport.List,
	)
	v1.Get(
		"/airports/:icao",
		h.airport.GetByICAO,
	)

	return nil
}

func buildDatabaseRouteHandlers(
	dbPool *pgxpool.Pool,
) databaseRouteHandlers {
	airportRepository :=
		postgres.NewAirportRepository(
			dbPool,
		)
	airportService :=
		airport.NewService(
			airportRepository,
		)
	airportHandler :=
		handlers.NewAirportHandler(
			airportService,
		)

	aircraftRepository :=
		postgres.NewAircraftRepository(
			dbPool,
		)
	aircraftService :=
		aircraft.NewService(
			aircraftRepository,
		)
	aircraftHandler :=
		handlers.NewAircraftHandler(
			aircraftService,
		)

	flightRepository :=
		postgres.NewFlightRepository(
			dbPool,
		)
	flightService :=
		flight.NewService(
			flightRepository,
		)
	flightHandler :=
		handlers.NewFlightHandler(
			flightService,
		)

	stateRepository :=
		postgres.NewFlightStateRepository(
			dbPool,
		)
	stateService :=
		flightstate.NewService(
			stateRepository,
		)
	stateHandler :=
		handlers.NewFlightStateHandler(
			stateService,
		)

	regionService :=
		region.NewService()
	regionHandler :=
		handlers.NewRegionHandler(
			regionService,
		)

	metricsRepository :=
		postgres.NewMetricsRepository(
			dbPool,
		)
	metricsService :=
		metrics.NewService(
			metricsRepository,
			regionService,
		)
	metricsHandler :=
		handlers.NewMetricsHandler(
			metricsService,
		)

	trafficRepository :=
		postgres.NewTrafficRepository(
			dbPool,
		)
	trafficService :=
		traffic.NewService(
			trafficRepository,
			regionService,
		)
	trafficHandler :=
		handlers.NewTrafficHandler(
			trafficService,
		)

	trajectoryRepository :=
		postgres.NewTrajectoryRepository(
			dbPool,
		)
	trajectoryService :=
		trafficquery.New(
			trafficquery.Config{
				TrajectoryRepository: trajectoryRepository,
			},
		)
	trajectoryHandler :=
		handlers.NewTrajectoryHandler(
			trajectoryService,
		)

	routeContextService :=
		trafficroutecontext.New(
			trafficroutecontext.Config{
				TrajectoryReader: trajectoryService,
				AirportLister:    airportService,
			},
		)
	routeContextHandler :=
		handlers.NewRouteContextHandler(
			routeContextService,
		)

	return databaseRouteHandlers{
		airport:      airportHandler,
		aircraft:     aircraftHandler,
		flight:       flightHandler,
		flightState:  stateHandler,
		metrics:      metricsHandler,
		region:       regionHandler,
		routeContext: routeContextHandler,
		traffic:      trafficHandler,
		trajectory:   trajectoryHandler,
	}
}
