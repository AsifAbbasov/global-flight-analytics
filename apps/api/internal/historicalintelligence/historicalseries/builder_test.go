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
	values, err := BindDatasetCoverage(
		[]BucketValue{
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
		DatasetCoverage{
			State:        DatasetReadComplete,
			MatchedCount: 5,
		},
	)
	if err != nil {
		t.Fatalf("bind complete coverage: %v", err)
	}

	result, err := Build(
		validSeriesRequest(plan, values),
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
	if result.Confidence.Score != 1 ||
		result.Confidence.SampleCount != 5 {
		t.Fatalf(
			"unexpected confidence: %#v",
			result.Confidence,
		)
	}
}

func TestBuildUsesBucketSpecificIncompleteCoverage(
	t *testing.T,
) {
	plan := seriesTestPlan()
	values, err := BindDatasetCoverage(
		[]BucketValue{
			{
				Bucket:      plan.Buckets[0],
				Value:       1,
				SampleCount: 1,
			},
			{
				Bucket: plan.Buckets[1],
			},
		},
		DatasetCoverage{
			State:        DatasetReadIncomplete,
			MatchedCount: 4,
		},
	)
	if err != nil {
		t.Fatalf("bind incomplete coverage: %v", err)
	}

	request := validSeriesRequest(plan, values)
	request.Metric = historicalcontract.Metric{
		Name: historicalcontract.MetricNameFlightCount,
		Unit: "flights",
		Aggregation: historicalcontract.
			AggregationCount,
	}
	request.SourceNames = []string{"flights"}

	result, err := Build(request)
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
	if len(result.Points) != 2 {
		t.Fatalf(
			"point count = %d, want 2",
			len(result.Points),
		)
	}
	if result.Points[0].Status !=
		historicalcontract.BucketStatusPartial ||
		result.Points[0].CoverageRatio != 0.25 {
		t.Fatalf(
			"first point = %#v",
			result.Points[0],
		)
	}
	if result.Points[1].Status !=
		historicalcontract.BucketStatusUnavailable ||
		result.Points[1].CoverageRatio != 0 {
		t.Fatalf(
			"second point = %#v",
			result.Points[1],
		)
	}
	if result.Confidence.Score != 0.125 {
		t.Fatalf(
			"series confidence = %f, want 0.125",
			result.Confidence.Score,
		)
	}
}

func TestBuildZeroRepresentedCoverageIsUnavailable(
	t *testing.T,
) {
	plan := seriesTestPlan()
	values, err := BindDatasetCoverage(
		[]BucketValue{
			{Bucket: plan.Buckets[0]},
			{Bucket: plan.Buckets[1]},
		},
		DatasetCoverage{
			State:        DatasetReadIncomplete,
			MatchedCount: 2,
		},
	)
	if err != nil {
		t.Fatalf("bind unavailable coverage: %v", err)
	}

	request := validSeriesRequest(plan, values)
	request.Metric = historicalcontract.Metric{
		Name: historicalcontract.MetricNameFlightCount,
		Unit: "flights",
		Aggregation: historicalcontract.
			AggregationCount,
	}
	request.SourceNames = []string{"flights"}

	result, err := Build(request)
	if err != nil {
		t.Fatalf("build unavailable series: %v", err)
	}
	if result.Status !=
		historicalcontract.SeriesStatusUnavailable {
		t.Fatalf(
			"status = %s, want unavailable",
			result.Status,
		)
	}
	if len(result.Points) != 0 ||
		result.Confidence.Score != 0 {
		t.Fatalf(
			"unexpected unavailable result: %#v",
			result,
		)
	}
}

func TestBuildCompleteEmptyBucketsRemainComplete(
	t *testing.T,
) {
	plan := seriesTestPlan()
	values, err := BindDatasetCoverage(
		[]BucketValue{
			{Bucket: plan.Buckets[0]},
			{Bucket: plan.Buckets[1]},
		},
		DatasetCoverage{
			State:        DatasetReadComplete,
			MatchedCount: 0,
		},
	)
	if err != nil {
		t.Fatalf("bind empty complete coverage: %v", err)
	}

	result, err := Build(
		validSeriesRequest(plan, values),
	)
	if err != nil {
		t.Fatalf("build complete empty series: %v", err)
	}
	if result.Status !=
		historicalcontract.SeriesStatusComplete ||
		result.Confidence.Score != 1 ||
		result.Confidence.SampleCount != 0 {
		t.Fatalf(
			"unexpected complete empty result: %#v",
			result,
		)
	}
}

func TestBuildRequiresRealProvenanceTimestamps(
	t *testing.T,
) {
	plan := seriesTestPlan()
	values := completeValues(plan)
	request := validSeriesRequest(plan, values)

	request.LatestSourceUpdatedAt = time.Time{}
	_, err := Build(request)
	if !errors.Is(
		err,
		ErrLatestSourceTimeRequired,
	) {
		t.Fatalf(
			"latest source error = %v",
			err,
		)
	}

	request = validSeriesRequest(plan, values)
	request.GeneratedAt = time.Time{}
	_, err = Build(request)
	if !errors.Is(
		err,
		ErrGeneratedAtRequired,
	) {
		t.Fatalf(
			"generated time error = %v",
			err,
		)
	}
}

func TestBuildRejectsMalformedAndDuplicateLimitations(
	t *testing.T,
) {
	plan := seriesTestPlan()
	request := validSeriesRequest(
		plan,
		completeValues(plan),
	)
	request.Limitations = []historicalcontract.Limitation{
		{
			Code:    "",
			Message: "missing code",
			Scope:   "series",
		},
	}
	_, err := Build(request)
	if !errors.Is(err, ErrLimitationInvalid) {
		t.Fatalf(
			"malformed limitation error = %v",
			err,
		)
	}

	request = validSeriesRequest(
		plan,
		completeValues(plan),
	)
	request.Limitations = []historicalcontract.Limitation{
		{
			Code:    "same",
			Message: "first",
			Scope:   "series",
		},
		{
			Code:    "same",
			Message: "second",
			Scope:   "series",
		},
	}
	_, err = Build(request)
	if !errors.Is(err, ErrLimitationDuplicate) {
		t.Fatalf(
			"duplicate limitation error = %v",
			err,
		)
	}
}

func TestBuildRejectsSampleCountOverflow(
	t *testing.T,
) {
	plan := seriesTestPlan()
	maximum := int(^uint(0) >> 1)
	values, err := BindDatasetCoverage(
		[]BucketValue{
			{
				Bucket:      plan.Buckets[0],
				SampleCount: maximum,
			},
			{
				Bucket:      plan.Buckets[1],
				SampleCount: 1,
			},
		},
		DatasetCoverage{
			State: DatasetReadComplete,
		},
	)
	if err != nil {
		t.Fatalf("bind overflow coverage: %v", err)
	}

	_, err = Build(
		validSeriesRequest(plan, values),
	)
	if !errors.Is(err, ErrSampleCountOverflow) {
		t.Fatalf(
			"overflow error = %v",
			err,
		)
	}
}

func TestBuildRejectsMismatchedBucketOrder(
	t *testing.T,
) {
	plan := seriesTestPlan()
	values := completeValues(plan)
	values[0].Bucket = plan.Buckets[1]
	values[1].Bucket = plan.Buckets[0]

	_, err := Build(
		validSeriesRequest(plan, values),
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

func validSeriesRequest(
	plan historicalwindow.Plan,
	values []BucketValue,
) BuildRequest {
	return BuildRequest{
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
		Plan:             plan,
		Values:           values,
		BuilderVersion:   Version,
		InputFingerprint: "sha256:" + strings.Repeat("a", 64),
		SourceNames:      []string{"flight_states"},
		LatestSourceUpdatedAt: plan.EffectiveWindow.
			EndTime.Add(-time.Minute),
		GeneratedAt: plan.AsOfTime,
	}
}

func completeValues(
	plan historicalwindow.Plan,
) []BucketValue {
	values, err := BindDatasetCoverage(
		[]BucketValue{
			{
				Bucket:      plan.Buckets[0],
				Value:       1,
				SampleCount: 1,
			},
			{
				Bucket:      plan.Buckets[1],
				Value:       1,
				SampleCount: 1,
			},
		},
		DatasetCoverage{
			State:        DatasetReadComplete,
			MatchedCount: 2,
		},
	)
	if err != nil {
		panic(err)
	}
	return values
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
