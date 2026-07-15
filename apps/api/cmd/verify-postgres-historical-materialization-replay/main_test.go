package main

import (
	"testing"
	"time"
)

func TestBuildVerificationScheduleUsesClosedHourlyWindows(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		15,
		12,
		37,
		45,
		123456789,
		time.FixedZone("Asia/Baku", 4*60*60),
	)

	schedule, err := buildVerificationSchedule(now)
	if err != nil {
		t.Fatalf("build verification schedule: %v", err)
	}

	expectedAsOf := now.UTC()
	expectedBoundary := expectedAsOf.Truncate(time.Hour)

	if !schedule.AsOfTime.Equal(expectedAsOf) ||
		!schedule.GeneratedAt.Equal(expectedAsOf) {
		t.Fatalf("unexpected as-of or generated time: %#v", schedule)
	}
	if !schedule.MaterializationStart.Equal(
		expectedBoundary.Add(-2*time.Hour),
	) ||
		!schedule.MaterializationEnd.Equal(expectedBoundary) {
		t.Fatalf("unexpected materialization range: %#v", schedule)
	}
	if !schedule.ReplayStart.Equal(schedule.MaterializationStart) ||
		!schedule.ReplayEnd.Equal(schedule.MaterializationEnd) {
		t.Fatalf("replay range differs from materialization range: %#v", schedule)
	}
	if schedule.MaterializationStart.Minute() != 0 ||
		schedule.MaterializationStart.Second() != 0 ||
		schedule.MaterializationEnd.Minute() != 0 ||
		schedule.MaterializationEnd.Second() != 0 {
		t.Fatalf("verification ranges are not hour-aligned: %#v", schedule)
	}
}

func TestBuildVerificationScheduleRejectsZeroTime(
	t *testing.T,
) {
	if _, err := buildVerificationSchedule(time.Time{}); err == nil {
		t.Fatal("expected zero verification time to be rejected")
	}
}

func TestRuntimeVerificationConstantsArePinned(
	t *testing.T,
) {
	if expectedMigrationVersion != "015" {
		t.Fatalf("unexpected migration version: %s", expectedMigrationVersion)
	}
	if expectedMigrationName !=
		"create_historical_aggregate_results" {
		t.Fatalf("unexpected migration name: %s", expectedMigrationName)
	}
	if len(expectedMigrationChecksum) != 64 {
		t.Fatalf(
			"unexpected migration checksum length: %d",
			len(expectedMigrationChecksum),
		)
	}
	if verificationDatasetLimit < 1 ||
		verificationMaximumBucketCount < 4 ||
		verificationMaximumWindowCount < 2 {
		t.Fatal("runtime verification bounds are unsafe")
	}
}
