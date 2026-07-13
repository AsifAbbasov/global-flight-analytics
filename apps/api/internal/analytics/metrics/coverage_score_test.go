package metrics

import (
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
)

func TestCoverageScoreMetric(t *testing.T) {
	metric := CoverageScoreMetric{}

	value, err := metric.Calculate(snapshot.Snapshot{
		ObservedSamples: 75,
		ExpectedSamples: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	const expected = 0.75
	if math.Abs(value-expected) > 0.000001 {
		t.Fatalf("expected %v, got %v", expected, value)
	}
}

func TestCoverageScoreMetricReturnsZeroWithoutObservedSamples(t *testing.T) {
	metric := CoverageScoreMetric{}

	value, err := metric.Calculate(snapshot.Snapshot{
		ExpectedSamples: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != 0 {
		t.Fatalf("expected zero, got %v", value)
	}
}

func TestCoverageScoreMetricCapsResultAtOne(t *testing.T) {
	metric := CoverageScoreMetric{}

	value, err := metric.Calculate(snapshot.Snapshot{
		ObservedSamples: 125,
		ExpectedSamples: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != 1 {
		t.Fatalf("expected one, got %v", value)
	}
}

func TestCoverageScoreMetricRejectsMissingExpectedSamples(t *testing.T) {
	metric := CoverageScoreMetric{}

	_, err := metric.Calculate(snapshot.Snapshot{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCoverageScoreMetricRejectsNegativeObservedSamples(t *testing.T) {
	metric := CoverageScoreMetric{}

	_, err := metric.Calculate(snapshot.Snapshot{
		ObservedSamples: -1,
		ExpectedSamples: 100,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
