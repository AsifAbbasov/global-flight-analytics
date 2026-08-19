package server

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	livetraffic "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/live"
	"github.com/gofiber/fiber/v2"
)

func registerLiveTrafficRoutes(
	v1 fiber.Router,
	store *livetraffic.Store,
) {
	handler := handlers.NewLiveTrafficHandler(store)
	v1.Get("/traffic/live", handler.GetSnapshot)
}
