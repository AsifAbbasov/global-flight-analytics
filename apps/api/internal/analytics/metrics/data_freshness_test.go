package metrics

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
)

func TestDataFreshnessMetric(t *testing.T) {
	observedAt := time.Date(2026, time.July, 13, 10, 0, 0, 0, time.UTC)
	metric := DataFreshnessMetric{MaxAge: 2 * time.Minute}

	value, err := metric.Calculate(
		snapshot.Snapshot{Time: observedAt},
		observedAt.Add(30*time.Second),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const expected = 0.75
	if math.Abs(value-expected) > 0.000001 {
		t.Fatalf("expected %v, got %v", expected, value)
	}
}

func TestDataFreshnessMetricReturnsZeroForStaleSnapshot(t *testing.T) {
	observedAt := time.Date(2026, time.July, 13, 10, 0, 0, 0, time.UTC)
	metric := DataFreshnessMetric{MaxAge: 2 * time.Minute}

	value, err := metric.Calculate(
		snapshot.Snapshot{Time: observedAt},
		observedAt.Add(3*time.Minute),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != 0 {
		t.Fatalf("expected zero, got %v", value)
	}
}

func TestDataFreshnessMetricRejectsMissingSnapshotTime(t *testing.T) {
	metric := DataFreshnessMetric{MaxAge: 2 * time.Minute}

	_, err := metric.Calculate(snapshot.Snapshot{}, time.Now().UTC())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDataFreshnessMetricRejectsFutureSnapshot(t *testing.T) {
	observedAt := time.Date(2026, time.July, 13, 10, 0, 0, 0, time.UTC)
	metric := DataFreshnessMetric{MaxAge: 2 * time.Minute}

	_, err := metric.Calculate(
		snapshot.Snapshot{Time: observedAt},
		observedAt.Add(-time.Second),
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDataFreshnessMetricRejectsInvalidMaximumAge(t *testing.T) {
	observedAt := time.Date(2026, time.July, 13, 10, 0, 0, 0, time.UTC)
	metric := DataFreshnessMetric{}

	_, err := metric.Calculate(
		snapshot.Snapshot{Time: observedAt},
		observedAt,
	)
	if err == nil {
		t.Fatal("expected error")
	}
}
