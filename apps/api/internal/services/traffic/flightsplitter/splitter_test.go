package flightsplitter

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestSplitSeparatesDifferentSourceFlightIdentifiers(t *testing.T) {
	observedAt := testTime()
	groups := Split([]Observation{
		observation("state-2", "ABC123", "22222222-2222-2222-2222-222222222222", "AHY102", false, observedAt.Add(time.Minute)),
		observation("state-1", "ABC123", "11111111-1111-1111-1111-111111111111", "AHY101", false, observedAt),
	})

	requireGroupCount(t, groups, 2)
	if groups[1].SplitReason != trajectory.FlightSplitReasonSourceFlightIDChanged {
		t.Fatalf("expected source flight identifier split, got %s", groups[1].SplitReason)
	}
	if groups[0].IdentityKey == groups[1].IdentityKey {
		t.Fatal("expected distinct identity keys")
	}
	if groups[0].IdentityBasis != trajectory.FlightIdentityBasisSourceFlightID {
		t.Fatalf("expected source identity basis, got %s", groups[0].IdentityBasis)
	}
}

func TestSplitIgnoresNonUUIDSourceFlightIdentifiers(t *testing.T) {
	observedAt := testTime()
	groups := Split([]Observation{
		observation("state-1", "ABC123", "flight-ABC123", "AHY101", false, observedAt),
		observation("state-2", "ABC123", "different-flight", "AHY101", false, observedAt.Add(time.Minute)),
	})

	requireGroupCount(t, groups, 1)
	if groups[0].IdentityBasis != trajectory.FlightIdentityBasisCallsignAndStartTime {
		t.Fatalf("expected callsign identity basis, got %s", groups[0].IdentityBasis)
	}
}

func TestSplitUsesCallsignChangeWhenSourceFlightIdentifierIsUnavailable(t *testing.T) {
	observedAt := testTime()
	groups := Split([]Observation{
		observation("state-1", "ABC123", "", "AHY101", false, observedAt),
		observation("state-2", "ABC123", "", "", false, observedAt.Add(time.Minute)),
		observation("state-3", "ABC123", "", "AHY102", false, observedAt.Add(2*time.Minute)),
	})

	requireGroupCount(t, groups, 2)
	if groups[1].SplitReason != trajectory.FlightSplitReasonCallsignChanged {
		t.Fatalf("expected callsign split, got %s", groups[1].SplitReason)
	}
}

func TestSplitUsesConfirmedGroundCycle(t *testing.T) {
	observedAt := testTime()
	groups := Split([]Observation{
		observation("state-1", "ABC123", "", "AHY101", false, observedAt),
		observation("state-2", "ABC123", "", "AHY101", true, observedAt.Add(time.Minute)),
		observation("state-3", "ABC123", "", "AHY101", true, observedAt.Add(2*time.Minute)),
		observation("state-4", "ABC123", "", "AHY101", false, observedAt.Add(3*time.Minute)),
	})

	requireGroupCount(t, groups, 2)
	if groups[1].SplitReason != trajectory.FlightSplitReasonGroundCycle {
		t.Fatalf("expected ground cycle split, got %s", groups[1].SplitReason)
	}
}

func TestSplitDoesNotTreatCoverageGapAsFlightBoundary(t *testing.T) {
	observedAt := testTime()
	groups := Split([]Observation{
		observation("state-1", "ABC123", "", "AHY101", false, observedAt),
		observation("state-2", "ABC123", "", "AHY101", false, observedAt.Add(12*time.Hour)),
	})

	requireGroupCount(t, groups, 1)
	if len(groups[0].Observations) != 2 {
		t.Fatalf("expected both observations in one flight, got %d", len(groups[0].Observations))
	}
}

func TestSplitIsDeterministicAcrossInputOrdering(t *testing.T) {
	observedAt := testTime()
	first := observation("state-1", "ABC123", "", "AHY101", false, observedAt)
	second := observation("state-2", "ABC123", "", "AHY101", false, observedAt.Add(time.Minute))

	ordered := Split([]Observation{first, second})
	reversed := Split([]Observation{second, first})

	requireGroupCount(t, ordered, 1)
	requireGroupCount(t, reversed, 1)
	if ordered[0].IdentityKey != reversed[0].IdentityKey {
		t.Fatalf("expected deterministic identity, got %s and %s", ordered[0].IdentityKey, reversed[0].IdentityKey)
	}
}

func TestSplitDoesNotUseCallsignWhenSourceFlightIdentifierIsActive(t *testing.T) {
	observedAt := testTime()
	groups := Split([]Observation{
		observation("state-1", "ABC123", "11111111-1111-1111-1111-111111111111", "AHY101", false, observedAt),
		observation("state-2", "ABC123", "11111111-1111-1111-1111-111111111111", "AHY999", false, observedAt.Add(time.Minute)),
	})

	requireGroupCount(t, groups, 1)
}

func requireGroupCount(t *testing.T, groups []Group, expected int) {
	t.Helper()
	if len(groups) != expected {
		t.Fatalf("expected %d groups, got %d", expected, len(groups))
	}
}

func observation(
	id string,
	icao24 string,
	flightID string,
	callsign string,
	onGround bool,
	observedAt time.Time,
) Observation {
	return Observation{
		State: flightstate.FlightState{
			ID:         id,
			ICAO24:     icao24,
			FlightID:   flightID,
			Callsign:   callsign,
			OnGround:   onGround,
			ObservedAt: observedAt,
		},
		QualityScore: 0.9,
	}
}

func testTime() time.Time {
	return time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
}
