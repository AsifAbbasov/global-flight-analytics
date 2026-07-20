package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestTrafficRepositoryOwnsTypedAltitudeSelection(
	t *testing.T,
) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve traffic altitude ownership test path")
	}

	repositoryPath := filepath.Join(
		filepath.Dir(currentFile),
		"traffic_repository.go",
	)
	contentBytes, err := os.ReadFile(repositoryPath)
	if err != nil {
		t.Fatalf(
			"read traffic repository: %v",
			err,
		)
	}
	content := string(contentBytes)

	for _, forbidden := range []string{
		"NULLIF(fs.geometric_altitude_m, 0)",
		"COALESCE(\n\t\t\t\tNULLIF(fs.geometric_altitude_m, 0)",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf(
				"traffic repository still uses zero as missing altitude: %q",
				forbidden,
			)
		}
	}

	for _, required := range []string{
		"fs.geometric_altitude_status",
		"fs.barometric_altitude_status",
		"traffic.ResolveCurrentAltitude",
		"item.AltitudeStatus",
		"item.AltitudeSource",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf(
				"traffic repository is missing altitude semantic token %q",
				required,
			)
		}
	}

	if count := strings.Count(
		content,
		"fs.geometric_altitude_status",
	); count != 2 {
		t.Fatalf(
			"geometric altitude status select count = %d, want 2",
			count,
		)
	}

	if count := strings.Count(
		content,
		"fs.barometric_altitude_status",
	); count != 2 {
		t.Fatalf(
			"barometric altitude status select count = %d, want 2",
			count,
		)
	}
}
