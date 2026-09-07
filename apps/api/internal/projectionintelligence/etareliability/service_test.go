package etareliability

import (
	"math"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
)

func TestSelectHistoricalAsOfUsesPersistedObservationNearComparableLead(t *testing.T) {
	base := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	item := trajectory.FlightTrajectory{Points: make([]trajectory.TrackPoint4D, 0, 10)}
	for index := 0; index < 10; index++ {
		item.Points = append(item.Points, trajectory.TrackPoint4D{ID: string(rune('a'+index)), ObservedAt: base.Add(time.Duration(index)*10*time.Minute)})
	}
	endpoint := item.Points[len(item.Points)-1].ObservedAt
	selected, ok := selectHistoricalAsOf(item, endpoint, 30*time.Minute, DefaultPolicy())
	if !ok {
		t.Fatal("expected comparable persisted observation")
	}
	want := endpoint.Add(-30*time.Minute)
	if !selected.Equal(want) {
		t.Fatalf("selected = %s, want %s", selected, want)
	}
}

func TestBuildMetricsPublishesMedianP80AndCoverage(t *testing.T) {
	metrics := buildMetrics([]sample{
		{AbsoluteErrorSeconds: 60, IntervalCovered: true},
		{AbsoluteErrorSeconds: 180, IntervalCovered: true},
		{AbsoluteErrorSeconds: 360, IntervalCovered: false},
		{AbsoluteErrorSeconds: 600, IntervalCovered: true},
		{AbsoluteErrorSeconds: 900, IntervalCovered: false},
	})
	if metrics.SampleCount != 5 || metrics.MedianAbsoluteErrorSeconds != 360 || metrics.P80AbsoluteErrorSeconds != 600 {
		t.Fatalf("unexpected error distribution: %#v", metrics)
	}
	if math.Abs(metrics.WithinFiveMinutesRatio-0.4) > 1e-9 || math.Abs(metrics.WithinTenMinutesRatio-0.8) > 1e-9 || math.Abs(metrics.IntervalCoverageRatio-0.6) > 1e-9 {
		t.Fatalf("unexpected reliability ratios: %#v", metrics)
	}
}

func TestObservedEndpointRejectsMissingPointsAndUsesLatestObservation(t *testing.T) {
	if _, ok := observedEndpoint(trajectory.FlightTrajectory{}); ok {
		t.Fatal("empty trajectory must not publish an endpoint proxy")
	}
	base := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	item := trajectory.FlightTrajectory{Points: []trajectory.TrackPoint4D{
		{ID: "old", Latitude: 40, Longitude: 49, ObservedAt: base},
		{ID: "new", Latitude: 41, Longitude: 50, ObservedAt: base.Add(time.Minute)},
	}}
	endpoint, ok := observedEndpoint(item)
	if !ok || endpoint.ID != "new" {
		t.Fatalf("endpoint = %#v, ok=%v", endpoint, ok)
	}
}

func TestDefaultPolicyIsZeroCostBounded(t *testing.T) {
	policy := DefaultPolicy()
	if err := policy.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if policy.MaximumCandidateCount != 8 || policy.CompleteSampleCount > policy.MaximumCandidateCount {
		t.Fatalf("unexpected bounded policy: %#v", policy)
	}
}
