package handlers

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceregionanalytics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/gofiber/fiber/v2"
)

type fakeAirspaceRegionAnalyticsReader struct {
	result  airspaceregionanalytics.Result
	request airspaceproduction.Request
	err     error
}

func (reader *fakeAirspaceRegionAnalyticsReader) GetAirspaceRegionAnalytics(
	_ context.Context,
	request airspaceproduction.Request,
) (airspaceregionanalytics.Result, error) {
	reader.request = request
	if reader.err != nil {
		return airspaceregionanalytics.Result{}, reader.err
	}
	return reader.result.Clone(), nil
}

func TestAirspaceRegionAnalyticsHandlerRejectsInvalidWindow(
	t *testing.T,
) {
	app := fiber.New()
	handler := NewAirspaceRegionAnalyticsHandler(
		&fakeAirspaceRegionAnalyticsReader{},
	)
	app.Get(
		"/regions/:code/analytics",
		handler.GetByRegionCode,
	)

	responseValue, err := app.Test(
		httptest.NewRequest(
			"GET",
			"/regions/azerbaijan/analytics?as_of_time=2026-07-17T12:00:00Z&window_seconds=61",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if responseValue.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d", responseValue.StatusCode)
	}
}

func TestAirspaceRegionAnalyticsHandlerMapsRegionNotFound(
	t *testing.T,
) {
	app := fiber.New()
	handler := NewAirspaceRegionAnalyticsHandler(
		&fakeAirspaceRegionAnalyticsReader{
			err: fmt.Errorf("wrapped: %w", region.ErrRegionNotFound),
		},
	)
	app.Get(
		"/regions/:code/analytics",
		handler.GetByRegionCode,
	)

	responseValue, err := app.Test(
		httptest.NewRequest(
			"GET",
			"/regions/missing/analytics?as_of_time=2026-07-17T12:00:00Z",
			nil,
		),
	)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if responseValue.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status = %d", responseValue.StatusCode)
	}
}

func TestParseAirspaceRegionAnalyticsRequestUsesDefaultWindow(
	t *testing.T,
) {
	request, err := parseAirspaceRegionAnalyticsRequest(
		"Azerbaijan",
		"2026-07-17T12:00:00Z",
		"",
	)
	if err != nil {
		t.Fatalf("parse error = %v", err)
	}
	if request.RegionCode != "azerbaijan" {
		t.Fatalf("RegionCode = %q", request.RegionCode)
	}
	if request.Window != 5*time.Minute {
		t.Fatalf("Window = %s", request.Window)
	}
}
