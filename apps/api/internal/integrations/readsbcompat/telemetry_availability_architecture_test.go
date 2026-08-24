package readsbcompat

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSharedMapperDeclaresTelemetryAvailability(
	t *testing.T,
) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve shared mapper source path")
	}

	content, err := os.ReadFile(
		filepath.Join(filepath.Dir(currentFile), "mapper.go"),
	)
	if err != nil {
		t.Fatalf("read shared mapper source: %v", err)
	}

	text := string(content)
	for _, required := range []string{
		"TelemetryAvailabilityKnown:",
		"VelocityAvailable:",
		"HeadingAvailable:",
		"VerticalRateAvailable:",
		"OnGroundAvailable:",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("shared readsb-compatible mapper is missing %q", required)
		}
	}
}
