package historicalwindow

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
)

func TestBuildCreatesAlignedHourlyPlan(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		1,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(
		3 * time.Hour,
	)
	asOfTime := endTime.Add(time.Hour)

	plan := mustBuild(
		t,
		Request{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
			Granularity: historicalcontract.
				GranularityHour,
		},
	)

	if plan.Version != Version ||
		!plan.HasBuckets() ||
		len(plan.Buckets) != 3 ||
		len(plan.Exclusions) != 0 ||
		plan.TruncatedByAsOfTime ||
		plan.MaximumBucketCount !=
			DefaultMaximumBucketCount {
		t.Fatalf(
			"unexpected plan: %#v",
			plan,
		)
	}

	for index, bucket := range plan.Buckets {
		wantStart := startTime.Add(
			time.Duration(index) *
				time.Hour,
		)
		wantEnd := wantStart.Add(time.Hour)

		if bucket.Sequence != index+1 ||
			!bucket.StartTime.Equal(
				wantStart,
			) ||
			!bucket.EndTime.Equal(
				wantEnd,
			) ||
			bucket.Duration() != time.Hour {
			t.Fatalf(
				"unexpected bucket %d: %#v",
				index,
				bucket,
			)
		}
	}

	if plan.EffectiveWindow == nil ||
		!plan.EffectiveWindow.StartTime.
			Equal(startTime) ||
		!plan.EffectiveWindow.EndTime.
			Equal(endTime) ||
		!plan.EffectiveWindow.AsOfTime.
			Equal(asOfTime) {
		t.Fatalf(
			"unexpected effective window: %#v",
			plan.EffectiveWindow,
		)
	}

	wantPreviousStart := startTime.Add(
		-3 * time.Hour,
	)
	if plan.PreviousWindow == nil ||
		!plan.PreviousWindow.StartTime.
			Equal(wantPreviousStart) ||
		!plan.PreviousWindow.EndTime.
			Equal(startTime) ||
		!plan.PreviousWindow.AsOfTime.
			Equal(asOfTime) {
		t.Fatalf(
			"unexpected previous window: %#v",
			plan.PreviousWindow,
		)
	}
}

func TestBuildExcludesIncompleteEdgesAndFutureTime(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		1,
		10,
		15,
		0,
		0,
		time.UTC,
	)
	endTime := time.Date(
		2026,
		time.July,
		1,
		14,
		30,
		0,
		0,
		time.UTC,
	)
	asOfTime := time.Date(
		2026,
		time.July,
		1,
		13,
		40,
		0,
		0,
		time.UTC,
	)

	plan := mustBuild(
		t,
		Request{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
			Granularity: historicalcontract.
				GranularityHour,
		},
	)

	wantEffectiveStart := time.Date(
		2026,
		time.July,
		1,
		11,
		0,
		0,
		0,
		time.UTC,
	)
	wantEffectiveEnd := time.Date(
		2026,
		time.July,
		1,
		13,
		0,
		0,
		0,
		time.UTC,
	)

	if len(plan.Buckets) != 2 ||
		plan.EffectiveWindow == nil ||
		!plan.EffectiveWindow.StartTime.
			Equal(wantEffectiveStart) ||
		!plan.EffectiveWindow.EndTime.
			Equal(wantEffectiveEnd) ||
		!plan.TruncatedByAsOfTime {
		t.Fatalf(
			"unexpected plan: %#v",
			plan,
		)
	}

	wantExclusions := []Exclusion{
		{
			Reason:    ExclusionReasonLeadingIncompleteBucket,
			StartTime: startTime,
			EndTime:   wantEffectiveStart,
		},
		{
			Reason:    ExclusionReasonTrailingIncompleteBucket,
			StartTime: wantEffectiveEnd,
			EndTime:   asOfTime,
		},
		{
			Reason:    ExclusionReasonFutureAfterAsOfTime,
			StartTime: asOfTime,
			EndTime:   endTime,
		},
	}

	if !reflect.DeepEqual(
		plan.Exclusions,
		wantExclusions,
	) {
		t.Fatalf(
			"exclusions = %#v, want %#v",
			plan.Exclusions,
			wantExclusions,
		)
	}
}

