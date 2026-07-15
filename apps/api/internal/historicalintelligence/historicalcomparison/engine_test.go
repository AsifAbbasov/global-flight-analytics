package historicalcomparison

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalseries"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func TestAttachBuildsIncreaseComparison(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		asOfTime.Add(-4*time.Hour),
		asOfTime.Add(-2*time.Hour),
		asOfTime,
		[]float64{1, 2},
	)
	current := comparisonSeries(
		t,
		asOfTime.Add(-2*time.Hour),
		asOfTime,
		asOfTime,
		[]float64{3, 3},
	)

	result, err := Attach(current, previous)
	if err != nil {
		t.Fatalf("attach comparison: %v", err)
	}
	if result.Comparison == nil {
		t.Fatal("expected period comparison")
	}
	if result.Comparison.CurrentValue != 6 ||
		result.Comparison.PreviousValue != 3 ||
		result.Comparison.AbsoluteChange != 3 {
		t.Fatalf(
			"unexpected comparison values: %#v",
			result.Comparison,
		)
	}
	if result.Comparison.PercentageChange == nil ||
		*result.Comparison.PercentageChange != 100 {
		t.Fatalf(
			"expected 100 percent increase, got %#v",
			result.Comparison.PercentageChange,
		)
	}
	if result.Comparison.Direction !=
		historicalcontract.TrendDirectionUp {
		t.Fatalf(
			"expected upward trend, got %s",
			result.Comparison.Direction,
		)
	}
}

func TestAttachOmitsPercentageWhenPreviousValueIsZero(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		asOfTime.Add(-4*time.Hour),
		asOfTime.Add(-2*time.Hour),
		asOfTime,
		[]float64{0, 0},
	)
	current := comparisonSeries(
		t,
		asOfTime.Add(-2*time.Hour),
		asOfTime,
		asOfTime,
		[]float64{1, 0},
	)

	result, err := Attach(current, previous)
	if err != nil {
		t.Fatalf("attach zero-base comparison: %v", err)
	}
	if result.Comparison.PercentageChange != nil {
		t.Fatalf(
			"expected undefined percentage change, got %f",
			*result.Comparison.PercentageChange,
		)
	}
}

func TestAttachRejectsNonAdjacentWindows(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		asOfTime.Add(-5*time.Hour),
		asOfTime.Add(-3*time.Hour),
		asOfTime,
		[]float64{1, 2},
	)
	current := comparisonSeries(
		t,
		asOfTime.Add(-2*time.Hour),
		asOfTime,
		asOfTime,
		[]float64{3, 3},
	)

	_, err := Attach(current, previous)
	if !errors.Is(err, ErrWindowNotAdjacent) {
		t.Fatalf(
			"expected non-adjacent window error, got %v",
			err,
		)
	}
}

func TestAttachDoesNotShareComparisonState(
	t *testing.T,
) {
	asOfTime := comparisonTestTime()
	previous := comparisonSeries(
		t,
		asOfTime.Add(-4*time.Hour),
		asOfTime.Add(-2*time.Hour),
		asOfTime,
		[]float64{1, 2},
	)
	current := comparisonSeries(
		t,
		asOfTime.Add(-2*time.Hour),
		asOfTime,
		asOfTime,
		[]float64{3, 3},
	)

	result, err := Attach(current, previous)
	if err != nil {
		t.Fatalf("attach comparison: %v", err)
	}

	clone := result.Clone()
	*clone.Comparison.PercentageChange = 999
	if *result.Comparison.PercentageChange == 999 {
		t.Fatal(
			"expected comparison percentage pointer to be cloned",
		)
	}
}

func comparisonSeries(
	t *testing.T,
	startTime time.Time,
	endTime time.Time,
	asOfTime time.Time,
	values []float64,
) historicalcontract.Result {
	t.Helper()

	window := historicalcontract.TimeWindow{
		StartTime: startTime,
		EndTime:   endTime,
		AsOfTime:  asOfTime,
	}
	buckets := make(
		[]historicalwindow.Bucket,
		0,
		len(values),
	)
	bucketValues := make(
		[]historicalseries.BucketValue,
		0,
		len(values),
	)

	for index, value := range values {
		bucket := historicalwindow.Bucket{
			Key:      "bucket-" + string(rune('a'+index)),
			Sequence: index,
			StartTime: startTime.Add(
				time.Duration(index) * time.Hour,
			),
			EndTime: startTime.Add(
				time.Duration(index+1) * time.Hour,
			),
		}
		buckets = append(buckets, bucket)
		bucketValues = append(
			bucketValues,
			historicalseries.BucketValue{
				Bucket:      bucket,
				Value:       value,
				SampleCount: int(value),
			},
		)
	}

	result, err := historicalseries.Build(
		historicalseries.BuildRequest{
			Metric: historicalcontract.Metric{
				Name: historicalcontract.
					MetricNameFlightCount,
				Unit: "flights",
				Aggregation: historicalcontract.
					AggregationCount,
			},
			Scope: historicalcontract.Scope{
				Type: historicalcontract.ScopeTypeGlobal,
			},
			Plan: historicalwindow.Plan{
				Version: historicalwindow.Version,
				Fingerprint: "comparison-plan-" +
					startTime.Format(time.RFC3339),
				RequestedStartTime: startTime,
				RequestedEndTime:   endTime,
				AsOfTime:           asOfTime,
				Granularity: historicalcontract.
					GranularityHour,
				EffectiveWindow:    &window,
				Buckets:            buckets,
				MaximumBucketCount: 100,
			},
			Values:            bucketValues,
			DataCoverageRatio: 1,
			BuilderVersion:    Version,
			InputFingerprint: "sha256:" +
				strings.Repeat(
					string(rune('a'+len(values))),
					64,
				),
			SourceNames:           []string{"test"},
			LatestSourceUpdatedAt: endTime,
			GeneratedAt:           asOfTime,
		},
	)
	if err != nil {
		t.Fatalf("build comparison fixture: %v", err)
	}

	return result
}

func comparisonTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		2,
		0,
		0,
		0,
		0,
		time.UTC,
	)
}
