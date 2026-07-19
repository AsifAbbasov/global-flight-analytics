package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
	"github.com/gofiber/fiber/v2"
)

type weatherRouteServiceStub struct {
	callCount int
	request   weatherservice.CurrentWeatherRequest
}

func (
	service *weatherRouteServiceStub,
) GetAndStoreCurrentWeather(
	_ context.Context,
	request weatherservice.CurrentWeatherRequest,
) (
	weatherservice.CurrentWeatherResult,
	error,
) {
	service.callCount++
	service.request = request

	now := time.Date(
		2026,
		time.July,
		20,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	return weatherservice.CurrentWeatherResult{
		SnapshotID: "weather-snapshot-test",
		Snapshot: domainweather.CurrentSnapshot{
			Provider:                 domainweather.ProviderOpenMeteo,
			Latitude:                 request.Latitude,
			Longitude:                request.Longitude,
			ObservedAt:               now,
			RetrievedAt:              now,
			TemperatureCelsius:       24,
			RelativeHumidityPercent:  50,
			PrecipitationMillimeters: 0,
			RainMillimeters:          0,
			WeatherCode:              1,
			CloudCoverPercent:        20,
			SurfacePressureHPA:       1013,
			WindSpeedMetersPerSecond: 4,
			WindDirectionDegrees:     180,
			WindGustsMetersPerSecond: 6,
		},
		StoredAt: now,
	}, nil
}

func TestRegisterCurrentWeatherRoutePreservesHTTPContract(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group("/api").Group("/v1")
	service := &weatherRouteServiceStub{}
	handler := handlers.NewWeatherHandler(
		service,
	)

	err := registerCurrentWeatherRoute(
		v1,
		handler,
	)
	if err != nil {
		t.Fatalf(
			"registerCurrentWeatherRoute() error = %v",
			err,
		)
	}

	response, err := app.Test(
		httptest.NewRequest(
			http.MethodGet,
			"/api/v1"+
				CurrentWeatherPath+
				"?lat=40.4&lon=49.8",
			nil,
		),
		-1,
	)
	if err != nil {
		t.Fatalf(
			"execute current weather request: %v",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			response.StatusCode,
			http.StatusOK,
		)
	}
	if service.callCount != 1 {
		t.Fatalf(
			"weather service calls = %d, want 1",
			service.callCount,
		)
	}
	if service.request.Latitude != 40.4 ||
		service.request.Longitude != 49.8 {
		t.Fatalf(
			"unexpected weather request: %#v",
			service.request,
		)
	}
}

func TestRegisterCurrentWeatherRouteRequiresRouterAndHandler(
	t *testing.T,
) {
	handler := handlers.NewWeatherHandler(
		&weatherRouteServiceStub{},
	)

	if err := registerCurrentWeatherRoute(
		nil,
		handler,
	); err == nil {
		t.Fatal(
			"nil weather router was accepted",
		)
	}

	app := fiber.New()
	v1 := app.Group("/api").Group("/v1")
	if err := registerCurrentWeatherRoute(
		v1,
		nil,
	); err == nil {
		t.Fatal(
			"nil weather handler was accepted",
		)
	}
}
