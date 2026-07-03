package processor

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestProcessEmptyInput(t *testing.T) {
	now := fixedTime()
	processor := New(Config{
		Now: func() time.Time {
			return now
		},
	})

	result := processor.Process(nil)

	if result.Stats.ReceivedCount != 0 {
		t.Fatalf("expected 0 received states, got %d", result.Stats.ReceivedCount)
	}

	if result.Stats.UsableCount != 0 {
		t.Fatalf("expected 0 usable states, got %d", result.Stats.UsableCount)
	}

	if result.Stats.InvalidCount != 0 {
		t.Fatalf("expected 0 invalid states, got %d", result.Stats.InvalidCount)
	}

	if result.Stats.TrajectoryCount != 0 {
		t.Fatalf("expected 0 trajectories, got %d", result.Stats.TrajectoryCount)
	}
}

func TestProcessValidPartialAndInvalidStates(t *testing.T) {
	now := fixedTime()
	processor := New(Config{
		Now: func() time.Time {
			return now
		},
	})

	validState := makeProcessorFlightState("state-1", "ABC123", "AHY101", 40.4100, 49.8700, now.Add(-60*time.Second))

	partialState := makeProcessorFlightState("state-2", "DEF456", "THY202", 40.4200, 49.8800, now.Add(-30*time.Second))
	partialState.VelocityMPS = -1

	invalidState := makeProcessorFlightState("state-3", "BAD", "BAD101", 40.4300, 49.8900, now.Add(-30*time.Second))

	result := processor.Process([]flightstate.FlightState{
		validState,
		partialState,
		invalidState,
	})

	if result.Stats.ReceivedCount != 3 {
		t.Fatalf("expected 3 received states, got %d", result.Stats.ReceivedCount)
	}

	if result.Stats.ValidCount != 1 {
		t.Fatalf("expected 1 valid state, got %d", result.Stats.ValidCount)
	}

	if result.Stats.PartialCount != 1 {
		t.Fatalf("expected 1 partial state, got %d", result.Stats.PartialCount)
	}

	if result.Stats.InvalidCount != 1 {
		t.Fatalf("expected 1 invalid state, got %d", result.Stats.InvalidCount)
	}

	if result.Stats.UsableCount != 2 {
		t.Fatalf("expected 2 usable states, got %d", result.Stats.UsableCount)
	}

	if len(result.UsableStates) != 2 {
		t.Fatalf("expected 2 usable state objects, got %d", len(result.UsableStates))
	}

	if len(result.InvalidStates) != 1 {
		t.Fatalf("expected 1 invalid state object, got %d", len(result.InvalidStates))
	}

	if result.InvalidStates[0].Quality.ValidationStatus != dataquality.ValidationStatusInvalid {
		t.Fatalf("expected invalid quality status, got %s", result.InvalidStates[0].Quality.ValidationStatus)
	}
}

func TestProcessNormalizesStatesBeforeValidationAndTrajectoryBuilding(t *testing.T) {
	now := fixedTime()
	processor := New(Config{
		Now: func() time.Time {
			return now
		},
	})

	state := makeProcessorFlightState(
		"state-1",
		" abc123 ",
		" ahy101 ",
		40.4100,
		49.8700,
		now.Add(-60*time.Second),
	)
	state.SourceName = " TEST "

	result := processor.Process([]flightstate.FlightState{
		state,
	})

	if result.Stats.ReceivedCount != 1 {
		t.Fatalf("expected 1 received state, got %d", result.Stats.ReceivedCount)
	}

	if result.Stats.UsableCount != 1 {
		t.Fatalf("expected 1 usable state, got %d", result.Stats.UsableCount)
	}

	if len(result.UsableStates) != 1 {
		t.Fatalf("expected 1 usable state object, got %d", len(result.UsableStates))
	}

	normalizedState := result.UsableStates[0].State

	if normalizedState.ICAO24 != "ABC123" {
		t.Fatalf("expected normalized ICAO24 ABC123, got %q", normalizedState.ICAO24)
	}

	if normalizedState.Callsign != "AHY101" {
		t.Fatalf("expected normalized callsign AHY101, got %q", normalizedState.Callsign)
	}

	if normalizedState.SourceName != "test" {
		t.Fatalf("expected normalized source name test, got %q", normalizedState.SourceName)
	}

	trajectoryItem, exists := result.Trajectories["ABC123"]
	if !exists {
		t.Fatal("expected trajectory under normalized ICAO24 key ABC123")
	}

	if trajectoryItem.ICAO24 != "ABC123" {
		t.Fatalf("expected normalized trajectory ICAO24 ABC123, got %q", trajectoryItem.ICAO24)
	}

	if trajectoryItem.Callsign != "AHY101" {
		t.Fatalf("expected normalized trajectory callsign AHY101, got %q", trajectoryItem.Callsign)
	}

	if trajectoryItem.SourceName != "test" {
		t.Fatalf("expected normalized trajectory source name test, got %q", trajectoryItem.SourceName)
	}
}

