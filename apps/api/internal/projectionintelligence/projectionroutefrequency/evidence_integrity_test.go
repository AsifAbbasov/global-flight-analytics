package projectionroutefrequency

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
)

func TestHistorySummaryValidateRejectsImpossibleEvidenceCounts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*HistorySummary)
	}{
		{
			name: "distinct days exceed observations",
			mutate: func(summary *HistorySummary) {
				summary.DistinctDayCount = summary.ObservationCount + 1
			},
		},
		{
			name: "distinct days exceed distinct flights",
			mutate: func(summary *HistorySummary) {
				summary.DistinctDayCount = summary.DistinctFlightCount + 1
			},
		},
		{
			name: "recent observations exceed distinct flights",
			mutate: func(summary *HistorySummary) {
				summary.RecentObservationCount = summary.DistinctFlightCount + 1
			},
		},
		{
			name: "zero observations retain distinct days",
			mutate: func(summary *HistorySummary) {
				summary.ObservationCount = 0
				summary.DistinctFlightCount = 0
				summary.DistinctDayCount = 1
				summary.RecentObservationCount = 0
				summary.LastObservedAt = time.Time{}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			summary := validRouteHistory()
			test.mutate(&summary)
			if err := summary.Validate(); err == nil {
				t.Fatal("Validate() accepted impossible route-history counts")
			}
		})
	}
}

func TestHistorySummaryValidateRejectsMissingSourceAndNonCanonicalRouteKey(t *testing.T) {
	missingSource := validRouteHistory()
	missingSource.SourceNames = nil
	if err := missingSource.Validate(); err == nil {
		t.Fatal("Validate() accepted route history without a provenance source")
	}

	nonCanonicalRoute := validRouteHistory()
	nonCanonicalRoute.RouteKey = "ubbb>ltba"
	if err := nonCanonicalRoute.Validate(); err == nil {
		t.Fatal("Validate() accepted a non-canonical route key")
	}
}

func TestEvaluateRejectsRecentWindowMismatch(t *testing.T) {
	evaluator := newRouteFrequencyEvaluator(t)
	history := validRouteHistory()
	history.RecentWindowStart = history.RecentWindowStart.Add(time.Minute)

	_, err := evaluator.Evaluate(validRouteFrequencyRoute(), history)
	if !errors.Is(err, ErrRouteHistoryRecentWindowMismatch) {
		t.Fatalf("error = %v, want %v", err, ErrRouteHistoryRecentWindowMismatch)
	}
}

func TestRouteFrequencyFingerprintBindsDecisionRelevantRouteFields(t *testing.T) {
	evaluator := newRouteFrequencyEvaluator(t)
	route := validRouteFrequencyRoute()
	history := validRouteHistory()

	baseline, err := evaluator.Evaluate(route, history)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	mutatedConfidence := route.Clone()
	mutatedConfidence.Confidence.Score = 0.8
	mutatedConfidence.Confidence.Level = validRouteFrequencyConfidence(0.8, "route", 2).Level
	mutatedConfidence.Confidence.Reasons[0].Contribution = 0.8
	confidenceResult, err := evaluator.Evaluate(mutatedConfidence, history)
	if err != nil {
		t.Fatalf("confidence mutation Evaluate() error = %v", err)
	}
	if confidenceResult.InputFingerprint == baseline.InputFingerprint {
		t.Fatal("route-confidence mutation did not change the route-frequency fingerprint")
	}

	mutatedStatus := route.Clone()
	mutatedStatus.Status = routecontract.RouteStatusPartial
	mutatedStatus.Origin = nil
	mutatedStatus.Confidence.EvidenceCount = 1
	statusResult, err := evaluator.Evaluate(mutatedStatus, history)
	if err != nil {
		t.Fatalf("status mutation Evaluate() error = %v", err)
	}
	if statusResult.InputFingerprint == baseline.InputFingerprint {
		t.Fatal("route-status mutation did not change the route-frequency fingerprint")
	}
}
