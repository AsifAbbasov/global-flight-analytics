package historicalaggregate

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHistoricalAggregateHardeningMigrationContract(
	t *testing.T,
) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve migration contract test path",
		)
	}
	path := filepath.Clean(
		filepath.Join(
			filepath.Dir(currentFile),
			"../../../../../database/migrations/029_harden_historical_aggregate_integrity.sql",
		),
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"read hardening migration: %v",
			err,
		)
	}
	text := string(content)

	required := []string{
		"region_code ~ '^[a-z0-9][a-z0-9_-]{1,31}$'",
		"historical_aggregate_results_timestamp_mirror_check",
		"extract(epoch FROM window_start) * 1000000000",
		"historical_aggregate_results_json_metadata_check",
		"historical_aggregate_results_stored_at_causality_check",
		"result_json #>> '{Metric,Name}'",
		"result_json #>> '{Window,StartTime}'",
		"result_json #>> '{Provenance,InputFingerprint}'",
	}
	for _, fragment := range required {
		if !strings.Contains(text, fragment) {
			t.Fatalf(
				"migration misses %q",
				fragment,
			)
		}
	}
	if strings.Contains(
		text,
		"region_code ~ '^[A-Z0-9_-]",
	) {
		t.Fatal(
			"migration retains uppercase region contract",
		)
	}
}
