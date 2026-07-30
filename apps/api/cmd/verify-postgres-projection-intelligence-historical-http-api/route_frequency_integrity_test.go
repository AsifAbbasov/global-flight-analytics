package main

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionroutefrequency"
)

func TestValidateIndependentRouteFrequencyEvidence(t *testing.T) {
	result := projectionproduction.Result{
		RouteFrequency: &projectionroutefrequency.Result{
			ObservationCount:       5,
			DistinctFlightCount:    5,
			DistinctDayCount:       5,
			RecentObservationCount: 5,
			LatestObservationAge:   24 * time.Hour,
		},
	}
	if err := validateIndependentRouteFrequencyEvidence(result, 5); err != nil {
		t.Fatalf("validateIndependentRouteFrequencyEvidence() error = %v", err)
	}

	result.RouteFrequency.ObservationCount = 6
	if err := validateIndependentRouteFrequencyEvidence(result, 5); err == nil {
		t.Fatal("validator accepted current-flight leakage")
	}
}
