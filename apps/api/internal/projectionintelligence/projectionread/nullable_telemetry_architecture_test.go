package projectionread

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProjectionReadNullableTelemetryBoundaryRemainsExplicit(
	t *testing.T,
) {
	queries := readProjectionReadSource(
		t,
		"postgres_queries.go",
	)
	source := readProjectionReadSource(
		t,
		"postgres_source.go",
	)

	for _, forbidden := range []string{
		"COALESCE(latitude, 0)",
		"COALESCE(longitude, 0)",
		"COALESCE(velocity_mps, 0)",
		"COALESCE(heading_degrees, 0)",
		"COALESCE(vertical_rate_mps, 0)",
		"COALESCE(on_ground, false)",
	} {
		if strings.Contains(
			queries,
			forbidden,
		) {
			t.Fatalf(
				"nullable telemetry regression found: %q",
				forbidden,
			)
		}
	}

	for _, required := range []string{
		"var latitude pgtype.Float8",
		"var longitude pgtype.Float8",
		"var velocity pgtype.Float8",
		"var heading pgtype.Float8",
		"var verticalRate pgtype.Float8",
		"var onGround pgtype.Bool",
		"completeRequiredTelemetry(",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf(
				"nullable telemetry boundary is missing %q",
				required,
			)
		}
	}
}

func readProjectionReadSource(
	t *testing.T,
	name string,
) string {
	t.Helper()

	_, currentFile, _, ok :=
		runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve Projection Intelligence source path",
		)
	}

	content, err := os.ReadFile(
		filepath.Join(
			filepath.Dir(currentFile),
			name,
		),
	)
	if err != nil {
		t.Fatalf(
			"read %s: %v",
			name,
			err,
		)
	}

	return string(content)
}
