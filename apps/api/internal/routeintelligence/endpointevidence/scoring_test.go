package endpointevidence

import (
	"math"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/airportresolver"
)

func TestScoreCandidateUsesStableWeights(t *testing.T) {
	breakdown := scoreCandidate(
		airportresolver.Candidate{
			ProximityScore: 0.8,
		},
		0.6,
		trajectory.SegmentStatusInterpolated,
		5,
	)

	want := 0.5*0.8 +
		0.25*0.6 +
		0.15*0.7 +
		0.10
	if math.Abs(breakdown.Total-want) > 1e-12 {
		t.Fatalf(
			"score = %.12f, want %.12f",
			breakdown.Total,
			want,
		)
	}
	if math.Abs(
		breakdown.ProximityContribution-0.4,
	) > 1e-12 ||
		math.Abs(
			breakdown.TrajectoryQualityContribution-0.15,
		) > 1e-12 ||
		math.Abs(
			breakdown.SegmentStatusContribution-0.105,
		) > 1e-12 ||
		math.Abs(
			breakdown.PointEvidenceContribution-0.10,
		) > 1e-12 {
		t.Fatalf(
			"unexpected breakdown: %#v",
			breakdown,
		)
	}
}

func TestSegmentStatusScore(t *testing.T) {
	tests := []struct {
		status trajectory.SegmentStatus
		want   float64
	}{
		{
			status: trajectory.SegmentStatusObserved,
			want:   1,
		},
		{
			status: trajectory.SegmentStatusInterpolated,
			want:   0.7,
		},
		{
			status: trajectory.SegmentStatusEstimated,
			want:   0.45,
		},
		{
			status: trajectory.SegmentStatusInvalid,
			want:   0,
		},
	}

	for _, test := range tests {
		if got := segmentStatusScore(
			test.status,
		); got != test.want {
			t.Fatalf(
				"segmentStatusScore(%q) = %v, want %v",
				test.status,
				got,
				test.want,
			)
		}
	}
}
