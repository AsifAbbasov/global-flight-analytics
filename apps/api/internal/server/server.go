package server

import (
	"log/slog"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flight"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/middleware"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/weatherprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	trafficquery "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/query"
	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(
	dbPool *pgxpool.Pool,
	log *slog.Logger,
) *fiber.App {
	app := fiber.New()

	app.Use(
		recover.New(),
	)

	app.Use(
		middleware.RequestID(),
	)

	app.Use(
		middleware.RequestLogger(log),
	)

	app.Use(
		cors.New(
			cors.Config{
				AllowOrigins: "http://localhost:3000,http://localhost:3001",
				AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
				AllowHeaders: "Origin,Content-Type,Accept,Authorization",
			},
		),
	)

	api := app.Group(
		"/api",
	)

	v1 := api.Group(
		"/v1",
	)

	v1.Get(
		"/health",
		handlers.Health,
	)

	v1.Get(
		"/version",
		handlers.Version,
	)

	if dbPool != nil {
		airportRepository := postgres.NewAirportRepository(
			dbPool,
		)

		airportService := airport.NewService(
			airportRepository,
		)

		airportHandler := handlers.NewAirportHandler(
			airportService,
		)

		aircraftRepository := postgres.NewAircraftRepository(
			dbPool,
		)

		aircraftService := aircraft.NewService(
			aircraftRepository,
		)

		aircraftHandler := handlers.NewAircraftHandler(
			aircraftService,
		)

		flightRepository := postgres.NewFlightRepository(
			dbPool,
		)

		flightService := flight.NewService(
			flightRepository,
		)

		flightHandler := handlers.NewFlightHandler(
			flightService,
		)

		flightStateRepository := postgres.NewFlightStateRepository(
			dbPool,
		)

		flightStateService := flightstate.NewService(
			flightStateRepository,
		)

		flightStateHandler := handlers.NewFlightStateHandler(
			flightStateService,
		)

		regionService := region.NewService()

		regionHandler := handlers.NewRegionHandler(
			regionService,
		)

		trafficRepository := postgres.NewTrafficRepository(
			dbPool,
		)

		trafficService := traffic.NewService(
			trafficRepository,
			regionService,
		)

		trafficHandler := handlers.NewTrafficHandler(
			trafficService,
		)

		trajectoryRepository := postgres.NewTrajectoryRepository(
			dbPool,
		)

		trajectoryQueryService := trafficquery.New(
			trafficquery.Config{
				TrajectoryRepository: trajectoryRepository,
			},
		)

		trajectoryHandler := handlers.NewTrajectoryHandler(
			trajectoryQueryService,
		)

		registerWeatherRoute(
			v1,
			dbPool,
			log,
		)

		v1.Get(
			"/regions",
			regionHandler.List,
		)

		v1.Get(
			"/regions/:code",
			regionHandler.GetByCode,
		)

		v1.Get(
			"/traffic/current",
			trafficHandler.GetCurrent,
		)

		v1.Get(
			"/aircraft/:icao24/trajectory",
			trajectoryHandler.GetLatestByICAO24,
		)

		v1.Get(
			"/trajectories/:id",
			trajectoryHandler.GetByID,
		)

		v1.Get(
			"/flights/:flightID/states",
			flightStateHandler.ListByFlightID,
		)

		v1.Get(
			"/aircraft/:icao24/latest-state",
			flightStateHandler.GetLatestByICAO24,
		)

		v1.Get(
			"/flights",
			flightHandler.List,
		)

		v1.Get(
			"/flights/:id",
			flightHandler.GetByID,
		)

		v1.Get(
			"/aircraft",
			aircraftHandler.List,
		)

		v1.Get(
			"/aircraft/:icao24",
			aircraftHandler.GetByICAO24,
		)

		v1.Get(
			"/airports",
			airportHandler.List,
		)

		v1.Get(
			"/airports/:icao",
			airportHandler.GetByICAO,
		)
	}

	return app
}

func registerWeatherRoute(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
	log *slog.Logger,
) {
	budgetManager, err := providerbudget.New(
		nil,
	)
	if err != nil {
		logInitializationError(
			log,
			"failed to initialize provider budget manager",
			err,
		)

		return
	}

	responseController, err := providerresponse.New(
		providerresponse.Config{
			BudgetManager: budgetManager,
		},
	)
	if err != nil {
		logInitializationError(
			log,
			"failed to initialize provider response controller",
			err,
		)

		return
	}

	responseObserver, err := providerresponse.NewIntegrationObserver(
		responseController,
	)
	if err != nil {
		logInitializationError(
			log,
			"failed to initialize provider response observer",
			err,
		)

		return
	}

	orchestrator, err := ingestionorchestrator.NewDefault(
		responseController,
	)
	if err != nil {
		logInitializationError(
			log,
			"failed to initialize ingestion orchestrator",
			err,
		)

		return
	}

	openMeteoClient, err := openmeteo.New(
		openmeteo.Config{
			ResponseObserver: responseObserver,
		},
	)
	if err != nil {
		logInitializationError(
			log,
			"failed to initialize open-meteo client",
			err,
		)

		return
	}

	orchestratedWeatherClient, err := weatherprovider.New(
		weatherprovider.Config{
			Client:   openMeteoClient,
			Executor: orchestrator,
		},
	)
	if err != nil {
		logInitializationError(
			log,
			"failed to initialize orchestrated weather client",
			err,
		)

		return
	}

	weatherRepository := postgres.NewWeatherRepository(
		dbPool,
	)

	weatherService := weatherservice.New(
		weatherservice.Config{
			Client:     orchestratedWeatherClient,
			Repository: weatherRepository,
		},
	)

	weatherHandler := handlers.NewWeatherHandler(
		weatherService,
	)

	v1.Get(
		"/weather/current",
		weatherHandler.GetCurrent,
	)
}

func logInitializationError(
	log *slog.Logger,
	message string,
	err error,
) {
	if log == nil {
		return
	}

	log.Error(
		message,
		"error",
		err,
	)
}
