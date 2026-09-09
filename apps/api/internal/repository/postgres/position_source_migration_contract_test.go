package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPositionSourceMigrationExpandsEvidenceWithoutRewritingHistory(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file")
	}
	path := filepath.Clean(filepath.Join(
		filepath.Dir(currentFile),
		"../../../../../database/migrations/031_expand_flight_state_position_sources.sql",
	))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	text := string(content)
	for _, required := range []string{
		"DROP CONSTRAINT flight_states_position_source_check",
		"'adsb'",
		"'adsr'",
		"'tisb'",
		"'adsc'",
		"'asterix'",
		"'mlat'",
		"'flarm'",
		"provider/feed identity remains in source_name",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	if strings.Contains(text, "UPDATE flight_states") {
		t.Fatal("migration must not rewrite historical position-source evidence")
	}
}
