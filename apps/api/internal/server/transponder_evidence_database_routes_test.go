package server

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestTransponderEvidenceRouteIsReadOnly(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api/v1")

	registerTransponderEvidenceDatabaseRoutes(
		v1,
		transponderEvidenceDatabaseRuntime{},
	)

	fullPath := "/api/v1" + TransponderEvidencePath
	getRouteCount := 0

	for _, route := range app.GetRoutes(true) {
		if route.Path != fullPath {
			continue
		}

		switch route.Method {
		case fiber.MethodGet:
			getRouteCount++
			if len(route.Handlers) != 1 {
				t.Fatalf(
					"GET handler count = %d, want 1",
					len(route.Handlers),
				)
			}

		case fiber.MethodHead:
			// Fiber automatically exposes HEAD for a registered GET route.
			// This remains read-only and is not a second application route.

		case fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete:
			t.Fatalf(
				"mutation method %s is registered for read-only path",
				route.Method,
			)
		}
	}

	if getRouteCount != 1 {
		t.Fatalf(
			"GET route count = %d, want 1",
			getRouteCount,
		)
	}
}
