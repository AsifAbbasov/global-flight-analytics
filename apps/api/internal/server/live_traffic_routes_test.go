package server

import (
	"testing"

	livetraffic "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/live"
	"github.com/gofiber/fiber/v2"
)

func TestRegisterLiveTrafficRoutes(t *testing.T) {
	store, err := livetraffic.NewStore(livetraffic.DefaultConfig())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	app := fiber.New()
	v1 := app.Group("/api/v1")
	registerLiveTrafficRoutes(v1, store)

	found := false
	for _, route := range app.GetRoutes() {
		if route.Method == fiber.MethodGet && route.Path == "/api/v1/traffic/live" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("live traffic route was not registered")
	}
}

func TestNormalizeConfigCreatesDefaultLiveTrafficStore(t *testing.T) {
	config, err := normalizeConfig(Config{})
	if err != nil {
		t.Fatalf("normalize config: %v", err)
	}
	if config.LiveTrafficStore == nil {
		t.Fatal("default live traffic store is nil")
	}
}
