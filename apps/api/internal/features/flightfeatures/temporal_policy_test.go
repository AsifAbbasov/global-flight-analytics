package flightfeatures

import (
	"testing"
	"time"
)

func TestTemporalDurationPolicyTruncatesFractionalSeconds(t *testing.T) {
	startTime := time.Date(
		2026,
		time.July,
		27,
		10,
		0,
		0,
		250_000_000,
		time.UTC,
	)
	endTime := startTime.Add(1750 * time.Millisecond)

	if CurrentTemporalDurationRoundingPolicy !=
		TemporalDurationRoundingPolicyTruncateFractionalSeconds {
		t.Fatalf(
			"current temporal duration policy = %q",
			CurrentTemporalDurationRoundingPolicy,
		)
	}
	if got := TemporalDurationSeconds(startTime, endTime); got != 1 {
		t.Fatalf("duration seconds = %d, want 1", got)
	}
	if got := TemporalDurationSeconds(endTime, startTime); got != -1 {
		t.Fatalf("negative duration seconds = %d, want -1", got)
	}
}

func TestTemporalDurationPolicyPreservesWholeAndZeroSeconds(t *testing.T) {
	startTime := time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC)

	if got := TemporalDurationSeconds(startTime, startTime); got != 0 {
		t.Fatalf("zero duration seconds = %d, want 0", got)
	}
	if got := TemporalDurationSeconds(
		startTime,
		startTime.Add(90*time.Second),
	); got != 90 {
		t.Fatalf("whole duration seconds = %d, want 90", got)
	}
}
