package openmeteo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetCurrentWeather(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()

		if query.Get("latitude") != "40.409300" {
			t.Fatalf("expected latitude 40.409300, got %s", query.Get("latitude"))
		}

		if query.Get("longitude") != "49.867100" {
			t.Fatalf("expected longitude 49.867100, got %s", query.Get("longitude"))
		}

		if query.Get("wind_speed_unit") != "ms" {
			t.Fatalf("expected wind_speed_unit ms, got %s", query.Get("wind_speed_unit"))
		}

		if query.Get("timezone") != "UTC" {
			t.Fatalf("expected timezone UTC, got %s", query.Get("timezone"))
		}

		currentVariables := query.Get("current")
		for _, variable := range currentWeatherVariables() {
			if !strings.Contains(currentVariables, variable) {
				t.Fatalf("expected current variables to contain %s, got %s", variable, currentVariables)
			}
		}

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{
			"latitude": 40.4093,
			"longitude": 49.8671,
			"current": {
				"time": "2026-07-03T06:00",
				"temperature_2m": 28.4,
				"relative_humidity_2m": 54,
				"precipitation": 0.0,
				"rain": 0.0,
				"weather_code": 1,
				"cloud_cover": 22,
				"surface_pressure": 1008.7,
				"wind_speed_10m": 6.2,
				"wind_direction_10m": 320,
				"wind_gusts_10m": 9.5
			}
		}`))
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snapshot, err := client.GetCurrentWeather(context.Background(), CurrentWeatherRequest{
		Latitude:  40.4093,
		Longitude: 49.8671,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if snapshot.Provider != "open_meteo" {
		t.Fatalf("expected provider open_meteo, got %s", snapshot.Provider)
	}

	if snapshot.TemperatureCelsius != 28.4 {
		t.Fatalf("expected temperature 28.4, got %f", snapshot.TemperatureCelsius)
	}

	if snapshot.WindSpeedMetersPerSecond != 6.2 {
		t.Fatalf("expected wind speed 6.2, got %f", snapshot.WindSpeedMetersPerSecond)
	}

	if snapshot.WindDirectionDegrees != 320 {
		t.Fatalf("expected wind direction 320, got %d", snapshot.WindDirectionDegrees)
	}

	if snapshot.ObservedAt.IsZero() {
		t.Fatal("expected observed time to be parsed")
	}
}

func TestGetCurrentWeatherRejectsInvalidCoordinates(t *testing.T) {
	client, err := New(Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = client.GetCurrentWeather(context.Background(), CurrentWeatherRequest{
		Latitude:  120,
		Longitude: 49.8671,
	})
	if !errors.Is(err, ErrInvalidCoordinates) {
		t.Fatalf("expected ErrInvalidCoordinates, got %v", err)
	}
}

func TestGetCurrentWeatherHandlesUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client, err := New(Config{
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = client.GetCurrentWeather(context.Background(), CurrentWeatherRequest{
		Latitude:  40.4093,
		Longitude: 49.8671,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewRejectsInvalidBaseURL(t *testing.T) {
	_, err := New(Config{
		BaseURL: "://bad-url",
	})
	if !errors.Is(err, ErrInvalidBaseURL) {
		t.Fatalf("expected ErrInvalidBaseURL, got %v", err)
	}
}
