package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/airportproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerAirportIntelligenceRoutes(v1 fiber.Router, pool *pgxpool.Pool) error {
	service, err := airportproduction.NewPostgres(pool)
	if err != nil {
		return fmt.Errorf("compose production Airport Intelligence service: %w", err)
	}
	handler := handlers.NewAirportIntelligenceHandler(service)
	v1.Get("/airports/intelligence/ranking", handler.GetRanking)
	v1.Get("/airports/:icao/intelligence/overview", handler.GetOverview)
	v1.Get("/airports/:icao/intelligence/history", handler.GetHistory)
	v1.Get("/airports/:icao/intelligence/trends", handler.GetTrends)
	return nil
}
