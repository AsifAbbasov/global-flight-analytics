package postgres

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
)

func TestNormalizeCurrentWeatherSnapshotAcceptsUnavailableMetrics(t *testing.T) {
	snapshot := weather.CurrentSnapshot{
		Provider:                weather.ProviderOpenMeteo,
		Latitude:                40.4093,
		Longitude:               49.8671,
		ObservedAt:              time.Now().UTC(),
		RetrievedAt:             time.Now().UTC(),
		MetricAvailabilityKnown: true,
	}

	normalized, err := normalizeCurrentWeatherSnapshot(snapshot)
	if err != nil {
		t.Fatalf("normalize unavailable weather metrics: %v", err)
	}

	availability := normalized.ResolvedMetricAvailability()
	if availability != (weather.CurrentMetricAvailability{}) {
		t.Fatalf("availability = %+v, want all unavailable", availability)
	}

	if nullableWeatherFloat64(0, false) != nil {
		t.Fatal("unavailable float weather metric must persist as NULL")
	}
	if nullableWeatherInt(0, false) != nil {
		t.Fatal("unavailable integer weather metric must persist as NULL")
	}
}

func TestLegacyWeatherSnapshotDefaultsMetricsToAvailable(t *testing.T) {
	availability := (weather.CurrentSnapshot{}).ResolvedMetricAvailability()
	if availability != weather.AllCurrentMetricsAvailable() {
		t.Fatalf("legacy availability = %+v, want all available", availability)
	}
}
