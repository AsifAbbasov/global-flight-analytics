package dataqualitycontract

import (
	"errors"
	"testing"
	"time"
)

func TestReportBuildsCompleteExplainableContract(t *testing.T) {
	evaluatedAt := time.Date(2026, time.July, 15, 1, 0, 0, 0, time.UTC)
	freshness, err := EvaluateFreshness(FreshnessInput{
		ObservedAt:       evaluatedAt.Add(-30 * time.Second),
		EvaluatedAt:      evaluatedAt,
		ExpectedInterval: time.Minute,
		StaleAfter:       5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("freshness: %v", err)
	}
	density, err := EvaluateSamplingDensity(SamplingDensityInput{
		WindowStart:      evaluatedAt.Add(-2 * time.Minute),
		WindowEnd:        evaluatedAt,
		ExpectedInterval: time.Minute,
		ObservationTimes: []time.Time{
			evaluatedAt.Add(-90 * time.Second),
			evaluatedAt.Add(-30 * time.Second),
		},
	})
	if err != nil {
		t.Fatalf("density: %v", err)
	}
	denied, err := DeniedPermission("insufficient_historical_depth")
	if err != nil {
		t.Fatalf("denied permission: %v", err)
	}
	permissions := AnalyticsPermissions{
		RouteInference:       AllowedPermission(),
		PhaseDetection:       AllowedPermission(),
		HistoricalAnalytics:  AllowedPermission(),
		HistoricalSimilarity: denied,
		Projection:           denied,
	}
	report, err := NewReport(
		validProvenance(evaluatedAt.Add(-30*time.Second)),
		freshness,
		density,
		permissions,
		[]string{"vertical_rate"},
		[]Notice{{Code: "partial_field_set", Message: "Vertical rate is unavailable."}},
		[]Notice{{Code: "open_data_only", Message: "Results use open-data observations."}},
		evaluatedAt,
	)
	if err != nil {
		t.Fatalf("new report: %v", err)
	}
	if report.ContractVersion != ContractVersion {
		t.Fatalf("unexpected contract version %q", report.ContractVersion)
	}
	clone := report.Clone()
	clone.Permissions.Projection.Reasons[0] = "mutated"
	clone.MissingFields[0] = "mutated"
	if report.Permissions.Projection.Reasons[0] != "insufficient_historical_depth" ||
		report.MissingFields[0] != "vertical_rate" {
		t.Fatal("expected report-owned slices")
	}
}

func TestReportRejectsMismatchedEvaluationTime(t *testing.T) {
	evaluatedAt := time.Now().UTC().Truncate(time.Second)
	freshness, err := EvaluateFreshness(FreshnessInput{
		ObservedAt:       evaluatedAt.Add(-time.Second),
		EvaluatedAt:      evaluatedAt,
		ExpectedInterval: time.Minute,
		StaleAfter:       5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("freshness: %v", err)
	}
	density, err := EvaluateSamplingDensity(SamplingDensityInput{
		WindowStart:      evaluatedAt.Add(-time.Minute),
		WindowEnd:        evaluatedAt,
		ExpectedInterval: time.Minute,
		ObservationTimes: []time.Time{evaluatedAt.Add(-time.Second)},
	})
	if err != nil {
		t.Fatalf("density: %v", err)
	}
	_, err = NewReport(
		validProvenance(evaluatedAt.Add(-time.Second)),
		freshness,
		density,
		allAllowedPermissions(),
		nil,
		nil,
		nil,
		evaluatedAt.Add(time.Second),
	)
	if !errors.Is(err, ErrEvaluatedAtMismatch) {
		t.Fatalf("expected evaluated-at mismatch, got %v", err)
	}
}

func validProvenance(sourceTime time.Time) Provenance {
	return Provenance{
		SourceName:       "airplanes.live",
		SourceRecordTime: sourceTime,
		ReceivedAt:       sourceTime.Add(time.Second),
		IngestionRunID:   "run-1",
		Transformation:   "flight-state-normalization",
		AlgorithmVersion: "normalizer-v1",
		InputFingerprint: "sha256:test",
	}
}

func allAllowedPermissions() AnalyticsPermissions {
	return AnalyticsPermissions{
		RouteInference:       AllowedPermission(),
		PhaseDetection:       AllowedPermission(),
		HistoricalAnalytics:  AllowedPermission(),
		HistoricalSimilarity: AllowedPermission(),
		Projection:           AllowedPermission(),
	}
}
