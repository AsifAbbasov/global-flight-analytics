package server

import (
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

type coreDatabaseRuntime struct {
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

func registerCoreDatabaseContext(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
) error {
	runtime := composeCoreDatabaseRuntime(dbPool)
	registerCoreDatabaseRoutes(v1, runtime)
	return nil
}

func composeCoreDatabaseRuntime(
	dbPool *pgxpool.Pool,
) coreDatabaseRuntime {
	regionService := region.NewService()
	airportService := airport.MustNewService(
		postgres.NewAirportRepository(dbPool),
	)
	trajectoryService := trafficquery.New(
		trafficquery.Config{
			TrajectoryRepository: postgres.NewTrajectoryRepository(
				dbPool,
			),
		},
	)

	return coreDatabaseRuntime{
		airport: handlers.NewAirportHandler(
			airportService,
		),
		aircraft: composeCoreAircraftDatabaseHandler(
			dbPool,
		),
		flight: composeCoreFlightDatabaseHandler(
			dbPool,
		),
		flightState: composeCoreFlightStateDatabaseHandler(
			dbPool,
		),
		metrics: composeCoreMetricsDatabaseHandler(
			dbPool,
			regionService,
		),
		region: handlers.NewRegionHandler(
			regionService,
		),
		routeContext: handlers.NewRouteContextHandler(
			trafficroutecontext.New(
				trafficroutecontext.Config{
					TrajectoryReader: trajectoryService,
					AirportLister:    airportService,
				},
			),
		),
		traffic: composeCoreTrafficDatabaseHandler(
			dbPool,
			regionService,
		),
		trajectory: handlers.NewTrajectoryHandler(
			trajectoryService,
		),
	}
}

func composeCoreAircraftDatabaseHandler(
	dbPool *pgxpool.Pool,
) *handlers.AircraftHandler {
	return handlers.NewAircraftHandler(
		aircraft.MustNewService(
			postgres.NewAircraftRepository(
				dbPool,
			),
		),
	)
}

func composeCoreFlightDatabaseHandler(
	dbPool *pgxpool.Pool,
) *handlers.FlightHandler {
	return handlers.NewFlightHandler(
		flight.MustNewService(
			postgres.NewFlightRepository(
				dbPool,
			),
		),
	)
}

func composeCoreFlightStateDatabaseHandler(
	dbPool *pgxpool.Pool,
) *handlers.FlightStateHandler {
	return handlers.NewFlightStateHandler(
		flightstate.MustNewService(
			postgres.NewFlightStateRepository(
				dbPool,
			),
		),
	)
}

func composeCoreMetricsDatabaseHandler(
	dbPool *pgxpool.Pool,
	regionService *region.Service,
) *handlers.MetricsHandler {
	return handlers.NewMetricsHandler(
		metrics.MustNewService(
			postgres.NewMetricsRepository(
				dbPool,
			),
			regionService,
		),
	)
}

func composeCoreTrafficDatabaseHandler(
	dbPool *pgxpool.Pool,
	regionService *region.Service,
) *handlers.TrafficHandler {
	return handlers.NewTrafficHandler(
		traffic.MustNewService(
			postgres.NewTrafficRepository(
				dbPool,
			),
			regionService,
		),
	)
}
