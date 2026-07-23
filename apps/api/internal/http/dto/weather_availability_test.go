package dto

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
)

func TestToCurrentWeatherPreservesNullAndTrueZero(t *testing.T) {
	response := ToCurrentWeather(
		weatherservice.CurrentWeatherResult{
			Snapshot: weather.CurrentSnapshot{
				TemperatureCelsius:      0,
				MetricAvailabilityKnown: true,
				MetricAvailability: weather.CurrentMetricAvailability{
					TemperatureCelsius: true,
				},
			},
		},
	)

	if response.TemperatureCelsius == nil {
		t.Fatal("available zero temperature must remain available")
	}
	if *response.TemperatureCelsius != 0 {
		t.Fatalf("temperature = %v, want 0", *response.TemperatureCelsius)
	}
	if response.RelativeHumidityPercent != nil {
		t.Fatal("unavailable humidity must be serialized as null")
	}
}
