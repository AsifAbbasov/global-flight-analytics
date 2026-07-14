package historicalread

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestSnapshotCloneDoesNotShareMutableState(t *testing.T) {
	latitude := 40.1
	longitude := 49.9
	onGround := false

	snapshot := Snapshot{
		Version: Version,
		Query: Query{
			Window: historicalcontract.TimeWindow{
				StartTime: time.Now().UTC().Add(-time.Hour),
				EndTime:   time.Now().UTC(),
				AsOfTime:  time.Now().UTC().Add(time.Hour),
			},
			Limit: 10,
		},
		Flights:      []FlightRecord{{ID: "flight-1"}},
		Trajectories: []TrajectoryRecord{{ID: "trajectory-1"}},
		Observations: []ObservationRecord{
			{
				ID:        "observation-1",
				Latitude:  &latitude,
				Longitude: &longitude,
				OnGround:  &onGround,
			},
		},
		Routes: []RouteRecord{
			{
				ID:        "route-1",
				RouteJSON: []byte(`{"status":"complete"}`),
			},
		},
	}

	cloned := snapshot.Clone()
	cloned.Flights[0].ID = "changed"
	cloned.Trajectories[0].ID = "changed"
	*cloned.Observations[0].Latitude = 0
	*cloned.Observations[0].Longitude = 0
	*cloned.Observations[0].OnGround = true
	cloned.Routes[0].RouteJSON[0] = '['

	if snapshot.Flights[0].ID != "flight-1" {
		t.Fatal("Snapshot.Clone() shared flight state")
	}
	if snapshot.Trajectories[0].ID != "trajectory-1" {
		t.Fatal("Snapshot.Clone() shared trajectory state")
	}
	if *snapshot.Observations[0].Latitude != latitude ||
		*snapshot.Observations[0].Longitude != longitude ||
		*snapshot.Observations[0].OnGround != onGround {
		t.Fatal("Snapshot.Clone() shared observation pointer state")
	}
	if string(snapshot.Routes[0].RouteJSON) != `{"status":"complete"}` {
		t.Fatal("Snapshot.Clone() shared route JSON")
	}
}

func TestRepositoryContractVersionAndLimits(t *testing.T) {
	if Version != "historical-read-repository-v1" {
		t.Fatalf("Version = %q", Version)
	}
	if DefaultDatasetLimit != 10_000 {
		t.Fatalf("DefaultDatasetLimit = %d", DefaultDatasetLimit)
	}
	if MaximumDatasetLimit != 100_000 {
		t.Fatalf("MaximumDatasetLimit = %d", MaximumDatasetLimit)
	}
}
