package trackbuilder

import (
	"testing"
	"time"
)

func TestBuildManySeparatesMultipleFlightsForSameAircraft(t *testing.T) {
	builder := mustNewBuilder(t, Config{})
	observedAt := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)

	firstFlight := makeFlightState(
		"state-1",
		"ABC123",
		"AHY101",
		40.41,
		49.87,
		observedAt,
	)
	firstFlight.FlightID = "11111111-1111-1111-1111-111111111111"

	secondFlight := makeFlightState(
		"state-2",
		"ABC123",
		"AHY102",
		40.42,
		49.88,
		observedAt.Add(time.Minute),
	)
	secondFlight.FlightID = "22222222-2222-2222-2222-222222222222"

	result := builder.BuildMany([]InputState{
		{State: secondFlight, QualityScore: 0.8},
		{State: firstFlight, QualityScore: 0.9},
	})

	if len(result) != 2 {
		t.Fatalf("expected 2 trajectories, got %d", len(result))
	}

	identities := make(map[string]struct{}, 2)
	for _, item := range result {
		if item.ICAO24 != "ABC123" {
			t.Fatalf("expected ABC123, got %s", item.ICAO24)
		}
		if item.PointCount != 1 {
			t.Fatalf("expected one point per flight, got %d", item.PointCount)
		}
		if item.IdentityKey == "" {
			t.Fatal("expected identity key")
		}
		identities[item.IdentityKey] = struct{}{}
	}

	if len(identities) != 2 {
		t.Fatalf("expected 2 distinct identities, got %d", len(identities))
	}
}
