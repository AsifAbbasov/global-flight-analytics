package analyticalresult

import (
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
)

func TestWithDataQualityAttachesDefensiveCopy(
	t *testing.T,
) {
	calculatedAt := analyticalResultTestTime()
	eligibility := allowedEligibility(calculatedAt)
	result, err := NewComplete(
		42,
		highConfidence(),
		&eligibility,
		validSources(calculatedAt),
		calculatedAt,
	)
	if err != nil {
		t.Fatalf("build analytical result: %v", err)
	}

	report := analyticalResultDataQualityReport(
		t,
		calculatedAt,
	)
	updated, err := result.WithDataQuality(&report)
	if err != nil {
		t.Fatalf("attach data quality: %v", err)
	}
	if updated.DataQuality == nil {
		t.Fatal("expected data-quality report")
	}

	updated.DataQuality.MissingFields[0] = "mutated"
	if report.MissingFields[0] != "callsign" {
		t.Fatal("expected attached report to be copied")
	}

	clone := updated.Clone()
	clone.DataQuality.Permissions.RouteInference.Reasons = append(
		clone.DataQuality.Permissions.RouteInference.Reasons,
		"mutated",
	)
	if len(
		updated.DataQuality.Permissions.RouteInference.Reasons,
	) != 0 {
		t.Fatal("expected cloned report permissions to be copied")
	}
}

func TestWithDataQualityRejectsFutureEvaluation(
	t *testing.T,
) {
	calculatedAt := analyticalResultTestTime()
	eligibility := allowedEligibility(calculatedAt)
	result, err := NewComplete(
		42,
		highConfidence(),
		&eligibility,
		validSources(calculatedAt),
		calculatedAt,
	)
	if err != nil {
		t.Fatalf("build analytical result: %v", err)
	}

	report := analyticalResultDataQualityReport(
		t,
		calculatedAt.Add(time.Second),
	)
	_, err = result.WithDataQuality(&report)
	if !errors.Is(
		err,
		ErrDataQualityEvaluationAfterCalculation,
	) {
		t.Fatalf(
			"expected future data-quality evaluation error, got %v",
			err,
		)
	}
}

func analyticalResultDataQualityReport(
	t *testing.T,
	evaluatedAt time.Time,
) dataqualitycontract.Report {
	t.Helper()

	freshness, err := dataqualitycontract.EvaluateFreshness(
		dataqualitycontract.FreshnessInput{
			ObservedAt:       evaluatedAt.Add(-5 * time.Second),
			EvaluatedAt:      evaluatedAt,
			ExpectedInterval: 10 * time.Second,
			StaleAfter:       time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("build freshness: %v", err)
	}

	density, err := dataqualitycontract.EvaluateSamplingDensity(
		dataqualitycontract.SamplingDensityInput{
			WindowStart:      evaluatedAt.Add(-20 * time.Second),
			WindowEnd:        evaluatedAt,
			ExpectedInterval: 10 * time.Second,
			ObservationTimes: []time.Time{
				evaluatedAt.Add(-15 * time.Second),
				evaluatedAt.Add(-5 * time.Second),
			},
		},
	)
	if err != nil {
		t.Fatalf("build density: %v", err)
	}

	allowed := dataqualitycontract.AllowedPermission()
	report, err := dataqualitycontract.NewReport(
		dataqualitycontract.Provenance{
			SourceName:       "airplanes.live",
			SourceRecordTime: evaluatedAt.Add(-5 * time.Second),
			ReceivedAt:       evaluatedAt.Add(-4 * time.Second),
			IngestionRunID:   "run-1",
			Transformation:   "metric",
			AlgorithmVersion: dataqualitycontract.ContractVersion,
			InputFingerprint: "fingerprint",
		},
		freshness,
		density,
		dataqualitycontract.AnalyticsPermissions{
			RouteInference:       allowed,
			PhaseDetection:       allowed,
			HistoricalAnalytics:  allowed,
			HistoricalSimilarity: allowed,
			Projection:           allowed,
		},
		[]string{"callsign"},
		nil,
		nil,
		evaluatedAt,
	)
	if err != nil {
		t.Fatalf("build data-quality report: %v", err)
	}

	return report
}
