package postgres

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFlightStateRepositoryDoesNotDelegateAltitudeRoundingToSQL(
	t *testing.T,
) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve altitude policy test path")
	}

	repositoryPath := filepath.Join(
		filepath.Dir(currentFile),
		"flightstate_repository.go",
	)
	content, err := os.ReadFile(repositoryPath)
	if err != nil {
		t.Fatalf("read flight state repository: %v", err)
	}

	source := string(content)
	for _, forbidden := range []string{
		"double precision AS integer",
		"CAST($7",
		"CAST($9",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf(
				"flight state repository still delegates altitude conversion to SQL through %q",
				forbidden,
			)
		}
	}

	for _, required := range []string{
		"altitudeMetersToPostgresInteger(value)",
		"pgtype.Int4",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf(
				"flight state repository is missing explicit altitude policy surface %q",
				required,
			)
		}
	}
}
