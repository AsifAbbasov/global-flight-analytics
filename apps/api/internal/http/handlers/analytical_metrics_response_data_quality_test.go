package handlers

import (
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/dataqualitycontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
)

func TestAnalyticalMetricResponseIncludesDataQuality(
	t *testing.T,
) {
	evaluatedAt := time.Date(
		2026,
		time.July,
		15,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	report := handlerDataQualityReport(
		t,
		evaluatedAt,
	)

	response := toAnalyticalMetricResponse(
		metricexecution.Execution[int]{
			MetricID: "traffic.active_aircraft",
			Result: analyticalresult.Result[int]{
				Status:   analyticalresult.StatusComplete,
				Value:    1,
				HasValue: true,
				Confidence: analyticalresult.Confidence{
					Level: analyticalresult.ConfidenceLevelHigh,
					Score: 1,
				},
				DataQuality:  &report,
				CalculatedAt: evaluatedAt,
			},
			Scope: metricexecution.ScopeSummary{
				EvaluatedAt: evaluatedAt,
			},
		},
	)

	if response.DataQuality == nil {
		t.Fatal("expected data quality in HTTP response")
	}
	if response.DataQuality.ContractVersion !=
		dataqualitycontract.ContractVersion {
		t.Fatalf(
			"expected contract version %q, got %q",
			dataqualitycontract.ContractVersion,
			response.DataQuality.ContractVersion,
		)
	}

	response.DataQuality.MissingFields[0] = "mutated"
	if report.MissingFields[0] != "callsign" {
		t.Fatal("expected HTTP data-quality response to be copied")
	}
}

func handlerDataQualityReport(
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
		t.Fatalf("build sampling density: %v", err)
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
