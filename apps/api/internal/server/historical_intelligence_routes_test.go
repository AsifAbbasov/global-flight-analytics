package server

import (
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRegisterHistoricalIntelligenceRoutes(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api/v1")

	err := registerHistoricalIntelligenceRoutes(
		v1,
		&pgxpool.Pool{},
	)
	if err != nil {
		t.Fatalf(
			"register Historical Intelligence routes: %v",
			err,
		)
	}

	expected := map[string]bool{
		"/api/v1" +
			historicalIntelligenceLatestPath: false,
		"/api/v1" +
			historicalIntelligenceHistoryPath: false,
	}

	for _, route := range app.GetRoutes() {
		if route.Method != fiber.MethodGet {
			continue
		}
		if _, exists := expected[route.Path]; exists {
			expected[route.Path] = true
		}
	}

	for path, found := range expected {
		if !found {
			t.Fatalf(
				"Historical Intelligence route %s was not registered",
				path,
			)
		}
	}
}

func TestNewRegistersHistoricalIntelligenceRoutes(
	t *testing.T,
) {
	app, err := New(
		Config{
			DatabasePool: &pgxpool.Pool{},
			Logger:       newDiscardLogger(),
			OpenMeteoTimeout: 5 *
				time.Second,
		},
	)
	if err != nil {
		t.Fatalf(
			"initialize server: %v",
			err,
		)
	}

	expected := map[string]bool{
		"/api/v1" +
			historicalIntelligenceLatestPath: false,
		"/api/v1" +
			historicalIntelligenceHistoryPath: false,
	}
	for _, route := range app.GetRoutes() {
		if route.Method != fiber.MethodGet {
			continue
		}
		if _, exists := expected[route.Path]; exists {
			expected[route.Path] = true
		}
	}
	for path, found := range expected {
		if !found {
			t.Fatalf(
				"server route %s was not registered",
				path,
			)
		}
	}
}

func TestRegisterHistoricalIntelligenceRoutesRejectsNilPool(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api/v1")

	err := registerHistoricalIntelligenceRoutes(
		v1,
		nil,
	)
	if err == nil {
		t.Fatal(
			"expected nil database pool to be rejected",
		)
	}
	if !strings.Contains(
		err.Error(),
		"Historical Intelligence aggregate store",
	) {
		t.Fatalf(
			"unexpected registration error: %v",
			err,
		)
	}
}
