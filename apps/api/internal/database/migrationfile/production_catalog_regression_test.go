package migrationfile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"testing"
)

var ingestionDurabilityMigrationFilenamePattern = regexp.MustCompile(
	`^([0-9]{3})_[a-z0-9_]+\.sql$`,
)

func TestProductionMigrationCatalogHasUniqueVersions(
	t *testing.T,
) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve production migration catalog test path")
	}
	migrationsDir := filepath.Clean(filepath.Join(
		filepath.Dir(currentFile),
		"../../../../../database/migrations",
	))
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read production migrations: %v", err)
	}

	versions := make(map[int]string, len(entries))
	orderedVersions := make([]int, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := ingestionDurabilityMigrationFilenamePattern.FindStringSubmatch(
			entry.Name(),
		)
		if matches == nil {
			continue
		}
		version, err := strconv.Atoi(matches[1])
		if err != nil {
			t.Fatalf("parse migration version %q: %v", matches[1], err)
		}
		if previous, exists := versions[version]; exists {
			t.Fatalf(
				"duplicate migration version %03d: %s and %s",
				version,
				previous,
				entry.Name(),
			)
		}
		versions[version] = entry.Name()
		orderedVersions = append(orderedVersions, version)
	}

	sort.Ints(orderedVersions)
	for index, version := range orderedVersions {
		expected := index + 1
		if version != expected {
			t.Fatalf(
				"migration sequence gap: got %03d at index %d, want %03d; catalog=%v",
				version,
				index,
				expected,
				orderedVersions,
			)
		}
	}

	expectedCanonical := map[int]string{
		19: "019_data_quality_parent_integrity.sql",
		20: "020_stage14_correctness_hardening.sql",
		21: "021_trajectory_query_profiles.sql",
		22: "022_provider_publication_lifecycle.sql",
		23: "023_ingestion_durability_replay_partial.sql",
		24: "024_provider_budget_durability.sql",
		25: "025_weather_metric_availability.sql",
	}
	for version, filename := range expectedCanonical {
		if actual := versions[version]; actual != filename {
			t.Fatalf(
				"canonical migration %03d = %q, want %q",
				version,
				actual,
				filename,
			)
		}
	}

	if len(orderedVersions) != 25 {
		t.Fatalf(
			"production migration count = %d, want 25 (%s)",
			len(orderedVersions),
			fmt.Sprint(orderedVersions),
		)
	}
}
