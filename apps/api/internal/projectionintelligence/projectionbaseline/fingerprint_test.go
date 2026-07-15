package projectionbaseline

import (
	"strings"
	"testing"
)

func TestInputFingerprintIsDeterministicAndSensitive(
	t *testing.T,
) {
	item := baselineTestTrajectory()
	point := item.Points[len(item.Points)-1]
	plan := baselineTestPlan()
	config := validBaselineConfig()

	first := inputFingerprint(
		item,
		point,
		plan,
		config,
	)
	second := inputFingerprint(
		item,
		point,
		plan,
		config,
	)

	if first != second {
		t.Fatalf(
			"fingerprints differ: %q != %q",
			first,
			second,
		)
	}
	if !strings.HasPrefix(
		first,
		fingerprintPrefix,
	) {
		t.Fatalf(
			"fingerprint = %q, missing prefix",
			first,
		)
	}

	point.HeadingDegrees++
	changed := inputFingerprint(
		item,
		point,
		plan,
		config,
	)
	if changed == first {
		t.Fatal(
			"fingerprint ignored changed kinematic input",
		)
	}
}
