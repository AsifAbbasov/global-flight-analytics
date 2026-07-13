package metricexecution

import (
	"context"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/analyticalresult"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/snapshot"
)

func TestCoverageScorePreservesExistingMetricFormula(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	execution, err := service.CoverageScore(
		context.Background(),
		CoverageScoreRequest{
			Snapshot: snapshot.Snapshot{
				ObservedSamples: 75,
				ExpectedSamples: 100,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected coverage score execution, got %v",
			err,
		)
	}

	if execution.Result.Value != 0.75 ||
		execution.Result.Status !=
			analyticalresult.StatusComplete ||
		execution.Result.Confidence.Level !=
			analyticalresult.ConfidenceLevelHigh {
		t.Fatalf(
			"unexpected coverage score result: %#v",
			execution.Result,
		)
	}

	if execution.Result.Eligibility != nil {
		t.Fatal(
			"expected no trajectory eligibility for snapshot metric",
		)
	}
}

func TestCoverageScoreMapsInvalidSnapshotToFailedResult(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	execution, err := service.CoverageScore(
		context.Background(),
		CoverageScoreRequest{},
	)
	if err != nil {
		t.Fatalf(
			"expected typed failed result, got %v",
			err,
		)
	}

	if !execution.IsFailed() ||
		execution.Result.Failure == nil {
		t.Fatalf(
			"expected failed coverage result, got %#v",
			execution.Result,
		)
	}
}

func TestDataFreshnessCanBeZeroWithHighCalculationConfidence(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)

	execution, err := service.DataFreshness(
		context.Background(),
		DataFreshnessRequest{
			Snapshot: snapshot.Snapshot{
				Time: metricTestTime().
					Add(-10 * time.Minute),
			},
			MaxAge: 2 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf(
			"expected data freshness execution, got %v",
			err,
		)
	}

	if execution.Result.Value != 0 {
		t.Fatalf(
			"expected stale freshness value zero, got %f",
			execution.Result.Value,
		)
	}

	if execution.Result.Status !=
		analyticalresult.StatusComplete ||
		execution.Result.Confidence.Level !=
			analyticalresult.ConfidenceLevelHigh {
		t.Fatalf(
			"expected complete high-confidence stale result, got %#v",
			execution.Result,
		)
	}
}

func TestCanceledContextProducesTypedFailedMetric(
	t *testing.T,
) {
	service := metricTestService(
		t,
		allowUnlessDeniedICAO,
	)
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	execution, err := service.CoverageScore(
		ctx,
		CoverageScoreRequest{
			Snapshot: snapshot.Snapshot{
				ObservedSamples: 1,
				ExpectedSamples: 1,
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"expected typed canceled result, got %v",
			err,
		)
	}

	if execution.Result.Failure == nil ||
		execution.Result.Failure.Code !=
			"analytical_operation_canceled" ||
		!execution.Result.Failure.Retriable {
		t.Fatalf(
			"unexpected canceled result: %#v",
			execution.Result,
		)
	}
}
