package airplaneslive

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAirplanesLiveMapperDeclaresTelemetryAvailability(
	t *testing.T,
) {
	_, currentFile, _, ok :=
		runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve mapper source path",
		)
	}

	content, err := os.ReadFile(
		filepath.Join(
			filepath.Dir(currentFile),
			"mapper.go",
		),
	)
	if err != nil {
		t.Fatalf(
			"read mapper source: %v",
			err,
		)
	}

	text := string(content)
	for _, required := range []string{
		"TelemetryAvailabilityKnown:",
		"VelocityAvailable:",
		"HeadingAvailable:",
		"VerticalRateAvailable:",
		"OnGroundAvailable:",
	} {
		if !strings.Contains(
			text,
			required,
		) {
			t.Fatalf(
				"airplanes.live mapper is missing %q",
				required,
			)
		}
	}
}