func TestProcessBuildsTrajectoriesByAircraft(t *testing.T) {
	now := fixedTime()
	processor := New(Config{
		Now: func() time.Time {
			return now
		},
	})

	states := []flightstate.FlightState{
		makeProcessorFlightState("state-1", "ABC123", "AHY101", 40.4100, 49.8700, now.Add(-90*time.Second)),
		makeProcessorFlightState("state-2", "ABC123", "AHY101", 40.4200, 49.8800, now.Add(-60*time.Second)),
		makeProcessorFlightState("state-3", "DEF456", "THY202", 41.0000, 49.0000, now.Add(-60*time.Second)),
	}

	result := processor.Process(states)

	if result.Stats.TrajectoryCount != 2 {
		t.Fatalf("expected 2 trajectories, got %d", result.Stats.TrajectoryCount)
	}

	if result.Trajectories["ABC123"].PointCount != 2 {
		t.Fatalf("expected ABC123 trajectory to have 2 points, got %d", result.Trajectories["ABC123"].PointCount)
	}

	if result.Trajectories["DEF456"].PointCount != 1 {
		t.Fatalf("expected DEF456 trajectory to have 1 point, got %d", result.Trajectories["DEF456"].PointCount)
	}
}

func TestProcessCountsCoverageGaps(t *testing.T) {
	now := fixedTime()
	processor := New(Config{
		Now: func() time.Time {
			return now
		},
	})

	states := []flightstate.FlightState{
		makeProcessorFlightState("state-1", "ABC123", "AHY101", 40.4100, 49.8700, now.Add(-240*time.Second)),
		makeProcessorFlightState("state-2", "ABC123", "AHY101", 40.4200, 49.8800, now.Add(-210*time.Second)),
		makeProcessorFlightState("state-3", "ABC123", "AHY101", 40.4300, 49.8900, now.Add(-60*time.Second)),
	}

	result := processor.Process(states)

	if result.Stats.TrajectoryCount != 1 {
		t.Fatalf("expected 1 trajectory, got %d", result.Stats.TrajectoryCount)
	}

	if result.Stats.CoverageGapCount != 1 {
		t.Fatalf("expected 1 coverage gap, got %d", result.Stats.CoverageGapCount)
	}

	if result.Trajectories["ABC123"].CoverageGapCount != 1 {
		t.Fatalf("expected ABC123 trajectory to have 1 coverage gap, got %d", result.Trajectories["ABC123"].CoverageGapCount)
	}
}

func fixedTime() time.Time {
	return time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
}

func makeProcessorFlightState(id string, icao24 string, callsign string, latitude float64, longitude float64, observedAt time.Time) flightstate.FlightState {
	return flightstate.FlightState{
		ID:                  id,
		FlightID:            "flight-" + icao24,
		AircraftID:          "aircraft-" + icao24,
		ICAO24:              icao24,
		Callsign:            callsign,
		Latitude:            latitude,
		Longitude:           longitude,
		BarometricAltitudeM: 10000,
		GeometricAltitudeM:  10000,
		VelocityMPS:         230,
		HeadingDegrees:      90,
		VerticalRateMPS:     0,
		OnGround:            false,
		OriginCountry:       "Azerbaijan",
		ObservedAt:          observedAt,
		SourceName:          "test",
	}
}
