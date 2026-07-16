package dto

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceregionanalytics"
)

func TestToAirspaceRegionAnalyticsResponse(
	t *testing.T,
) {
	now := time.Date(
		2026, time.July, 17, 12, 0, 0, 0, time.UTC,
	)
	result := airspaceregionanalytics.Result{
		SchemaVersion: airspaceregionanalytics.SchemaVersionV1,
		Status:        airspaceregionanalytics.ResultStatusLimited,
		RegionCode:    "AZERBAIJAN",
		WindowStart:   now.Add(-time.Minute),
		WindowEnd:     now,
		Occupancy: airspaceregionanalytics.TemporalOccupancyIndex{
			BucketDuration: time.Minute,
			Metrics: airspaceregionanalytics.OccupancyIndexMetrics{
				ExpectedBucketCount: 1,
			},
		},
		Metrics: airspaceregionanalytics.RegionMetrics{
			SnapshotCount: 1,
		},
		ScopeGuard:  airspaceregionanalytics.ScopeGuardResearchOnly,
		GeneratedAt: now,
	}

	response := ToAirspaceRegionAnalyticsResponse(result)
	if response.Version != airspaceproduction.Version {
		t.Fatalf("Version = %q", response.Version)
	}
	if response.RegionCode != "AZERBAIJAN" {
		t.Fatalf("RegionCode = %q", response.RegionCode)
	}
	if response.Occupancy.BucketDurationSeconds != 60 {
		t.Fatalf(
			"BucketDurationSeconds = %d",
			response.Occupancy.BucketDurationSeconds,
		)
	}
	if response.Metrics.SnapshotCount != 1 {
		t.Fatalf("SnapshotCount = %d", response.Metrics.SnapshotCount)
	}
}