func TestBuildReturnsNoCompleteBucketWithoutError(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		1,
		10,
		15,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(
		30 * time.Minute,
	)

	plan := mustBuild(
		t,
		Request{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  endTime.Add(time.Hour),
			Granularity: historicalcontract.
				GranularityHour,
		},
	)

	if plan.HasBuckets() ||
		plan.EffectiveWindow != nil ||
		plan.PreviousWindow != nil ||
		len(plan.Exclusions) != 1 ||
		plan.Exclusions[0].Reason !=
			ExclusionReasonNoCompleteBucket ||
		!plan.Exclusions[0].StartTime.
			Equal(startTime) ||
		!plan.Exclusions[0].EndTime.
			Equal(endTime) {
		t.Fatalf(
			"unexpected plan: %#v",
			plan,
		)
	}
}

func TestBuildTreatsEntireFutureRequestAsExcluded(
	t *testing.T,
) {
	asOfTime := time.Date(
		2026,
		time.July,
		1,
		10,
		0,
		0,
		0,
		time.UTC,
	)
	startTime := asOfTime.Add(time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	plan := mustBuild(
		t,
		Request{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
			Granularity: historicalcontract.
				GranularityHour,
		},
	)

	if plan.HasBuckets() ||
		!plan.TruncatedByAsOfTime ||
		len(plan.Exclusions) != 1 ||
		plan.Exclusions[0].Reason !=
			ExclusionReasonFutureAfterAsOfTime ||
		!plan.Exclusions[0].StartTime.
			Equal(startTime) ||
		!plan.Exclusions[0].EndTime.
			Equal(endTime) {
		t.Fatalf(
			"unexpected future plan: %#v",
			plan,
		)
	}
}

func TestBuildCreatesDailyAndWeeklyBuckets(
	t *testing.T,
) {
	tests := []struct {
		name         string
		startTime    time.Time
		endTime      time.Time
		granularity  historicalcontract.Granularity
		wantCount    int
		wantDuration time.Duration
	}{
		{
			name: "daily",
			startTime: time.Date(
				2026,
				time.July,
				1,
				0,
				0,
				0,
				0,
				time.UTC,
			),
			endTime: time.Date(
				2026,
				time.July,
				4,
				0,
				0,
				0,
				0,
				time.UTC,
			),
			granularity: historicalcontract.
				GranularityDay,
			wantCount:    3,
			wantDuration: 24 * time.Hour,
		},
		{
			name: "weekly",
			startTime: time.Date(
				2026,
				time.June,
				29,
				0,
				0,
				0,
				0,
				time.UTC,
			),
			endTime: time.Date(
				2026,
				time.July,
				20,
				0,
				0,
				0,
				0,
				time.UTC,
			),
			granularity: historicalcontract.
				GranularityWeek,
			wantCount:    3,
			wantDuration: 7 * 24 * time.Hour,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				plan := mustBuild(
					t,
					Request{
						StartTime: test.startTime,
						EndTime:   test.endTime,
						AsOfTime: test.endTime.Add(
							time.Hour,
						),
						Granularity: test.granularity,
					},
				)

				if len(plan.Buckets) !=
					test.wantCount {
					t.Fatalf(
						"bucket count = %d, want %d",
						len(plan.Buckets),
						test.wantCount,
					)
				}
				for _, bucket := range plan.Buckets {
					if bucket.Duration() !=
						test.wantDuration {
						t.Fatalf(
							"bucket duration = %s, want %s",
							bucket.Duration(),
							test.wantDuration,
						)
					}
				}
			},
		)
	}
}

