package trajectory

import (
	"reflect"
	"testing"
)

func TestFlightTrajectoryCloneDoesNotShareEvidenceSlices(t *testing.T) {
	original := FlightTrajectory{
		ID:           "trajectory-one",
		Points:       []TrackPoint4D{{ID: "point-one"}},
		Segments:     []TrajectorySegment{{ID: "segment-one"}},
		CoverageGaps: []CoverageGap{{ID: "gap-one"}},
	}
	cloned := original.Clone()
	if !reflect.DeepEqual(original, cloned) {
		t.Fatalf("Clone() differs from source: source=%#v clone=%#v", original, cloned)
	}

	cloned.Points[0].ID = "changed-point"
	cloned.Segments[0].ID = "changed-segment"
	cloned.CoverageGaps[0].ID = "changed-gap"
	if original.Points[0].ID != "point-one" ||
		original.Segments[0].ID != "segment-one" ||
		original.CoverageGaps[0].ID != "gap-one" {
		t.Fatal("Clone() shares mutable evidence slice storage")
	}
}

func TestFlightTrajectoryClonePreservesNilSlices(t *testing.T) {
	cloned := (FlightTrajectory{ID: "trajectory-one"}).Clone()
	if cloned.Points != nil || cloned.Segments != nil || cloned.CoverageGaps != nil {
		t.Fatalf("Clone() changed nil slice semantics: %#v", cloned)
	}
}
