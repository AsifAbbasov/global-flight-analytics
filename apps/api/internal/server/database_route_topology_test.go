package server

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type routeTopologyExpectation struct {
	method       string
	path         string
	handlerCount int
}

func TestCoreDatabaseRouteTopology(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api/v1")

	registerCoreDatabaseRoutes(
		v1,
		coreDatabaseRuntime{},
	)

	mutationAuthorization :=
		func(ctx *fiber.Ctx) error {
			return ctx.Next()
		}
	registerRouteIntelligenceDatabaseRoutes(
		v1,
		routeIntelligenceDatabaseRuntime{},
		mutationAuthorization,
	)

	expectations := []routeTopologyExpectation{
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/regions",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/regions/:code",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/metrics/active-aircraft",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/traffic/current",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/aircraft/:icao24/trajectory",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/aircraft/:icao24/route-context",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/trajectories/:id",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodPost,
			path:         "/api/v1/trajectories/:id/route-intelligence",
			handlerCount: 2,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/trajectories/:id/route-intelligence/latest",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/trajectories/:id/route-intelligence/history",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/flights/:flightID/states",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/aircraft/:icao24/latest-state",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/flights",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/flights/:id",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/aircraft",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/aircraft/:icao24",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/airports",
			handlerCount: 1,
		},
		{
			method:       fiber.MethodGet,
			path:         "/api/v1/airports/:icao",
			handlerCount: 1,
		},
	}

	actual := make(
		map[string]fiber.Route,
		len(expectations),
	)
	for _, route := range app.GetRoutes(true) {
		if route.Method != fiber.MethodGet &&
			route.Method != fiber.MethodPost {
			continue
		}
		key := routeTopologyKey(
			route.Method,
			route.Path,
		)
		if _, exists := actual[key]; exists {
			t.Fatalf(
				"duplicate registered route: %s",
				key,
			)
		}
		actual[key] = route
	}

	if len(actual) != len(expectations) {
		t.Fatalf(
			"registered route count = %d, want %d\nactual=%v",
			len(actual),
			len(expectations),
			sortedRouteKeys(actual),
		)
	}

	for _, expectation := range expectations {
		key := routeTopologyKey(
			expectation.method,
			expectation.path,
		)
		route, exists := actual[key]
		if !exists {
			t.Fatalf(
				"missing route: %s",
				key,
			)
		}
		if len(route.Handlers) !=
			expectation.handlerCount {
			t.Fatalf(
				"handler count for %s = %d, want %d",
				key,
				len(route.Handlers),
				expectation.handlerCount,
			)
		}
	}

	postKey := routeTopologyKey(
		fiber.MethodPost,
		"/api/v1/trajectories/:id/route-intelligence",
	)
	postRoute := actual[postKey]
	if reflect.ValueOf(
		postRoute.Handlers[0],
	).Pointer() != reflect.ValueOf(
		mutationAuthorization,
	).Pointer() {
		t.Fatal(
			"mutation authorization is not the first POST route handler",
		)
	}
}

func routeTopologyKey(
	method string,
	path string,
) string {
	return fmt.Sprintf(
		"%s %s",
		method,
		path,
	)
}

func sortedRouteKeys(
	routes map[string]fiber.Route,
) []string {
	keys := make(
		[]string,
		0,
		len(routes),
	)
	for key := range routes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
