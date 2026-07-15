package historicalreplay

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
)

type materializeFunc func(
	context.Context,
	historicalmaterialization.Request,
) (historicalmaterialization.Outcome, error)

func (function materializeFunc) Materialize(
	ctx context.Context,
	request historicalmaterialization.Request,
) (historicalmaterialization.Outcome, error) {
	return function(ctx, request)
}

func TestRunReplaysClosedWindowsChronologically(
	t *testing.T,
) {
	asOfTime := replayTestTime().
		Add(time.Hour + 45*time.Minute)
	requests := make(
		[]historicalmaterialization.Request,
		0,
		2,
	)
	runner, err := New(
		Config{
			Materializer: materializeFunc(
				func(
					_ context.Context,
					request historicalmaterialization.Request,
				) (
					historicalmaterialization.Outcome,
					error,
				) {
					requests = append(requests, request)
					return replayOutcome(
						request,
						len(requests),
					), nil
				},
			),
			Now: func() time.Time {
				return asOfTime
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create replay runner: %v",
			err,
		)
	}

	result, err := runner.Run(
		context.Background(),
		Request{
			StartTime: replayTestTime().
				Add(-105 * time.Minute),
			EndTime:  asOfTime,
			AsOfTime: asOfTime,
			Granularity: historicalcontract.
				GranularityHour,
			MetricName: historicalcontract.
				MetricNameFlightCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			MaximumWindowCount: 10,
			GeneratedAt:        asOfTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"run historical replay: %v",
			err,
		)
	}

	if len(requests) != 2 ||
		len(result.Windows) != 2 {
		t.Fatalf(
			"materialized windows=%d result windows=%d want=2",
			len(requests),
			len(result.Windows),
		)
	}

	expectedFirstStart := replayTestTime().
		Add(-time.Hour)
	expectedSecondStart := replayTestTime()
	if !requests[0].StartTime.Equal(
		expectedFirstStart,
	) ||
		!requests[0].EndTime.Equal(
			expectedSecondStart,
		) ||
		!requests[1].StartTime.Equal(
			expectedSecondStart,
		) ||
		!requests[1].EndTime.Equal(
			expectedSecondStart.Add(time.Hour),
		) {
		t.Fatalf(
			"unexpected replay order: %#v",
			requests,
		)
	}
	if result.Windows[0].Bucket.Sequence != 1 ||
		result.Windows[1].Bucket.Sequence != 2 {
		t.Fatalf(
			"unexpected replay sequences: %#v",
			result.Windows,
		)
	}
}

func TestRunReturnsCompletedPrefixWhenWindowFails(
	t *testing.T,
) {
	sentinel := errors.New("materialization failed")
	callCount := 0
	runner, err := New(
		Config{
			Materializer: materializeFunc(
				func(
					_ context.Context,
					request historicalmaterialization.Request,
				) (
					historicalmaterialization.Outcome,
					error,
				) {
					callCount++
					if callCount == 2 {
						return historicalmaterialization.
								Outcome{},
							sentinel
					}
					return replayOutcome(
						request,
						callCount,
					), nil
				},
			),
			Now: replayTestTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"create replay runner: %v",
			err,
		)
	}

	result, err := runner.Run(
		context.Background(),
		Request{
			StartTime: replayTestTime().
				Add(-2 * time.Hour),
			EndTime: replayTestTime().
				Add(time.Hour),
			AsOfTime: replayTestTime().
				Add(time.Hour),
			Granularity: historicalcontract.
				GranularityHour,
			MetricName: historicalcontract.
				MetricNameFlightCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			GeneratedAt: replayTestTime().
				Add(time.Hour),
		},
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf(
			"expected materialization failure, got %v",
			err,
		)
	}
	if len(result.Windows) != 1 {
		t.Fatalf(
			"completed prefix length=%d want=1",
			len(result.Windows),
		)
	}

	var windowErr *WindowError
	if !errors.As(err, &windowErr) ||
		windowErr.Sequence != 2 {
		t.Fatalf(
			"unexpected window error: %#v",
			err,
		)
	}
}

func TestRunRejectsExcessiveWindowCount(
	t *testing.T,
) {
	runner, err := New(
		Config{
			Materializer: materializeFunc(
				func(
					context.Context,
					historicalmaterialization.Request,
				) (
					historicalmaterialization.Outcome,
					error,
				) {
					return historicalmaterialization.
							Outcome{},
						nil
				},
			),
			Now: replayTestTime,
		},
	)
	if err != nil {
		t.Fatalf(
			"create replay runner: %v",
			err,
		)
	}

	_, err = runner.Run(
		context.Background(),
		Request{
			StartTime: replayTestTime().
				Add(-4 * time.Hour),
			EndTime:  replayTestTime(),
			AsOfTime: replayTestTime(),
			Granularity: historicalcontract.
				GranularityHour,
			MetricName: historicalcontract.
				MetricNameFlightCount,
			Scope: historicalcontract.Scope{
				Type: historicalcontract.
					ScopeTypeGlobal,
			},
			MaximumWindowCount: 2,
			GeneratedAt:        replayTestTime(),
		},
	)

	var countErr *WindowCountExceededError
	if !errors.As(err, &countErr) ||
		countErr.Count != 4 ||
		countErr.Maximum != 2 {
		t.Fatalf(
			"unexpected window-count error: %#v",
			err,
		)
	}
}

func replayOutcome(
	request historicalmaterialization.Request,
	sequence int,
) historicalmaterialization.Outcome {
	result := historicalcontract.Result{
		SchemaVersion: historicalcontract.SchemaVersionV1,
		Metric: historicalcontract.Metric{
			Name: request.MetricName,
		},
		Scope:       request.Scope,
		Granularity: request.Granularity,
		Window: historicalcontract.TimeWindow{
			StartTime: request.StartTime,
			EndTime:   request.EndTime,
			AsOfTime:  request.AsOfTime,
		},
	}
	return historicalmaterialization.Outcome{
		Version: historicalmaterialization.Version,
		Record: historicalaggregate.Record{
			ID: "historical-aggregate-record-" +
				strings.Repeat(
					string(rune('a'+sequence)),
					64,
				),
			Result: result,
		},
	}
}

func replayTestTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}
