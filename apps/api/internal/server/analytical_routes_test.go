package server

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRegisterAnalyticalMetricRoutes(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api/v1")

	if err := registerAnalyticalMetricRoutes(
		v1,
		&pgxpool.Pool{},
	); err != nil {
		t.Fatalf(
			"expected analytical routes, got %v",
			err,
		)
	}

	expected := map[string]bool{
		"/api/v1/analytics/metrics/active-aircraft":  false,
		"/api/v1/analytics/metrics/traffic-density":  false,
		"/api/v1/analytics/metrics/airport-activity": false,
		"/api/v1/analytics/metrics/coverage-score":   false,
		"/api/v1/analytics/metrics/data-freshness":   false,
	}

	for _, route := range app.GetRoutes() {
		if route.Method != fiber.MethodGet {
			continue
		}
		if _, exists := expected[route.Path]; exists {
			expected[route.Path] = true
		}
	}

	for route, found := range expected {
		if !found {
			t.Fatalf(
				"expected route %s",
				route,
			)
		}
	}
}
