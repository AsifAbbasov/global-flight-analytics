package processor

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestProcessDoesNotMergeDifferentFlightsOfSameAircraft(t *testing.T) {
	now := fixedTime()
	trafficProcessor := mustNewProcessor(t, Config{
		Now: func() time.Time { return now },
	})

	first := makeProcessorFlightState(
		"state-1",
		"ABC123",
		"AHY101",
		40.41,
		49.87,
		now.Add(-2*time.Minute),
	)
	first.FlightID = "11111111-1111-1111-1111-111111111111"

	second := makeProcessorFlightState(
		"state-2",
		"ABC123",
		"AHY102",
		40.42,
		49.88,
		now.Add(-time.Minute),
	)
	second.FlightID = "22222222-2222-2222-2222-222222222222"

	result := trafficProcessor.Process([]flightstate.FlightState{second, first})
	if result.Stats.TrajectoryCount != 2 {
		t.Fatalf("expected 2 trajectories, got %d", result.Stats.TrajectoryCount)
	}

	for _, item := range result.Trajectories {
		if item.PointCount != 1 {
			t.Fatalf("expected one point per flight, got %d", item.PointCount)
		}
	}
}
