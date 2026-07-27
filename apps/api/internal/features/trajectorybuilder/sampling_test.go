package trajectorybuilder

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/features/flightfeatures"
)

func TestSamplingUsesUniqueCanonicalTimestamps(t *testing.T) {
	base := time.Date(2026, time.July, 14, 8, 0, 0, 0, time.UTC)
	evidence, err := canonicalizeEvidence(context.Background(), trajectory.FlightTrajectory{
		ID: "trajectory", StartTime: base, EndTime: base.Add(30 * time.Second), PointCount: 4,
		Points: []trajectory.TrackPoint4D{
			{ID: "d", ObservedAt: base.Add(30 * time.Second)},
			{ID: "a", ObservedAt: base},
			{ID: "c", ObservedAt: base.Add(10 * time.Second)},
			{ID: "b", ObservedAt: base.Add(10 * time.Second)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	metrics, limitations, err := calculateSamplingMetrics(context.Background(), evidence.points)
	if err != nil {
		t.Fatal(err)
	}
	if !metrics.available || metrics.meanSeconds != 15 || metrics.maximumSeconds != 20 {
		t.Fatalf("metrics = %#v", metrics)
	}
	if len(limitations) != 0 || !hasLimitation(evidence.limitations, flightfeatures.TrajectoryLimitationDuplicateTimestampsCollapsed) {
		t.Fatalf("limitations = %#v / %#v", limitations, evidence.limitations)
	}
}

func TestSamplingRequiresTwoUniqueTimestamps(t *testing.T) {
	metrics, limitations, err := calculateSamplingMetrics(context.Background(), []canonicalPoint{{observedAt: time.Now().UTC()}})
	if err != nil {
		t.Fatal(err)
	}
	if metrics.available || !hasLimitation(limitations, flightfeatures.TrajectoryLimitationSamplingEvidenceInsufficient) {
		t.Fatalf("metrics=%#v limitations=%#v", metrics, limitations)
	}
}
