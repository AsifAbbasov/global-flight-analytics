package metrics

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
)

func TestCalculateActiveAircraftFailsClosedForFutureObservation(t *testing.T) {
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	repository := &activeAircraftRepositoryStub{
		summary: ActiveAircraftObservationSummary{
			Count:            1,
			FirstObservedAt:  now.Add(-time.Minute),
			LatestObservedAt: now.Add(time.Second),
			SourceNames:      []string{"opensky"},
			HasObservations:  true,
		},
	}
	service := mustNewServiceWithClock(repository, region.NewService(), func() time.Time { return now })

	metric, err := service.CalculateActiveAircraft(context.Background(), ActiveAircraftRequest{})
	if err != nil {
		t.Fatalf("CalculateActiveAircraft() error = %v", err)
	}
	if metric.Confidence.Level != ConfidenceLevelNone {
		t.Fatalf("confidence level = %q, want %q", metric.Confidence.Level, ConfidenceLevelNone)
	}
	if metric.Confidence.Score != 0 {
		t.Fatalf("confidence score = %v, want 0", metric.Confidence.Score)
	}
	if !slices.Contains(metric.Confidence.Reasons, "latest_observation_is_in_future") {
		t.Fatalf("confidence reasons = %v", metric.Confidence.Reasons)
	}
}
