package postgres

import (
	"os"
	"strings"
	"testing"
)

func TestTrajectoryPointQueriesPreserveNullableOperationalTelemetry(t *testing.T) {
	for name, query := range map[string]string{
		"flight": trajectoryPointsByFlightIDAndWindowQuery,
		"icao24": trajectoryPointsByICAO24AndWindowQuery,
	} {
		upper := strings.ToUpper(query)
		for _, forbidden := range []string{
			"COALESCE(STATE.VELOCITY_MPS",
			"COALESCE(STATE.HEADING_DEGREES",
			"COALESCE(STATE.VERTICAL_RATE_MPS",
			"COALESCE(STATE.ON_GROUND",
		} {
			if strings.Contains(upper, forbidden) {
				t.Fatalf("%s query fabricates nullable telemetry with %q", name, forbidden)
			}
		}
		for _, required := range []string{
			"STATE.OBSERVED_AT >= $2",
			"STATE.OBSERVED_AT <= $3",
			"STATE.LATITUDE IS NOT NULL",
			"STATE.LONGITUDE IS NOT NULL",
			"ORDER BY",
			"STATE.OBSERVED_AT ASC",
			"STATE.ID ASC",
		} {
			if !strings.Contains(upper, required) {
				t.Fatalf("%s query is missing %q", name, required)
			}
		}
	}
}

func TestFeatureTrajectoryReaderOwnsExplicitPointHydration(t *testing.T) {
	content, err := os.ReadFile("trajectory_feature_read.go")
	if err != nil {
		t.Fatalf("read trajectory_feature_read.go: %v", err)
	}
	source := string(content)
	for _, required := range []string{
		"type FeatureTrajectoryReader struct",
		"repository.withTrajectoryReadSnapshot(",
		"snapshot.getTrajectoryByID(",
		"snapshot.getLatestTrajectoryByICAO24(",
		"repository.listTrajectoryPoints(",
		"item.Points = points",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("feature trajectory reader is missing %q", required)
		}
	}

	ordinaryContent, err := os.ReadFile("trajectory_child_read.go")
	if err != nil {
		t.Fatalf("read trajectory_child_read.go: %v", err)
	}
	if strings.Contains(string(ordinaryContent), "listTrajectoryPoints") {
		t.Fatal("ordinary trajectory reads unexpectedly hydrate point telemetry")
	}
}
