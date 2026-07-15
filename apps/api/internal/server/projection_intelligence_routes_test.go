package server

import (
	"context"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/gofiber/fiber/v2"
)

type projectionIntelligenceRouteReaderStub struct{}

func (
	projectionIntelligenceRouteReaderStub,
) GetProjectionIntelligence(
	context.Context,
	handlers.ProjectionIntelligenceReadRequest,
) (projectionproduction.Result, error) {
	return projectionproduction.Result{},
		nil
}

func TestRegisterProjectionIntelligenceReadRoute(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api/v1")

	err := RegisterProjectionIntelligenceReadRoute(
		v1,
		projectionIntelligenceRouteReaderStub{},
	)
	if err != nil {
		t.Fatalf(
			"register Projection Intelligence read route: %v",
			err,
		)
	}

	expectedPath :=
		"/api/v1" +
			ProjectionIntelligencePath
	found := false
	for _, route := range app.GetRoutes() {
		if route.Method == fiber.MethodGet &&
			route.Path == expectedPath {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf(
			"Projection Intelligence route %s was not registered",
			expectedPath,
		)
	}
}

func TestRegisterProjectionIntelligenceReadRouteRejectsNilReader(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api/v1")

	err := RegisterProjectionIntelligenceReadRoute(
		v1,
		nil,
	)
	if err == nil {
		t.Fatal(
			"expected nil Projection Intelligence reader to be rejected",
		)
	}
	if !strings.Contains(
		err.Error(),
		"reader is required",
	) {
		t.Fatalf(
			"unexpected registration error: %v",
			err,
		)
	}
}
