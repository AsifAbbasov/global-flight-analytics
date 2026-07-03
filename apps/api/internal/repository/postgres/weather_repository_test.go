package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
)

func TestNormalizeCurrentWeatherSnapshotDefaultsProviderAndRetrievedAt(t *testing.T) {
	observedAt := time.Date(2026, 7, 3, 6, 0, 0, 0, time.UTC)

	snapshot, err := normalizeCurrentWeatherSnapshot(weather.CurrentSnapshot{
		Latitude:                 40.4093,
		Longitude:                49.8671,
		ObservedAt:               observedAt,
		TemperatureCelsius:       28.4,
		RelativeHumidityPercent:  54,
		PrecipitationMillimeters: 0,
		RainMillimeters:          0,
		WeatherCode:              1,
		CloudCoverPercent:        22,
		SurfacePressureHPA:       1008.7,
		WindSpeedMetersPerSecond: 6.2,
		WindDirectionDegrees:     320,
		WindGustsMetersPerSecond: 9.5,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if snapshot.Provider != weather.ProviderOpenMeteo {
		t.Fatalf("expected provider %s, got %s", weather.ProviderOpenMeteo, snapshot.Provider)
	}

	if snapshot.RetrievedAt.IsZero() {
		t.Fatal("expected retrieved time to be defaulted")
	}

	if !snapshot.ObservedAt.Equal(observedAt) {
		t.Fatalf("expected observed at %s, got %s", observedAt, snapshot.ObservedAt)
	}
}

func TestNormalizeCurrentWeatherSnapshotTrimsProvider(t *testing.T) {
	snapshot, err := normalizeCurrentWeatherSnapshot(weather.CurrentSnapshot{
		Provider:                "  open_meteo  ",
		Latitude:                40.4093,
		Longitude:               49.8671,
		ObservedAt:              time.Now(),
		RelativeHumidityPercent: 50,
		CloudCoverPercent:       10,
		SurfacePressureHPA:      1000,
		WindDirectionDegrees:    90,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if snapshot.Provider != weather.ProviderOpenMeteo {
		t.Fatalf("expected provider %s, got %s", weather.ProviderOpenMeteo, snapshot.Provider)
	}
}

func TestNormalizeCurrentWeatherSnapshotRejectsInvalidCoordinates(t *testing.T) {
	_, err := normalizeCurrentWeatherSnapshot(weather.CurrentSnapshot{
		Provider:                weather.ProviderOpenMeteo,
		Latitude:                120,
		Longitude:               49.8671,
		ObservedAt:              time.Now(),
		RelativeHumidityPercent: 50,
		CloudCoverPercent:       10,
		SurfacePressureHPA:      1000,
		WindDirectionDegrees:    90,
	})
	if !errors.Is(err, ErrInvalidWeatherCoordinates) {
		t.Fatalf("expected ErrInvalidWeatherCoordinates, got %v", err)
	}
}

func TestNormalizeCurrentWeatherSnapshotRejectsZeroObservedAt(t *testing.T) {
	_, err := normalizeCurrentWeatherSnapshot(weather.CurrentSnapshot{
		Provider:                weather.ProviderOpenMeteo,
		Latitude:                40.4093,
		Longitude:               49.8671,
		RelativeHumidityPercent: 50,
		CloudCoverPercent:       10,
		SurfacePressureHPA:      1000,
		WindDirectionDegrees:    90,
	})
	if !errors.Is(err, ErrInvalidWeatherObservedAt) {
		t.Fatalf("expected ErrInvalidWeatherObservedAt, got %v", err)
	}
}

func TestNormalizeCurrentWeatherSnapshotRejectsInvalidHumidity(t *testing.T) {
	_, err := normalizeCurrentWeatherSnapshot(weather.CurrentSnapshot{
		Provider:                weather.ProviderOpenMeteo,
		Latitude:                40.4093,
		Longitude:               49.8671,
		ObservedAt:              time.Now(),
		RelativeHumidityPercent: 120,
		CloudCoverPercent:       10,
		SurfacePressureHPA:      1000,
		WindDirectionDegrees:    90,
	})
	if !errors.Is(err, ErrInvalidWeatherHumidity) {
		t.Fatalf("expected ErrInvalidWeatherHumidity, got %v", err)
	}
}

func TestNormalizeCurrentWeatherSnapshotRejectsInvalidCloudCover(t *testing.T) {
	_, err := normalizeCurrentWeatherSnapshot(weather.CurrentSnapshot{
		Provider:                weather.ProviderOpenMeteo,
		Latitude:                40.4093,
		Longitude:               49.8671,
		ObservedAt:              time.Now(),
		RelativeHumidityPercent: 50,
		CloudCoverPercent:       130,
		SurfacePressureHPA:      1000,
		WindDirectionDegrees:    90,
	})
	if !errors.Is(err, ErrInvalidWeatherCloudCover) {
		t.Fatalf("expected ErrInvalidWeatherCloudCover, got %v", err)
	}
}

func TestNormalizeCurrentWeatherSnapshotRejectsNegativePrecipitation(t *testing.T) {
	_, err := normalizeCurrentWeatherSnapshot(weather.CurrentSnapshot{
		Provider:                 weather.ProviderOpenMeteo,
		Latitude:                 40.4093,
		Longitude:                49.8671,
		ObservedAt:               time.Now(),
		RelativeHumidityPercent:  50,
		PrecipitationMillimeters: -1,
		CloudCoverPercent:        10,
		SurfacePressureHPA:       1000,
		WindDirectionDegrees:     90,
	})
	if !errors.Is(err, ErrInvalidWeatherPrecipitation) {
		t.Fatalf("expected ErrInvalidWeatherPrecipitation, got %v", err)
	}
}

func TestNormalizeCurrentWeatherSnapshotRejectsInvalidWind(t *testing.T) {
	_, err := normalizeCurrentWeatherSnapshot(weather.CurrentSnapshot{
		Provider:                 weather.ProviderOpenMeteo,
		Latitude:                 40.4093,
		Longitude:                49.8671,
		ObservedAt:               time.Now(),
		RelativeHumidityPercent:  50,
		CloudCoverPercent:        10,
		SurfacePressureHPA:       1000,
		WindSpeedMetersPerSecond: 5,
		WindDirectionDegrees:     400,
		WindGustsMetersPerSecond: 8,
	})
	if !errors.Is(err, ErrInvalidWeatherWind) {
		t.Fatalf("expected ErrInvalidWeatherWind, got %v", err)
	}
}

func TestSaveCurrentSnapshotRequiresPool(t *testing.T) {
	repository := NewWeatherRepository(nil)

	_, err := repository.SaveCurrentSnapshot(nil, weather.CurrentSnapshot{})
	if !errors.Is(err, ErrWeatherRepositoryPoolRequired) {
		t.Fatalf("expected ErrWeatherRepositoryPoolRequired, got %v", err)
	}
}
