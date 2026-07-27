package historicalseries

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

func TestBuildCanonicalizesMutablePlan(
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
	plan, err := historicalwindow.Build(
		context.Background(),
		historicalwindow.Request{
			StartTime:   startTime,
			EndTime:     startTime.Add(time.Hour),
			AsOfTime:    startTime.Add(2 * time.Hour),
			Granularity: historicalcontract.GranularityHour,
		},
	)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	plan.Fingerprint = "sha256:" +
		strings.Repeat("0", 64)
	plan.Buckets[0].Key = "tampered"
	plan.Buckets[0].Sequence = 99
	plan.PreviousWindow.StartTime =
		plan.PreviousWindow.StartTime.Add(time.Hour)

	values, err := BindDatasetCoverage(
		[]BucketValue{
			{
				Bucket:      plan.Buckets[0],
				Value:       1,
				SampleCount: 1,
			},
		},
		DatasetCoverage{
			State:        DatasetReadComplete,
			MatchedCount: 1,
		},
	)
	if err != nil {
		t.Fatalf("bind coverage: %v", err)
	}

	result, err := Build(
		BuildRequest{
			Metric: historicalcontract.Metric{
				Name:        historicalcontract.MetricNameObservationCount,
				Unit:        "observations",
				Aggregation: historicalcontract.AggregationCount,
			},
			Scope: historicalcontract.Scope{
				Type: historicalcontract.ScopeTypeGlobal,
			},
			Plan:             plan,
			Values:           values,
			BuilderVersion:   Version,
			InputFingerprint: "sha256:" + strings.Repeat("a", 64),
			SourceNames:      []string{"flight_states"},
			LatestSourceUpdatedAt: startTime.
				Add(30 * time.Minute),
			GeneratedAt: startTime.Add(2 * time.Hour),
		},
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if result.Status !=
		historicalcontract.SeriesStatusComplete {
		t.Fatalf(
			"status = %s, want complete",
			result.Status,
		)
	}
}

func TestBuildRejectsInvalidPlanRequestSemantics(
	t *testing.T,
) {
	plan := historicalwindow.Plan{
		Version: historicalwindow.Version,
		RequestedStartTime: time.Date(
			2026,
			time.July,
			2,
			0,
			0,
			0,
			0,
			time.UTC,
		),
		RequestedEndTime: time.Date(
			2026,
			time.July,
			1,
			0,
			0,
			0,
			0,
			time.UTC,
		),
		AsOfTime: time.Date(
			2026,
			time.July,
			3,
			0,
			0,
			0,
			0,
			time.UTC,
		),
		Granularity: historicalcontract.GranularityHour,
		MaximumBucketCount: historicalwindow.
			DefaultMaximumBucketCount,
	}

	_, err := Build(
		BuildRequest{
			Plan: plan,
		},
	)
	if !errors.Is(
		err,
		historicalwindow.ErrPlanIntegrityInvalid,
	) {
		t.Fatalf(
			"Build() error = %v, want %v",
			err,
			historicalwindow.ErrPlanIntegrityInvalid,
		)
	}
}