func TestBuildCustomWindowClipsAtAsOfTime(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		1,
		10,
		15,
		0,
		0,
		time.UTC,
	)
	asOfTime := startTime.Add(90 * time.Minute)
	endTime := asOfTime.Add(30 * time.Minute)

	plan := mustBuild(
		t,
		Request{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
			Granularity: historicalcontract.
				GranularityCustom,
		},
	)

	if len(plan.Buckets) != 1 ||
		!plan.Buckets[0].StartTime.
			Equal(startTime) ||
		!plan.Buckets[0].EndTime.
			Equal(asOfTime) ||
		plan.Buckets[0].Duration() !=
			90*time.Minute ||
		!plan.TruncatedByAsOfTime ||
		len(plan.Exclusions) != 1 ||
		plan.Exclusions[0].Reason !=
			ExclusionReasonFutureAfterAsOfTime {
		t.Fatalf(
			"unexpected custom plan: %#v",
			plan,
		)
	}
}

func TestBuildNormalizesAllTimesToUTC(
	t *testing.T,
) {
	location := time.FixedZone(
		"Asia/Baku",
		4*60*60,
	)
	startTime := time.Date(
		2026,
		time.July,
		1,
		12,
		0,
		0,
		0,
		location,
	)
	endTime := startTime.Add(2 * time.Hour)
	asOfTime := endTime.Add(time.Hour)

	plan := mustBuild(
		t,
		Request{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
			Granularity: historicalcontract.
				GranularityHour,
		},
	)

	for _, value := range []time.Time{
		plan.RequestedStartTime,
		plan.RequestedEndTime,
		plan.AsOfTime,
		plan.Buckets[0].StartTime,
		plan.Buckets[0].EndTime,
		plan.EffectiveWindow.StartTime,
		plan.PreviousWindow.StartTime,
	} {
		if value.Location() != time.UTC {
			t.Fatalf(
				"time is not UTC: %s %s",
				value,
				value.Location(),
			)
		}
	}
}

func TestBuildRejectsInvalidRequests(
	t *testing.T,
) {
	validStart := time.Date(
		2026,
		time.July,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	validEnd := validStart.Add(time.Hour)
	validAsOf := validEnd.Add(time.Hour)

	tests := []struct {
		name    string
		request Request
		want    error
	}{
		{
			name: "start",
			request: Request{
				EndTime:  validEnd,
				AsOfTime: validAsOf,
				Granularity: historicalcontract.
					GranularityHour,
			},
			want: ErrStartTimeRequired,
		},
		{
			name: "end",
			request: Request{
				StartTime: validStart,
				AsOfTime:  validAsOf,
				Granularity: historicalcontract.
					GranularityHour,
			},
			want: ErrEndTimeRequired,
		},
		{
			name: "as of",
			request: Request{
				StartTime: validStart,
				EndTime:   validEnd,
				Granularity: historicalcontract.
					GranularityHour,
			},
			want: ErrAsOfTimeRequired,
		},
		{
			name: "window",
			request: Request{
				StartTime: validEnd,
				EndTime:   validStart,
				AsOfTime:  validAsOf,
				Granularity: historicalcontract.
					GranularityHour,
			},
			want: ErrWindowNotPositive,
		},
		{
			name: "granularity",
			request: Request{
				StartTime:   validStart,
				EndTime:     validEnd,
				AsOfTime:    validAsOf,
				Granularity: "minute",
			},
			want: ErrUnsupportedGranularity,
		},
		{
			name: "maximum low",
			request: Request{
				StartTime: validStart,
				EndTime:   validEnd,
				AsOfTime:  validAsOf,
				Granularity: historicalcontract.
					GranularityHour,
				MaximumBucketCount: -1,
			},
			want: ErrInvalidMaximumBucketCount,
		},
		{
			name: "maximum high",
			request: Request{
				StartTime: validStart,
				EndTime:   validEnd,
				AsOfTime:  validAsOf,
				Granularity: historicalcontract.
					GranularityHour,
				MaximumBucketCount: MaximumBucketCount + 1,
			},
			want: ErrInvalidMaximumBucketCount,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				_, err := Build(
					context.Background(),
					test.request,
				)
				if !errors.Is(
					err,
					test.want,
				) {
					t.Fatalf(
						"Build() error = %v, want %v",
						err,
						test.want,
					)
				}
			},
		)
	}
}

