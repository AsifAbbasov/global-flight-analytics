package historicalseries

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func TestBuildCompleteSeries(
	t *testing.T,
) {
	plan := seriesTestPlan()
	result, err := Build(
		BuildRequest{
			Metric: historicalcontract.Metric{
				Name: historicalcontract.
					MetricNameObservationCount,
				Unit: "observations",
				Aggregation: historicalcontract.
					AggregationCount,
			},
			Scope: historicalcontract.Scope{
				Type: historicalcontract.ScopeTypeGlobal,
			},
			Plan: plan,
			Values: []BucketValue{
				{
					Bucket:      plan.Buckets[0],
					Value:       2,
					SampleCount: 2,
				},
				{
					Bucket:      plan.Buckets[1],
					Value:       3,
					SampleCount: 3,
				},
			},
			DataCoverageRatio: 1,
			BuilderVersion:    Version,
			InputFingerprint: "sha256:" +
				strings.Repeat("a", 64),
			SourceNames: []string{
				"flight_states",
			},
			LatestSourceUpdatedAt: plan.EffectiveWindow.EndTime.
				Add(-time.Minute),
			GeneratedAt: plan.AsOfTime,
		},
	)
	if err != nil {
		t.Fatalf("build complete series: %v", err)
	}

	if result.Status !=
		historicalcontract.SeriesStatusComplete {
		t.Fatalf(
			"expected complete status, got %s",
			result.Status,
		)
	}
	if result.Summary.Total != 5 {
		t.Fatalf(
			"expected total 5, got %f",
			result.Summary.Total,
		)
	}
	if result.Confidence.SampleCount != 5 {
		t.Fatalf(
			"expected five samples, got %d",
			result.Confidence.SampleCount,
		)
	}
}

func TestBuildPartialSeries(
	t *testing.T,
) {
	plan := seriesTestPlan()
	result, err := Build(
		BuildRequest{
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
			Plan: plan,
			Values: []BucketValue{
				{
					Bucket:      plan.Buckets[0],
					Value:       1,
					SampleCount: 1,
				},
				{
					Bucket:      plan.Buckets[1],
					Value:       0,
					SampleCount: 0,
				},
			},
			DataCoverageRatio: 0.5,
			BuilderVersion:    Version,
			InputFingerprint: "sha256:" +
				strings.Repeat("b", 64),
			SourceNames: []string{"flights"},
			GeneratedAt: plan.AsOfTime,
		},
	)
	if err != nil {
		t.Fatalf("build partial series: %v", err)
	}

	if result.Status !=
		historicalcontract.SeriesStatusPartial {
		t.Fatalf(
			"expected partial status, got %s",
			result.Status,
		)
	}
	for _, point := range result.Points {
		if point.Status !=
			historicalcontract.BucketStatusPartial {
			t.Fatalf(
				"expected partial point, got %s",
				point.Status,
			)
		}
	}
}

func TestBuildRejectsMismatchedBucketOrder(
	t *testing.T,
) {
	plan := seriesTestPlan()
	_, err := Build(
		BuildRequest{
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
			Plan: plan,
			Values: []BucketValue{
				{Bucket: plan.Buckets[1]},
				{Bucket: plan.Buckets[0]},
			},
			DataCoverageRatio: 1,
			BuilderVersion:    Version,
			InputFingerprint: "sha256:" +
				strings.Repeat("c", 64),
			SourceNames: []string{"flights"},
			GeneratedAt: plan.AsOfTime,
		},
	)
	if !errors.Is(
		err,
		ErrBucketValueOrderInvalid,
	) {
		t.Fatalf(
			"expected bucket order error, got %v",
			err,
		)
	}
}

func seriesTestPlan() historicalwindow.Plan {
	start := time.Date(
		2026,
		time.July,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	end := start.Add(2 * time.Hour)
	window := historicalcontract.TimeWindow{
		StartTime: start,
		EndTime:   end,
		AsOfTime:  end,
	}

	return historicalwindow.Plan{
		Version:            historicalwindow.Version,
		Fingerprint:        "test-plan",
		RequestedStartTime: start,
		RequestedEndTime:   end,
		AsOfTime:           end,
		Granularity: historicalcontract.
			GranularityHour,
		EffectiveWindow: &window,
		Buckets: []historicalwindow.Bucket{
			{
				Key:       "bucket-0",
				Sequence:  0,
				StartTime: start,
				EndTime:   start.Add(time.Hour),
			},
			{
				Key:       "bucket-1",
				Sequence:  1,
				StartTime: start.Add(time.Hour),
				EndTime:   end,
			},
		},
		MaximumBucketCount: 100,
	}
}
