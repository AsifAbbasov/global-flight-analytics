package projectionread

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestFilterSegmentsAtExcludesIncompleteSegment(t *testing.T) {
	cutoff := time.Date(2026, time.July, 15, 17, 0, 0, 0, time.UTC)
	items := []trajectory.TrajectorySegment{
		{
			ID:              "complete",
			SequenceNumber:  1,
			StartTime:       cutoff.Add(-time.Minute),
			EndTime:         cutoff,
			DurationSeconds: 60,
			PointCount:      2,
			QualityScore:    0.9,
		},
		{
			ID:              "incomplete",
			SequenceNumber:  2,
			StartTime:       cutoff.Add(-30 * time.Second),
			EndTime:         cutoff.Add(30 * time.Second),
			DurationSeconds: 60,
			PointCount:      2,
			QualityScore:    0.1,
		},
	}

	result := filterSegmentsAt(items, cutoff)
	if len(result) != 1 || result[0].ID != "complete" {
		t.Fatalf("segments = %#v, want only complete segment", result)
	}
}

func TestFilterCoverageGapsAtExcludesGapEndingAfterCutoff(t *testing.T) {
	cutoff := time.Date(2026, time.July, 15, 17, 0, 0, 0, time.UTC)
	items := []trajectory.CoverageGap{
		{
			ID:        "complete",
			StartTime: cutoff.Add(-time.Minute),
			EndTime:   cutoff,
		},
		{
			ID:        "future-created",
			StartTime: cutoff.Add(-30 * time.Second),
			EndTime:   cutoff.Add(30 * time.Second),
		},
	}

	result := filterCoverageGapsAt(items, cutoff)
	if len(result) != 1 || result[0].ID != "complete" {
		t.Fatalf("gaps = %#v, want only completed gap", result)
	}
}
