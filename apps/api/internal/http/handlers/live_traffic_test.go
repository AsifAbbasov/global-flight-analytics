package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	livetraffic "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/live"
	"github.com/gofiber/fiber/v2"
)

func TestLiveTrafficHandlerIncludesSelectedAircraftOutsideBounds(t *testing.T) {
	store, err := livetraffic.NewStore(livetraffic.DefaultConfig())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	store.UpsertBatch([]livetraffic.Aircraft{
		{
			ICAO24:     "abc123",
			Latitude:   40.4,
			Longitude:  49.8,
			ObservedAt: now.Add(-2 * time.Second),
			ReceivedAt: now.Add(-time.Second),
			Source:     "provider-a",
		},
		{
			ICAO24:     "def456",
			Latitude:   42,
			Longitude:  51,
			ObservedAt: now.Add(-3 * time.Second),
			ReceivedAt: now.Add(-2 * time.Second),
			Source:     "provider-b",
		},
	})

	app := fiber.New()
	handler := newLiveTrafficHandlerWithClock(store, func() time.Time { return now })
	app.Get("/traffic/live", handler.GetSnapshot)
	request := httptest.NewRequest(
		fiber.MethodGet,
		"/traffic/live?min_lat=40&min_lon=49&max_lat=41&max_lon=50&selected=DEF456&limit=10",
		nil,
	)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}

	var payload struct {
		Success bool                    `json:"success"`
		Data    dto.LiveTrafficSnapshot `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Success || len(payload.Data.Aircraft) != 2 || payload.Data.Aircraft[0].ICAO24 != "def456" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestLiveTrafficHandlerRejectsPartialBounds(t *testing.T) {
	store, err := livetraffic.NewStore(livetraffic.DefaultConfig())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	handler := NewLiveTrafficHandler(store)
	app := fiber.New()
	app.Get("/traffic/live", handler.GetSnapshot)

	response, err := app.Test(httptest.NewRequest(
		fiber.MethodGet,
		"/traffic/live?min_lat=40",
		nil,
	))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.StatusCode)
	}
}
