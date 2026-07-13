package metrics

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
)

func TestTrafficDensityMetric(t *testing.T) {
	metric := TrafficDensityMetric{}

	value, err := metric.Calculate(snapshot.Snapshot{
		ActiveAircraft:       250,
		AreaSquareKilometers: 5000,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := 0.05

	if value != expected {
		t.Fatalf("expected %v, got %v", expected, value)
	}
}

func TestTrafficDensityMetricRejectsZeroArea(t *testing.T) {
	metric := TrafficDensityMetric{}

	_, err := metric.Calculate(snapshot.Snapshot{})

	if err == nil {
		t.Fatal("expected error")
	}
}
