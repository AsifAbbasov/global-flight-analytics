package server

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
)

func New() *fiber.App {
	app := fiber.New()

	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Get("/health", handlers.Health)
	v1.Get("/version", handlers.Version)

	return app
}
