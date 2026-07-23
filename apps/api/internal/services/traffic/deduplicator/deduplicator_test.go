package deduplicator

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestRemoveExactDuplicatesRemovesRepeatedMovementPoint(t *testing.T) {
	observedAt := time.Date(2026, time.July, 4, 10, 0, 0, 0, time.UTC)

	first := flightstate.FlightState{
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
		VerticalRateMPS:     2.5,
		OnGround:            false,
		OriginCountry:       "Azerbaijan",
		ObservedAt:          observedAt,
		SourceName:          "airplanes.live",
	}

	duplicate := first
	duplicate.ID = "state-2"
	duplicate.FlightID = "flight-2"
	duplicate.AircraftID = "aircraft-2"
	duplicate.IngestionRunID = "run-2"

	result := RemoveExactDuplicates([]flightstate.FlightState{
		first,
		duplicate,
	})

	if result.DuplicateCount != 1 {
		t.Fatalf("expected DuplicateCount to be 1, got %d", result.DuplicateCount)
	}

	if len(result.UniqueStates) != 1 {
		t.Fatalf("expected 1 unique state, got %d", len(result.UniqueStates))
	}

	if result.UniqueStates[0].ID != first.ID {
		t.Fatalf(
			"expected first observation to be preserved, got state ID %q",
			result.UniqueStates[0].ID,
		)
	}
}

func TestRemoveExactDuplicatesKeepsDistinctMovementPoint(t *testing.T) {
	observedAt := time.Date(2026, time.July, 4, 10, 0, 0, 0, time.UTC)

	first := flightstate.FlightState{
		ICAO24:              "ABC123",
		Latitude:            40.4093,
		Longitude:           49.8671,
		BarometricAltitudeM: 10000,
		VelocityMPS:         230,
		HeadingDegrees:      90,
		ObservedAt:          observedAt,
	}

	second := first
	second.HeadingDegrees = 91

	result := RemoveExactDuplicates([]flightstate.FlightState{
		first,
		second,
	})

	if result.DuplicateCount != 0 {
		t.Fatalf("expected DuplicateCount to be 0, got %d", result.DuplicateCount)
	}

	if len(result.UniqueStates) != 2 {
		t.Fatalf("expected 2 unique states, got %d", len(result.UniqueStates))
	}
}

func TestRemoveExactDuplicatesTreatsSameInstantAsSameObservedAt(t *testing.T) {
	utcTime := time.Date(2026, time.July, 4, 10, 0, 0, 0, time.UTC)
	bakuLocation := time.FixedZone("UTC+4", 4*60*60)
	bakuTime := utcTime.In(bakuLocation)

	first := flightstate.FlightState{
		ICAO24:              "ABC123",
		Latitude:            40.4093,
		Longitude:           49.8671,
		BarometricAltitudeM: 10000,
		VelocityMPS:         230,
		HeadingDegrees:      90,
		ObservedAt:          utcTime,
	}

	second := first
	second.ObservedAt = bakuTime

	result := RemoveExactDuplicates([]flightstate.FlightState{
		first,
		second,
	})

	if result.DuplicateCount != 1 {
		t.Fatalf("expected DuplicateCount to be 1, got %d", result.DuplicateCount)
	}

	if len(result.UniqueStates) != 1 {
		t.Fatalf("expected 1 unique state, got %d", len(result.UniqueStates))
	}
}

func TestRemoveExactDuplicatesPreservesUniqueStateOrder(t *testing.T) {
	baseTime := time.Date(2026, time.July, 4, 10, 0, 0, 0, time.UTC)

	first := flightstate.FlightState{
		ID:                  "state-1",
		ICAO24:              "ABC123",
		Latitude:            40.4093,
		Longitude:           49.8671,
		BarometricAltitudeM: 10000,
		VelocityMPS:         230,
		HeadingDegrees:      90,
		ObservedAt:          baseTime,
	}

	duplicateOfFirst := first
	duplicateOfFirst.ID = "state-duplicate"

	second := flightstate.FlightState{
		ID:                  "state-2",
		ICAO24:              "ABC123",
		Latitude:            40.4193,
		Longitude:           49.8771,
		BarometricAltitudeM: 10100,
		VelocityMPS:         232,
		HeadingDegrees:      92,
		ObservedAt:          baseTime.Add(time.Second),
	}

	third := flightstate.FlightState{
		ID:                  "state-3",
		ICAO24:              "ABC123",
		Latitude:            40.4293,
		Longitude:           49.8871,
		BarometricAltitudeM: 10200,
		VelocityMPS:         234,
		HeadingDegrees:      94,
		ObservedAt:          baseTime.Add(2 * time.Second),
	}

	result := RemoveExactDuplicates([]flightstate.FlightState{
		first,
		duplicateOfFirst,
		second,
		third,
	})

	if result.DuplicateCount != 1 {
		t.Fatalf("expected DuplicateCount to be 1, got %d", result.DuplicateCount)
	}

	if len(result.UniqueStates) != 3 {
		t.Fatalf("expected 3 unique states, got %d", len(result.UniqueStates))
	}

	expectedIDs := []string{
		"state-1",
		"state-2",
		"state-3",
	}

	for index, expectedID := range expectedIDs {
		actualID := result.UniqueStates[index].ID

		if actualID != expectedID {
			t.Fatalf(
				"expected state at index %d to have ID %q, got %q",
				index,
				expectedID,
				actualID,
			)
		}
	}
}
