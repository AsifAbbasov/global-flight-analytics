package trajectorybuilder

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestCalculateSamplingMetricsSortsTimestampCopy(t *testing.T) {
	base := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	points := []trajectory.TrackPoint4D{
		{ObservedAt: base.Add(30 * time.Second)},
		{ObservedAt: time.Time{}},
		{ObservedAt: base},
		{ObservedAt: base.Add(10 * time.Second)},
		{ObservedAt: base.Add(10 * time.Second)},
	}

	metrics, limitations :=
		calculateSamplingMetrics(points)

	if !metrics.available ||
		metrics.meanSeconds != 10 ||
		metrics.maximumSeconds != 20 {
		t.Fatalf(
			"metrics = %#v",
			metrics,
		)
	}
	if !hasLimitation(
		limitations,
		"trajectory_sampling_timestamp_missing",
	) || !hasLimitation(
		limitations,
		"trajectory_sampling_input_nonmonotonic",
	) {
		t.Fatalf(
			"missing sampling limitations: %#v",
			limitations,
		)
	}
}

func TestCalculateSamplingMetricsRequiresTwoTimestamps(
	t *testing.T,
) {
	metrics, limitations :=
		calculateSamplingMetrics(
			[]trajectory.TrackPoint4D{
				{
					ObservedAt: time.Date(
						2026,
						time.July,
						14,
						8,
						0,
						0,
						0,
						time.UTC,
					),
				},
			},
		)

	if metrics.available {
		t.Fatalf(
			"metrics unexpectedly available: %#v",
			metrics,
		)
	}
	if !hasLimitation(
		limitations,
		"trajectory_sampling_evidence_insufficient",
	) {
		t.Fatalf(
			"missing insufficient-evidence limitation: %#v",
			limitations,
		)
	}
}
