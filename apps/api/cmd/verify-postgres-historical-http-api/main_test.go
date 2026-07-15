package main

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestBuildVerificationScheduleUsesClosedHourlyBoundary(
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
		t.Fatalf("build schedule: %v", err)
	}

	expectedAsOf := now.UTC()
	expectedBoundary := expectedAsOf.Truncate(time.Hour)
	if !schedule.AsOfTime.Equal(expectedAsOf) ||
		!schedule.GeneratedAt.Equal(expectedAsOf) ||
		!schedule.ClosedBoundary.Equal(expectedBoundary) {
		t.Fatalf("unexpected schedule: %#v", schedule)
	}
}

func TestBuildVerificationScheduleRejectsZeroTime(
	t *testing.T,
) {
	if _, err := buildVerificationSchedule(time.Time{}); err == nil {
		t.Fatal("expected zero time to be rejected")
	}
}

func TestBuildVerificationResultsCreatesOrderedGlobalAndRouteEvidence(
	t *testing.T,
) {
	schedule, err := buildVerificationSchedule(
		time.Date(
			2026,
			time.July,
			15,
			12,
			30,
			0,
			123,
			time.UTC,
		),
	)
	if err != nil {
		t.Fatalf("build schedule: %v", err)
	}

	results, err := buildVerificationResults(schedule)
	if err != nil {
		t.Fatalf("build results: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("result count = %d, want 4", len(results))
	}

	expectedTotals := []float64{1, 2, 3}
	for index := 0; index < 3; index++ {
		result := results[index]
		if result.Metric.Name != historicalcontract.MetricNameFlightCount ||
			result.Scope.Type != historicalcontract.ScopeTypeGlobal ||
			result.Summary.Total != expectedTotals[index] ||
			result.Comparison == nil {
			t.Fatalf("unexpected global result[%d]: %#v", index, result)
		}
	}

	route := results[3]
	if route.Metric.Name != historicalcontract.MetricNameRouteObservations ||
		route.Scope.Type != historicalcontract.ScopeTypeRoute ||
		route.Scope.OriginICAOCode != verificationOriginICAO ||
		route.Scope.DestinationICAOCode != verificationDestinationICAO ||
		route.Summary.Total != 4 {
		t.Fatalf("unexpected route result: %#v", route)
	}
}

func TestRuntimeVerificationIdentityIsPinned(
	t *testing.T,
) {
	if expectedMigrationVersion != "015" ||
		expectedMigrationName != "create_historical_aggregate_results" ||
		len(expectedMigrationChecksum) != 64 {
		t.Fatal("migration 015 identity is not pinned")
	}
}