func TestBuildRejectsExcessiveBucketCount(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	_, err := Build(
		context.Background(),
		Request{
			StartTime: startTime,
			EndTime:   startTime.Add(3 * time.Hour),
			AsOfTime:  startTime.Add(4 * time.Hour),
			Granularity: historicalcontract.
				GranularityHour,
			MaximumBucketCount: 2,
		},
	)

	var countErr *BucketCountExceededError
	if !errors.As(err, &countErr) ||
		countErr.Count != 3 ||
		countErr.Maximum != 2 ||
		countErr.Granularity !=
			historicalcontract.
				GranularityHour {
		t.Fatalf(
			"unexpected error: %#v",
			err,
		)
	}
}

func TestBuildPreservesContextCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	_, err := Build(
		ctx,
		Request{
			StartTime: time.Now().
				Add(-time.Hour),
			EndTime:  time.Now(),
			AsOfTime: time.Now().Add(time.Hour),
			Granularity: historicalcontract.
				GranularityHour,
		},
	)
	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"Build() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestGeneratedPlanCanBackValidHistoricalContract(
	t *testing.T,
) {
	startTime := time.Date(
		2026,
		time.July,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	endTime := startTime.Add(
		2 * 24 * time.Hour,
	)
	asOfTime := endTime.Add(time.Hour)

	plan := mustBuild(
		t,
		Request{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime:  asOfTime,
			Granularity: historicalcontract.
				GranularityDay,
		},
	)

	points := make(
		[]historicalcontract.Point,
		0,
		len(plan.Buckets),
	)
	for index, bucket := range plan.Buckets {
		sampleCount := index + 1
		points = append(
			points,
			historicalcontract.Point{
				StartTime: bucket.StartTime,
				EndTime:   bucket.EndTime,
				Status: historicalcontract.
					BucketStatusComplete,
				Value:         float64(sampleCount),
				SampleCount:   sampleCount,
				CoverageRatio: 1,
				Confidence: historicalcontract.
					Confidence{
					Score: 1,
					Level: historicalcontract.
						ConfidenceLevelHigh,
					SampleCount: sampleCount,
					Reasons: []historicalcontract.
						ConfidenceReason{
						{
							Code:         "complete_bucket",
							Message:      "Bucket is complete.",
							Contribution: 1,
						},
					},
				},
			},
		)
	}

	totalSamples := 0
	for _, point := range points {
		totalSamples += point.SampleCount
	}

	result := historicalcontract.Result{
		SchemaVersion: historicalcontract.
			SchemaVersionV1,
		Status: historicalcontract.
			SeriesStatusComplete,
		Metric: historicalcontract.Metric{
			Name: historicalcontract.
				MetricNameFlightCount,
			Unit: "flights",
			Aggregation: historicalcontract.
				AggregationCount,
		},
		Scope: historicalcontract.Scope{
			Type: historicalcontract.
				ScopeTypeGlobal,
		},
		Window:      *plan.EffectiveWindow,
		Granularity: plan.Granularity,
		Points:      points,
		Summary: historicalcontract.
			Summarize(points),
		Confidence: historicalcontract.Confidence{
			Score: 1,
			Level: historicalcontract.
				ConfidenceLevelHigh,
			SampleCount: totalSamples,
			Reasons: []historicalcontract.
				ConfidenceReason{
				{
					Code:         "complete_series",
					Message:      "Series is complete.",
					Contribution: 1,
				},
			},
		},
		Provenance: historicalcontract.Provenance{
			BuilderVersion: "test-builder-v1",
			InputFingerprint: "sha256:" +
				strings.Repeat(
					"a",
					64,
				),
			SourceNames: []string{
				"flight_trajectories",
			},
			LatestSourceUpdatedAt: endTime,
		},
		GeneratedAt: asOfTime.Add(time.Second),
	}

	report := historicalcontract.Validate(
		result,
	)
	if report.Status !=
		historicalcontract.
			ValidationStatusValid {
		t.Fatalf(
			"generated contract is invalid: %#v",
			report,
		)
	}
}

func mustBuild(
	t *testing.T,
	request Request,
) Plan {
	t.Helper()

	plan, err := Build(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf(
			"Build() error = %v",
			err,
		)
	}

	return plan
}
