package trajectorypolicy

import (
	"testing"
	"time"
)

func TestDefaultTrajectoryPolicyValues(t *testing.T) {
	if DefaultMaxTimeGap != 90*time.Second {
		t.Fatalf("expected default max time gap 90 seconds, got %s", DefaultMaxTimeGap)
	}

	if DefaultMaxGroundSpeedMetersPerSecond != 420.0 {
		t.Fatalf("expected default max ground speed 420.0, got %f", DefaultMaxGroundSpeedMetersPerSecond)
	}
}
