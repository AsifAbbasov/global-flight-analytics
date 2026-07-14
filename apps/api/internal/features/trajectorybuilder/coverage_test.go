package trajectorybuilder

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestCalculateCoverageRatioMergesAndClipsGaps(
	t *testing.T,
) {
	start := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	end := start.Add(100 * time.Second)
	item := trajectory.FlightTrajectory{
		StartTime: start,
		EndTime:   end,
		CoverageGaps: []trajectory.CoverageGap{
			{
				ID:              "left-clipped",
				StartTime:       start.Add(-10 * time.Second),
				EndTime:         start.Add(20 * time.Second),
				DurationSeconds: 30,
			},
			{
				ID:              "overlap",
				StartTime:       start.Add(10 * time.Second),
				EndTime:         start.Add(40 * time.Second),
				DurationSeconds: 30,
			},
			{
				ID:              "right",
				StartTime:       start.Add(80 * time.Second),
				EndTime:         end.Add(10 * time.Second),
				DurationSeconds: 30,
			},
			{
				ID:              "outside",
				StartTime:       end.Add(10 * time.Second),
				EndTime:         end.Add(20 * time.Second),
				DurationSeconds: 10,
			},
		},
	}

	metric, limitations :=
		calculateCoverageRatio(item)

	if !metric.available ||
		!approximatelyEqual(metric.value, 0.4, 1e-12) {
		t.Fatalf(
			"coverage = %#v, want 0.4",
			metric,
		)
	}
	if !hasLimitation(
		limitations,
		"trajectory_coverage_gap_outside_window",
	) {
		t.Fatalf(
			"missing outside-window limitation: %#v",
			limitations,
		)
	}
}

func TestCalculateCoverageRatioDetectsDurationMismatch(
	t *testing.T,
) {
	start := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	item := trajectory.FlightTrajectory{
		StartTime: start,
		EndTime:   start.Add(time.Minute),
		CoverageGaps: []trajectory.CoverageGap{
			{
				ID:              "gap",
				StartTime:       start.Add(10 * time.Second),
				EndTime:         start.Add(20 * time.Second),
				DurationSeconds: 99,
			},
		},
	}

	metric, limitations :=
		calculateCoverageRatio(item)

	if !metric.available ||
		!approximatelyEqual(
			metric.value,
			50.0/60.0,
			1e-12,
		) {
		t.Fatalf(
			"coverage = %#v",
			metric,
		)
	}
	if !hasLimitation(
		limitations,
		"trajectory_coverage_gap_duration_metadata_mismatch",
	) {
		t.Fatalf(
			"missing duration mismatch: %#v",
			limitations,
		)
	}
}

func TestCalculateCoverageRatioRejectsInvalidGapWindow(
	t *testing.T,
) {
	start := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	item := trajectory.FlightTrajectory{
		StartTime: start,
		EndTime:   start.Add(time.Minute),
		CoverageGaps: []trajectory.CoverageGap{
			{
				ID:        "reversed",
				StartTime: start.Add(20 * time.Second),
				EndTime:   start.Add(10 * time.Second),
			},
		},
	}

	metric, limitations :=
		calculateCoverageRatio(item)

	if metric.available {
		t.Fatalf(
			"coverage unexpectedly available: %#v",
			metric,
		)
	}
	if !hasLimitation(
		limitations,
		"trajectory_coverage_gap_window_invalid",
	) {
		t.Fatalf(
			"missing invalid-gap limitation: %#v",
			limitations,
		)
	}
}

func TestCalculateCoverageRatioHandlesZeroDuration(
	t *testing.T,
) {
	instant := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)

	metric, limitations :=
		calculateCoverageRatio(
			trajectory.FlightTrajectory{
				StartTime: instant,
				EndTime:   instant,
			},
		)
	if !metric.available ||
		metric.value != 1 ||
		len(limitations) != 0 {
		t.Fatalf(
			"zero-duration coverage = %#v, %#v",
			metric,
			limitations,
		)
	}

	metric, limitations =
		calculateCoverageRatio(
			trajectory.FlightTrajectory{
				StartTime: instant,
				EndTime:   instant,
				CoverageGaps: []trajectory.CoverageGap{
					{},
				},
			},
		)
	if metric.available ||
		!hasLimitation(
			limitations,
			"trajectory_coverage_zero_duration_with_gaps",
		) {
		t.Fatalf(
			"zero-duration gap coverage = %#v, %#v",
			metric,
			limitations,
		)
	}
}

func TestUnionDurationDoesNotDoubleCountOverlap(
	t *testing.T,
) {
	start := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)
	duration := unionDuration(
		[]timeInterval{
			{
				start: start.Add(10 * time.Second),
				end:   start.Add(30 * time.Second),
			},
			{
				start: start.Add(20 * time.Second),
				end:   start.Add(40 * time.Second),
			},
			{
				start: start.Add(50 * time.Second),
				end:   start.Add(60 * time.Second),
			},
		},
	)

	if duration != 40*time.Second {
		t.Fatalf(
			"union duration = %v, want 40s",
			duration,
		)
	}
}
