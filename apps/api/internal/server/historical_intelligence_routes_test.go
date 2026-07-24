package server

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type historicalRouteReaderStub struct{}

func (historicalRouteReaderStub) GetLatest(
	context.Context,
	historicalaggregate.ListQuery,
) (historicalaggregate.Record, error) {
	return historicalaggregate.Record{}, nil
}

func (historicalRouteReaderStub) List(
	context.Context,
	historicalaggregate.ListQuery,
) (historicalaggregate.Page, error) {
	return historicalaggregate.Page{}, nil
}

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

	assertHistoricalIntelligenceRoutes(
		t,
		app,
	)
}

func TestRegisterHistoricalIntelligenceReadRoutes(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api/v1")

	err := RegisterHistoricalIntelligenceReadRoutes(
		v1,
		historicalRouteReaderStub{},
	)
	if err != nil {
		t.Fatalf(
			"register reader-backed Historical Intelligence routes: %v",
			err,
		)
	}

	assertHistoricalIntelligenceRoutes(
		t,
		app,
	)
}

func TestRegisterHistoricalIntelligenceRoutesRejectsNilDependencies(
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
			"unexpected pool registration error: %v",
			err,
		)
	}

	err = RegisterHistoricalIntelligenceReadRoutes(
		v1,
		nil,
	)
	if !errors.Is(
		err,
		ErrHistoricalIntelligenceReaderRequired,
	) {
		t.Fatalf(
			"expected reader dependency error, got %v",
			err,
		)
	}
}

func assertHistoricalIntelligenceRoutes(
	t *testing.T,
	app *fiber.App,
) {
	t.Helper()

	expected := map[string]bool{
		"/api/v1" +
			HistoricalIntelligenceLatestPath: false,
		"/api/v1" +
			HistoricalIntelligenceHistoryPath: false,
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
