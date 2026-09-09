package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMessageObservationMigrationKeepsPositionTimeCanonical(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file")
	}
	path := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../../../../database/migrations/030_add_flight_state_message_observation_time.sql"))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	text := string(content)
	for _, required := range []string{"ADD COLUMN message_observed_at timestamptz", "Time of the position observation", "must not replace position freshness"} {
		if !strings.Contains(text, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	if strings.Contains(text, "UPDATE flight_states") {
		t.Fatal("migration must not fabricate historical message timestamps")
	}
}
