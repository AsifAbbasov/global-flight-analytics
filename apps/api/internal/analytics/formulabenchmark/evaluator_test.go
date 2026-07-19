package formulabenchmark

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/researchbenchmark"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/researchdataset"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectioncontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionevaluation"
)

func TestEvaluateProducesPassingNonCalibrationReport(t *testing.T) {
	request := validRequest()

	report, err := Evaluate(request)
	if err != nil {
		t.Fatalf("evaluate benchmark: %v", err)
	}
	if report.Status != StatusBenchmarkPassed {
		t.Fatalf(
			"status = %q, want %q",
			report.Status,
			StatusBenchmarkPassed,
		)
	}
	if report.CalibrationAllowed ||
		report.AutomaticFormulaChangesAllowed ||
		!report.ManualReviewRequired {
		t.Fatalf("unsafe calibration flags: %+v", report)
	}
	if report.MaximumClaim != MaximumClaim {
		t.Fatalf("maximum claim = %q", report.MaximumClaim)
	}
	if err := report.Validate(); err != nil {
		t.Fatalf("validate report: %v", err)
	}
}

func TestEvaluateReportsInsufficientEvidence(t *testing.T) {
	request := validRequest()
	request.ProjectionAggregate.EvaluationCount = 5
	method := &request.ProjectionAggregate.Methods[0]
	method.EvaluationCount = 5
	method.CompleteEvaluationCount = 5
	method.ForecastPointCount = 50
	method.EvaluatedPointCount = 50
	method.AltitudeEvaluatedPointCount = 30
	method.ArrivalEvaluationCount = 4

	report, err := Evaluate(request)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusInsufficientEvidence {
		t.Fatalf(
			"status = %q, want %q",
			report.Status,
			StatusInsufficientEvidence,
		)
	}
}

func TestEvaluateReportsThresholdFailure(t *testing.T) {
	request := validRequest()
	request.ProjectionAggregate.Methods[0].
		P95HorizontalErrorM = 100_000

	report, err := Evaluate(request)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != StatusBenchmarkFailed {
		t.Fatalf(
			"status = %q, want %q",
			report.Status,
			StatusBenchmarkFailed,
		)
	}
}

func TestEvaluateRejectsDatasetMismatch(t *testing.T) {
	request := validRequest()
	request.Manifest.DatasetID =
		researchdataset.IDClimbingAircraft

	_, err := Evaluate(request)
	if !errors.Is(err, ErrDatasetMismatch) {
		t.Fatalf(
			"error = %v, want %v",
			err,
			ErrDatasetMismatch,
		)
	}
}

func validRequest() Request {
	generatedAt := time.Date(
		2026,
		time.July,
		19,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	return Request{
		PlanID: researchbenchmark.
			ProjectionFormulaEvaluationPlanID,
		Manifest: researchdataset.Manifest{
			DatasetID: researchdataset.
				IDWeeklyStateVectors,
			Version: "bounded-test-v1",
			Files: []researchdataset.File{
				{
					Name:      "regional-sample.avro",
					Format:    "avro",
					SizeBytes: 1_024,
					SHA256: "sha256:" +
						strings.Repeat("a", 64),
				},
			},
			TotalBytes:           1_024,
			MaximumRecords:       1_000,
			RegionFilter:         "AZ,GE,AM,TR",
			OfflineOnly:          true,
			ProductionDependency: false,
			LicenseReviewed:      true,
			AttributionProvided:  true,
			PreparedAt:           generatedAt.Add(-time.Hour),
		},
		ProjectionAggregate: validAggregate(generatedAt),
		Policy:              DefaultPolicy(),
		GeneratedAt:         generatedAt,
	}
}

func validAggregate(
	generatedAt time.Time,
) projectionevaluation.AggregateResult {
	return projectionevaluation.AggregateResult{
		Version: projectionevaluation.AggregateVersion,
		Status:  projectionevaluation.StatusComplete,

		EvaluationCount: 50,
		MethodCount:     1,

		Methods: []projectionevaluation.MethodSummary{
			{
				MethodName:    "bounded-test-method",
				MethodVersion: "v1",
				DecisionClass: projectioncontract.
					DecisionClassProjectDerived,

				EvaluationCount:         50,
				CompleteEvaluationCount: 50,

				ForecastPointCount:  500,
				EvaluatedPointCount: 500,
				PointCoverageRatio:  1,

				MeanHorizontalErrorM:               8_000,
				MedianHorizontalErrorM:             6_000,
				P95HorizontalErrorM:                20_000,
				HorizontalRMSEM:                    10_000,
				HorizontalUncertaintyCoverageRatio: 0.90,

				AltitudeEvaluatedPointCount:      300,
				MeanAltitudeAbsoluteErrorM:       500,
				AltitudeRMSEM:                    700,
				VerticalUncertaintyCoverageRatio: 0.80,

				ArrivalEvaluationCount:          40,
				MeanArrivalAbsoluteErrorSeconds: 300,
				ArrivalIntervalCoverageRatio:    0.80,
			},
		},

		InputFingerprint: "sha256:" +
			strings.Repeat("b", 64),
		GeneratedAt: generatedAt.Add(-time.Minute),
	}
}
