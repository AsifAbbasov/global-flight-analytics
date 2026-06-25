package server

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flight"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(dbPool *pgxpool.Pool) *fiber.App {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
	AllowOrigins: "http://localhost:3000,http://localhost:3001",
	AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	AllowHeaders: "Origin,Content-Type,Accept,Authorization",
}))

	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Get("/health", handlers.Health)
	v1.Get("/version", handlers.Version)

	if dbPool != nil {
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

		trafficRepository := postgres.NewTrafficRepository(dbPool)
		trafficService := traffic.NewService(trafficRepository, regionService)
		trafficHandler := handlers.NewTrafficHandler(trafficService)

		v1.Get("/regions", regionHandler.List)
		v1.Get("/regions/:code", regionHandler.GetByCode)

		v1.Get("/traffic/current", trafficHandler.GetCurrent)

		v1.Get("/flights/:flightID/states", flightStateHandler.ListByFlightID)
		v1.Get("/aircraft/:icao24/latest-state", flightStateHandler.GetLatestByICAO24)

		v1.Get("/flights", flightHandler.List)
		v1.Get("/flights/:id", flightHandler.GetByID)

		v1.Get("/aircraft", aircraftHandler.List)
		v1.Get("/aircraft/:icao24", aircraftHandler.GetByICAO24)

		v1.Get("/airports", airportHandler.List)
		v1.Get("/airports/:icao", airportHandler.GetByICAO)
	}

	return app
}
