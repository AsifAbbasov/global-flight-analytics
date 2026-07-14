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
	trafficquery "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/query"
	trafficroutecontext "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/routecontext"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseRouteHandlers struct {
	airport      *handlers.AirportHandler
	aircraft     *handlers.AircraftHandler
	flight       *handlers.FlightHandler
	flightState  *handlers.FlightStateHandler
	metrics      *handlers.MetricsHandler
	region       *handlers.RegionHandler
	routeContext *handlers.RouteContextHandler
	traffic      *handlers.TrafficHandler
	trajectory   *handlers.TrajectoryHandler
}

func registerDatabaseRoutes(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
	openMeteoTimeout time.Duration,
) error {
	routeHandlers := buildDatabaseRouteHandlers(
		dbPool,
	)

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

	v1.Get("/regions", routeHandlers.region.List)
	v1.Get("/regions/:code", routeHandlers.region.GetByCode)
	v1.Get("/metrics/active-aircraft", routeHandlers.metrics.GetActiveAircraft)
	v1.Get("/traffic/current", routeHandlers.traffic.GetCurrent)
	v1.Get("/aircraft/:icao24/trajectory", routeHandlers.trajectory.GetLatestByICAO24)
	v1.Get("/aircraft/:icao24/route-context", routeHandlers.routeContext.GetByICAO24)
	v1.Get("/trajectories/:id", routeHandlers.trajectory.GetByID)
	v1.Get("/flights/:flightID/states", routeHandlers.flightState.ListByFlightID)
	v1.Get("/aircraft/:icao24/latest-state", routeHandlers.flightState.GetLatestByICAO24)
	v1.Get("/flights", routeHandlers.flight.List)
	v1.Get("/flights/:id", routeHandlers.flight.GetByID)
	v1.Get("/aircraft", routeHandlers.aircraft.List)
	v1.Get("/aircraft/:icao24", routeHandlers.aircraft.GetByICAO24)
	v1.Get("/airports", routeHandlers.airport.List)
	v1.Get("/airports/:icao", routeHandlers.airport.GetByICAO)

	return nil
}

func buildDatabaseRouteHandlers(
	dbPool *pgxpool.Pool,
) databaseRouteHandlers {
	airportRepository := postgres.NewAirportRepository(dbPool)
	airportService := airport.NewService(airportRepository)
	airportHandler := handlers.NewAirportHandler(airportService)

	aircraftRepository := postgres.NewAircraftRepository(dbPool)
	aircraftService := aircraft.NewService(aircraftRepository)
	aircraftHandler := handlers.NewAircraftHandler(aircraftService)

	flightRepository := postgres.NewFlightRepository(dbPool)
	flightService := flight.NewService(flightRepository)
	flightHandler := handlers.NewFlightHandler(flightService)

	flightStateRepository := postgres.NewFlightStateRepository(dbPool)
	flightStateService := flightstate.NewService(flightStateRepository)
	flightStateHandler := handlers.NewFlightStateHandler(flightStateService)

	regionService := region.NewService()
	regionHandler := handlers.NewRegionHandler(regionService)

	metricsRepository := postgres.NewMetricsRepository(dbPool)
	metricsService := metrics.NewService(
		metricsRepository,
		regionService,
	)
	metricsHandler := handlers.NewMetricsHandler(
		metricsService,
	)

	trafficRepository := postgres.NewTrafficRepository(dbPool)
	trafficService := traffic.NewService(
		trafficRepository,
		regionService,
	)
	trafficHandler := handlers.NewTrafficHandler(trafficService)

	trajectoryRepository := postgres.NewTrajectoryRepository(dbPool)
	trajectoryQueryService := trafficquery.New(
		trafficquery.Config{
			TrajectoryRepository: trajectoryRepository,
		},
	)
	trajectoryHandler := handlers.NewTrajectoryHandler(
		trajectoryQueryService,
	)

	routeContextService := trafficroutecontext.New(
		trafficroutecontext.Config{
			TrajectoryReader: trajectoryQueryService,
			AirportLister:    airportService,
		},
	)
	routeContextHandler := handlers.NewRouteContextHandler(
		routeContextService,
	)

	return databaseRouteHandlers{
		airport:      airportHandler,
		aircraft:     aircraftHandler,
		flight:       flightHandler,
		flightState:  flightStateHandler,
		metrics:      metricsHandler,
		region:       regionHandler,
		routeContext: routeContextHandler,
		traffic:      trafficHandler,
		trajectory:   trajectoryHandler,
	}
}
