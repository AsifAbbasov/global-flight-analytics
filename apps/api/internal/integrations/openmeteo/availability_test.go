package openmeteo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetCurrentWeatherPreservesMissingMetricsAndTrueZeroes(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write(
					[]byte(`{
						"latitude": 40.4093,
						"longitude": 49.8671,
						"current": {
							"time": "2026-07-24T00:00",
							"temperature_2m": 0,
							"relative_humidity_2m": null,
							"precipitation": null,
							"rain": 0,
							"weather_code": null,
							"cloud_cover": 0,
							"surface_pressure": null,
							"wind_speed_10m": 0,
							"wind_direction_10m": null,
							"wind_gusts_10m": null
						}
					}`),
				)
			},
		),
	)
	defer server.Close()

	client, err := New(
		Config{
			BaseURL: server.URL,
			Timeout: time.Second,
		},
	)
	if err != nil {
		t.Fatalf("create Open-Meteo client: %v", err)
	}

	snapshot, err := client.GetCurrentWeather(
		context.Background(),
		CurrentWeatherRequest{
			Latitude:  40.4093,
			Longitude: 49.8671,
		},
	)
	if err != nil {
		t.Fatalf("get current weather: %v", err)
	}

	if !snapshot.MetricAvailabilityKnown {
		t.Fatal("expected explicit metric availability")
	}

	availability := snapshot.ResolvedMetricAvailability()
	if !availability.TemperatureCelsius ||
		!availability.RainMillimeters ||
		!availability.CloudCoverPercent ||
		!availability.WindSpeedMetersPerSecond {
		t.Fatalf("true zero metric availability was lost: %+v", availability)
	}

	if availability.RelativeHumidityPercent ||
		availability.PrecipitationMillimeters ||
		availability.WeatherCode ||
		availability.SurfacePressureHPA ||
		availability.WindDirectionDegrees ||
		availability.WindGustsMetersPerSecond {
		t.Fatalf("missing metrics were marked available: %+v", availability)
	}

	if snapshot.TemperatureCelsius != 0 ||
		snapshot.RainMillimeters != 0 ||
		snapshot.CloudCoverPercent != 0 ||
		snapshot.WindSpeedMetersPerSecond != 0 {
		t.Fatalf("true zero values changed: %+v", snapshot)
	}
}
