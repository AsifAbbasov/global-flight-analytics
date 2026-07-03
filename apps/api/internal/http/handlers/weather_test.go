package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
	"github.com/gofiber/fiber/v2"
)

func TestWeatherHandlerGetCurrent(t *testing.T) {
	service := &fakeWeatherHTTPService{
		result: makeWeatherHTTPResult(),
	}

	handler := NewWeatherHandler(service)

	app := fiber.New()
	app.Get("/api/v1/weather/current", handler.GetCurrent)

	response := performWeatherRequest(t, app, http.MethodGet, "/api/v1/weather/current?lat=40.4675&lon=50.0467")

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	if !service.called {
		t.Fatal("expected weather service to be called")
	}

	if service.lastRequest.Latitude != 40.4675 {
		t.Fatalf("expected latitude 40.4675, got %f", service.lastRequest.Latitude)
	}

	if service.lastRequest.Longitude != 50.0467 {
		t.Fatalf("expected longitude 50.0467, got %f", service.lastRequest.Longitude)
	}

	body := readWeatherResponseBody(t, response)

	if !strings.Contains(body, `"snapshot_id":"weather-snapshot-1"`) {
		t.Fatalf("expected snapshot id in response, got %s", body)
	}

	if !strings.Contains(body, `"provider":"open_meteo"`) {
		t.Fatalf("expected provider in response, got %s", body)
	}
}

func TestWeatherHandlerGetCurrentRejectsMissingLatitude(t *testing.T) {
	service := &fakeWeatherHTTPService{
		result: makeWeatherHTTPResult(),
	}

	handler := NewWeatherHandler(service)

	app := fiber.New()
	app.Get("/api/v1/weather/current", handler.GetCurrent)

	response := performWeatherRequest(t, app, http.MethodGet, "/api/v1/weather/current?lon=50.0467")

	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	if service.called {
		t.Fatal("expected weather service not to be called")
	}
}

func TestWeatherHandlerGetCurrentRejectsInvalidLatitude(t *testing.T) {
	service := &fakeWeatherHTTPService{
		result: makeWeatherHTTPResult(),
	}

	handler := NewWeatherHandler(service)

	app := fiber.New()
	app.Get("/api/v1/weather/current", handler.GetCurrent)

	response := performWeatherRequest(t, app, http.MethodGet, "/api/v1/weather/current?lat=abc&lon=50.0467")

	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	if service.called {
		t.Fatal("expected weather service not to be called")
	}
}

func TestWeatherHandlerGetCurrentReturnsServiceCoordinateError(t *testing.T) {
	service := &fakeWeatherHTTPService{
		err: weatherservice.ErrInvalidWeatherCoordinates,
	}

	handler := NewWeatherHandler(service)

	app := fiber.New()
	app.Get("/api/v1/weather/current", handler.GetCurrent)

	response := performWeatherRequest(t, app, http.MethodGet, "/api/v1/weather/current?lat=91&lon=50.0467")

	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
}

func TestWeatherHandlerGetCurrentReturnsServiceUnavailable(t *testing.T) {
	handler := NewWeatherHandler(nil)

	app := fiber.New()
	app.Get("/api/v1/weather/current", handler.GetCurrent)

	response := performWeatherRequest(t, app, http.MethodGet, "/api/v1/weather/current?lat=40.4675&lon=50.0467")

	if response.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", response.StatusCode)
	}
}

func TestWeatherHandlerGetCurrentReturnsInternalError(t *testing.T) {
	service := &fakeWeatherHTTPService{
		err: errors.New("weather service failed"),
	}

	handler := NewWeatherHandler(service)

	app := fiber.New()
	app.Get("/api/v1/weather/current", handler.GetCurrent)

	response := performWeatherRequest(t, app, http.MethodGet, "/api/v1/weather/current?lat=40.4675&lon=50.0467")

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.StatusCode)
	}

	body := readWeatherResponseBody(t, response)

	if !strings.Contains(body, `"code":"WEATHER_LOAD_FAILED"`) {
		t.Fatalf("expected weather load error code, got %s", body)
	}
}

func performWeatherRequest(t *testing.T, app *fiber.App, method string, path string) *http.Response {
	t.Helper()

	request := httptest.NewRequest(method, path, nil)

	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("expected no request error, got %v", err)
	}

	return response
}

func readWeatherResponseBody(t *testing.T, response *http.Response) string {
	t.Helper()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("expected readable response body, got %v", err)
	}

	return string(body)
}

func makeWeatherHTTPResult() weatherservice.CurrentWeatherResult {
	now := time.Date(2026, 7, 3, 8, 15, 0, 0, time.UTC)

	return weatherservice.CurrentWeatherResult{
		SnapshotID: "weather-snapshot-1",
		Snapshot: domainweather.CurrentSnapshot{
			Provider:                 domainweather.ProviderOpenMeteo,
			Latitude:                 40.4375,
			Longitude:                50.0625,
			ObservedAt:               now,
			TemperatureCelsius:       29.5,
			RelativeHumidityPercent:  55,
			PrecipitationMillimeters: 0,
			RainMillimeters:          0,
			WeatherCode:              0,
			CloudCoverPercent:        0,
			SurfacePressureHPA:       1010.3,
			WindSpeedMetersPerSecond: 5.36,
			WindDirectionDegrees:     194,
			WindGustsMetersPerSecond: 9.7,
			RetrievedAt:              now.Add(time.Second),
		},
		StoredAt: now.Add(2 * time.Second),
	}
}

type fakeWeatherHTTPService struct {
	called      bool
	lastRequest weatherservice.CurrentWeatherRequest
	result      weatherservice.CurrentWeatherResult
	err         error
}

func (service *fakeWeatherHTTPService) GetAndStoreCurrentWeather(
	ctx context.Context,
	request weatherservice.CurrentWeatherRequest,
) (weatherservice.CurrentWeatherResult, error) {
	service.called = true
	service.lastRequest = request

	if service.err != nil {
		return weatherservice.CurrentWeatherResult{}, service.err
	}

	return service.result, nil
}
