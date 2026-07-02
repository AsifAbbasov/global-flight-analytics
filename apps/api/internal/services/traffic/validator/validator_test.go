package validator

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestEvaluateFlightStateValidComplete(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	item := makeValidFlightState(now)

	result := EvaluateFlightState(item, now)

	if result.ValidationStatus != dataquality.ValidationStatusValid {
		t.Fatalf("expected valid status, got %s", result.ValidationStatus)
	}

	if result.Completeness != dataquality.CompletenessLevelComplete {
		t.Fatalf("expected complete completeness, got %s", result.Completeness)
	}

	if result.Confidence != dataquality.ConfidenceLevelHigh {
		t.Fatalf("expected high confidence, got %s", result.Confidence)
	}

	if result.Score < 0.99 {
		t.Fatalf("expected score close to 1, got %f", result.Score)
	}

	if len(result.MissingFields) != 0 {
		t.Fatalf("expected no missing fields, got %v", result.MissingFields)
	}

	if len(result.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", result.Warnings)
	}
}

func TestEvaluateFlightStateInvalidICAO24(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	item := makeValidFlightState(now)
	item.ICAO24 = "INVALID"

	result := EvaluateFlightState(item, now)

	if result.ValidationStatus != dataquality.ValidationStatusInvalid {
		t.Fatalf("expected invalid status, got %s", result.ValidationStatus)
	}

	if result.Completeness != dataquality.CompletenessLevelInsufficient {
		t.Fatalf("expected insufficient completeness, got %s", result.Completeness)
	}

	if result.Confidence != dataquality.ConfidenceLevelNone {
		t.Fatalf("expected no confidence, got %s", result.Confidence)
	}

	if IsValidFlightState(item, now) {
		t.Fatal("expected invalid flight state")
	}
}

func TestEvaluateFlightStateFutureObservedAt(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	item := makeValidFlightState(now)
	item.ObservedAt = now.Add(1 * time.Minute)

	result := EvaluateFlightState(item, now)

	if result.ValidationStatus != dataquality.ValidationStatusInvalid {
		t.Fatalf("expected invalid status, got %s", result.ValidationStatus)
	}

	if result.Confidence != dataquality.ConfidenceLevelNone {
		t.Fatalf("expected no confidence, got %s", result.Confidence)
	}
}

func TestEvaluateFlightStatePositionOnlyPartial(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	item := makeValidFlightState(now)
	item.VelocityMPS = -1
	item.HeadingDegrees = 720

	result := EvaluateFlightState(item, now)

	if result.ValidationStatus != dataquality.ValidationStatusPartial {
		t.Fatalf("expected partial status, got %s", result.ValidationStatus)
	}

	if result.Completeness != dataquality.CompletenessLevelPositionOnly {
		t.Fatalf("expected position only completeness, got %s", result.Completeness)
	}

	if result.Confidence == dataquality.ConfidenceLevelNone {
		t.Fatalf("expected some confidence, got %s", result.Confidence)
	}

	if !IsValidFlightState(item, now) {
		t.Fatal("expected position-only flight state to remain usable")
	}
}

func TestFilterValidFlightStatesKeepsPartialAndDropsInvalid(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	validItem := makeValidFlightState(now)

	partialItem := makeValidFlightState(now)
	partialItem.ICAO24 = "DEF456"
	partialItem.VelocityMPS = -1

	invalidItem := makeValidFlightState(now)
	invalidItem.ICAO24 = "BAD"

	result := FilterValidFlightStates(
		[]flightstate.FlightState{
			validItem,
			partialItem,
			invalidItem,
		},
		now,
	)

	if len(result) != 2 {
		t.Fatalf("expected 2 usable flight states, got %d", len(result))
	}
}

func TestEvaluateFlightStateRejectsInvalidCoordinates(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	item := makeValidFlightState(now)
	item.Latitude = math.NaN()

	result := EvaluateFlightState(item, now)

	if result.ValidationStatus != dataquality.ValidationStatusInvalid {
		t.Fatalf("expected invalid status, got %s", result.ValidationStatus)
	}

	if result.Completeness != dataquality.CompletenessLevelInsufficient {
		t.Fatalf("expected insufficient completeness, got %s", result.Completeness)
	}
}

func makeValidFlightState(now time.Time) flightstate.FlightState {
	return flightstate.FlightState{
		ID:                  "state-1",
		FlightID:            "flight-1",
		AircraftID:          "aircraft-1",
		ICAO24:              "ABC123",
		Callsign:            "AHY101",
		Latitude:            40.4093,
		Longitude:           49.8671,
		BarometricAltitudeM: 10000,
		GeometricAltitudeM:  10050,
		VelocityMPS:         230,
		HeadingDegrees:      90,
		VerticalRateMPS:     0,
		OnGround:            false,
		OriginCountry:       "Azerbaijan",
		ObservedAt:          now.Add(-30 * time.Second),
		SourceName:          "test",
	}
}
