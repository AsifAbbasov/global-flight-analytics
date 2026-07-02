package gapdetector

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestHaversineDistanceKmSamePoint(t *testing.T) {
	distanceKm := HaversineDistanceKm(40.4093, 49.8671, 40.4093, 49.8671)

	if math.Abs(distanceKm) > 0.000001 {
		t.Fatalf("expected zero distance, got %f", distanceKm)
	}
}

func TestDetectWithoutGap(t *testing.T) {
	observedAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	previous := flightstate.FlightState{
		ICAO24:     "ABC123",
		Latitude:   40.4093,
		Longitude:  49.8671,
		ObservedAt: observedAt,
	}

	next := flightstate.FlightState{
		ICAO24:     "ABC123",
		Latitude:   40.4100,
		Longitude:  49.8680,
		ObservedAt: observedAt.Add(30 * time.Second),
	}

	result := Detect(previous, next, DefaultConfig())

	if result.HasGap {
		t.Fatalf("expected no coverage gap, got reason %s", result.Reason)
	}

	if result.EstimatedSpeedMPS <= 0 {
		t.Fatalf("expected positive estimated speed, got %f", result.EstimatedSpeedMPS)
	}
}

func TestDetectTimeGap(t *testing.T) {
	observedAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	previous := flightstate.FlightState{
		ICAO24:     "ABC123",
		Latitude:   40.4093,
		Longitude:  49.8671,
		ObservedAt: observedAt,
	}

	next := flightstate.FlightState{
		ICAO24:     "ABC123",
		Latitude:   40.4100,
		Longitude:  49.8680,
		ObservedAt: observedAt.Add(2 * time.Minute),
	}

	result := Detect(previous, next, DefaultConfig())

	if !result.HasGap {
		t.Fatal("expected coverage gap")
	}

	if result.Reason != trajectory.CoverageGapReasonTimeGap {
		t.Fatalf("expected time gap reason, got %s", result.Reason)
	}
}

func TestDetectMovementJump(t *testing.T) {
	observedAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	previous := flightstate.FlightState{
		ICAO24:     "ABC123",
		Latitude:   40.4093,
		Longitude:  49.8671,
		ObservedAt: observedAt,
	}

	next := flightstate.FlightState{
		ICAO24:     "ABC123",
		Latitude:   41.4093,
		Longitude:  50.8671,
		ObservedAt: observedAt.Add(30 * time.Second),
	}

	result := Detect(previous, next, DefaultConfig())

	if !result.HasGap {
		t.Fatal("expected coverage gap")
	}

	if result.Reason != trajectory.CoverageGapReasonMovementJump {
		t.Fatalf("expected movement jump reason, got %s", result.Reason)
	}
}
