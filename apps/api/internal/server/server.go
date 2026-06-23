package server

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(dbPool *pgxpool.Pool) *fiber.App {
	app := fiber.New()

	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Get("/health", handlers.Health)
	v1.Get("/version", handlers.Version)

	if dbPool != nil {
		airportRepository := postgres.NewAirportRepository(dbPool)
		airportService := airport.NewService(airportRepository)
		airportHandler := handlers.NewAirportHandler(airportService)

		v1.Get("/airports", airportHandler.List)
		v1.Get("/airports/:icao", airportHandler.GetByICAO)
	}

	return app
}
