package trackbuilder

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestBuildEmptyTrajectory(t *testing.T) {
	builder := NewBuilder(Config{})

	result := builder.Build(nil)

	if result.PointCount != 0 {
		t.Fatalf("expected zero points, got %d", result.PointCount)
	}

	if result.SegmentCount != 0 {
		t.Fatalf("expected zero segments, got %d", result.SegmentCount)
	}
}

func TestBuildContinuousTrajectory(t *testing.T) {
	builder := NewBuilder(Config{})
	observedAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	states := []flightstate.FlightState{
		makeFlightState("state-3", "ABC123", "AHY101", 40.4300, 49.8900, observedAt.Add(60*time.Second)),
		makeFlightState("state-1", "ABC123", "AHY101", 40.4100, 49.8700, observedAt),
		makeFlightState("state-2", "ABC123", "AHY101", 40.4200, 49.8800, observedAt.Add(30*time.Second)),
	}

	result := builder.Build(states)

	if result.ICAO24 != "ABC123" {
		t.Fatalf("expected ICAO24 ABC123, got %s", result.ICAO24)
	}

	if result.PointCount != 3 {
		t.Fatalf("expected 3 points, got %d", result.PointCount)
	}

	if result.SegmentCount != 1 {
		t.Fatalf("expected 1 segment, got %d", result.SegmentCount)
	}

	if result.CoverageGapCount != 0 {
		t.Fatalf("expected 0 coverage gaps, got %d", result.CoverageGapCount)
	}

	if !result.StartTime.Equal(observedAt) {
		t.Fatalf("expected start time %s, got %s", observedAt, result.StartTime)
	}

	if result.QualityScore <= 0 {
		t.Fatalf("expected positive quality score, got %f", result.QualityScore)
	}
}

func TestBuildTrajectoryWithCoverageGap(t *testing.T) {
	builder := NewBuilder(Config{})
	observedAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	states := []flightstate.FlightState{
		makeFlightState("state-1", "ABC123", "AHY101", 40.4100, 49.8700, observedAt),
		makeFlightState("state-2", "ABC123", "AHY101", 40.4200, 49.8800, observedAt.Add(30*time.Second)),
		makeFlightState("state-3", "ABC123", "AHY101", 40.4300, 49.8900, observedAt.Add(150*time.Second)),
	}

	result := builder.Build(states)

	if result.PointCount != 3 {
		t.Fatalf("expected 3 points, got %d", result.PointCount)
	}

	if result.SegmentCount != 2 {
		t.Fatalf("expected 2 segments, got %d", result.SegmentCount)
	}

	if result.CoverageGapCount != 1 {
		t.Fatalf("expected 1 coverage gap, got %d", result.CoverageGapCount)
	}

	if len(result.CoverageGaps) != 1 {
		t.Fatalf("expected 1 coverage gap object, got %d", len(result.CoverageGaps))
	}

	if result.CoverageGaps[0].Reason != trajectory.CoverageGapReasonTimeGap {
		t.Fatalf("expected time gap reason, got %s", result.CoverageGaps[0].Reason)
	}
}

func TestBuildManyGroupsStatesByAircraft(t *testing.T) {
	builder := NewBuilder(Config{})
	observedAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	states := []flightstate.FlightState{
		makeFlightState("state-1", "ABC123", "AHY101", 40.4100, 49.8700, observedAt),
		makeFlightState("state-2", "ABC123", "AHY101", 40.4200, 49.8800, observedAt.Add(30*time.Second)),
		makeFlightState("state-3", "DEF456", "THY202", 41.0000, 49.0000, observedAt),
	}

	result := builder.BuildMany(states)

	if len(result) != 2 {
		t.Fatalf("expected 2 trajectories, got %d", len(result))
	}

	if result["ABC123"].PointCount != 2 {
		t.Fatalf("expected ABC123 to have 2 points, got %d", result["ABC123"].PointCount)
	}

	if result["DEF456"].PointCount != 1 {
		t.Fatalf("expected DEF456 to have 1 point, got %d", result["DEF456"].PointCount)
	}
}

func makeFlightState(id string, icao24 string, callsign string, latitude float64, longitude float64, observedAt time.Time) flightstate.FlightState {
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
