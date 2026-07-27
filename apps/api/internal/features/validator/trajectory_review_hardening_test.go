package validator

import (
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestTrajectoryPathEfficiencyComparisonSkipsDifferentEvidenceSemantics(t *testing.T) {
	for _, code := range []string{
		flightfeatures.TrajectoryLimitationPathDiscontinuityExcluded,
		flightfeatures.TrajectoryLimitationPathSegmentFallback,
		flightfeatures.TrajectoryLimitationPathEvidenceInsufficient,
		flightfeatures.TrajectoryLimitationDuplicateTimestampsCollapsed,
		flightfeatures.TrajectoryLimitationInvalidPointCoordinates,
	} {
		evidence := flightfeatures.GroupEvidence{
			Limitations: []flightfeatures.FeatureLimitation{{Code: code}},
		}
		if trajectoryPathEfficiencyComparable(evidence) {
			t.Fatalf("evidence with %q remained cross-group comparable", code)
		}
	}
	if !trajectoryPathEfficiencyComparable(flightfeatures.GroupEvidence{}) {
		t.Fatal("ordinary point-path evidence became non-comparable")
	}
}
