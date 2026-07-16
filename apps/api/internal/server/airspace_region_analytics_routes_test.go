package server

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRegisterAirspaceRegionAnalyticsReadRouteRejectsNilReader(
	t *testing.T,
) {
	app := fiber.New()
	err := RegisterAirspaceRegionAnalyticsReadRoute(
		app.Group("/api/v1"),
		nil,
	)
	if err == nil {
		t.Fatal("error = nil, want non-nil")
	}
}
