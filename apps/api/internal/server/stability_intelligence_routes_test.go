package server

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRegisterStabilityIntelligenceReadRouteRequiresReader(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api/v1")
	if err := RegisterStabilityIntelligenceReadRoute(
		v1,
		nil,
	); err == nil {
		t.Fatal(
			"nil Stability Intelligence reader was accepted",
		)
	}
}
