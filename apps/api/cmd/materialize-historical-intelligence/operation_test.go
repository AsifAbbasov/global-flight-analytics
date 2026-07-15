package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalreplay"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

type materializerStub struct {
	request historicalmaterialization.Request
	outcome historicalmaterialization.Outcome
	err     error
	calls   int
}

func (stub *materializerStub) Materialize(
	_ context.Context,
	request historicalmaterialization.Request,
) (historicalmaterialization.Outcome, error) {
	stub.calls++
	stub.request = request
	return stub.outcome.Clone(), stub.err
}

type replayRunnerStub struct {
	request historicalreplay.Request
	result  historicalreplay.Result
	err     error
	calls   int
}

func (stub *replayRunnerStub) Run(
	_ context.Context,
	request historicalreplay.Request,
) (historicalreplay.Result, error) {
	stub.calls++
	stub.request = request
	return stub.result.Clone(), stub.err
}

func TestCommandOperationMaterializesOneAggregate(
	t *testing.T,
) {
	options := operationTestOptions(
		operationModeMaterialize,
	)
	record := operationTestRecord(
		options.StartTime,
		options.EndTime,
		5,
	)
	materializer := &materializerStub{
		outcome: historicalmaterialization.Outcome{
			Version: historicalmaterialization.Version,
			ReadSummary: historicalmaterialization.ReadSummary{
				Window: historicalcontract.TimeWindow{
					StartTime: options.
						StartTime.Add(
						-2 * time.Hour,
					),
					EndTime:  options.EndTime,
					AsOfTime: options.AsOfTime,
				},
				FlightCount:      7,
				TrajectoryCount:  6,
				ObservationCount: 15,
				RouteCount:       5,
			},
			Record: record,
		},
	}
	replayer := &replayRunnerStub{}
	operation, err := newCommandOperation(
		materializer,
		replayer,
		func() time.Time {
			return operationTestNow()
		},
	)
	if err != nil {
		t.Fatalf(
			"compose command operation: %v",
			err,
		)
	}

	report, err := operation.Execute(
		context.Background(),
		options,
	)
	if err != nil {
		t.Fatalf(
			"execute materialization: %v",
			err,
		)
	}

	if materializer.calls != 1 ||
		replayer.calls != 0 {
		t.Fatalf(
			"materializer calls=%d replay calls=%d",
			materializer.calls,
			replayer.calls,
		)
	}
	if materializer.request.MetricName !=
		options.MetricName ||
		materializer.request.Scope !=
			options.Scope ||
		materializer.request.DatasetLimit !=
			options.DatasetLimit ||
		materializer.request.MaximumBucketCount !=
			options.MaximumBucketCount ||
		!materializer.request.GeneratedAt.Equal(
			operationTestNow(),
		) {
		t.Fatalf(
			"unexpected materialization request: %#v",
			materializer.request,
		)
	}
	if report.Version != commandVersion ||
		report.Mode != "materialize" ||
		report.MaterializedRecordCount != 1 ||
		report.ReplayWindowCount != 0 ||
		len(report.Records) != 1 ||
		report.Records[0].ID != record.ID ||
		report.ReadSummary == nil ||
		report.ReadSummary.FlightCount != 7 {
		t.Fatalf(
			"unexpected materialization report: %#v",
			report,
		)
	}
}

func TestCommandOperationReplaysBoundedWindows(
	t *testing.T,
) {
	options := operationTestOptions(
		operationModeReplay,
	)
	firstEnd := options.StartTime.Add(
		time.Hour,
	)
	secondEnd := firstEnd.Add(
		time.Hour,
	)
	replayer := &replayRunnerStub{
		result: historicalreplay.Result{
			Version: historicalreplay.Version,
			Plan: historicalwindow.Plan{
				Version: historicalwindow.Version,
			},
			Windows: []historicalreplay.WindowResult{
				{
					Bucket: historicalwindow.Bucket{
						Sequence:  1,
						StartTime: options.StartTime,
						EndTime:   firstEnd,
					},
					Record: operationTestRecord(
						options.StartTime,
						firstEnd,
						2,
					),
				},
				{
					Bucket: historicalwindow.Bucket{
						Sequence:  2,
						StartTime: firstEnd,
						EndTime:   secondEnd,
					},
					Record: operationTestRecord(
						firstEnd,
						secondEnd,
						3,
					),
				},
			},
		},
	}
	materializer := &materializerStub{}
	operation, err := newCommandOperation(
		materializer,
		replayer,
		func() time.Time {
			return operationTestNow()
		},
	)
	if err != nil {
		t.Fatalf(
			"compose command operation: %v",
			err,
		)
	}

	report, err := operation.Execute(
		context.Background(),
		options,
	)
	if err != nil {
		t.Fatalf(
			"execute replay: %v",
			err,
		)
	}

	if materializer.calls != 0 ||
		replayer.calls != 1 {
		t.Fatalf(
			"materializer calls=%d replay calls=%d",
			materializer.calls,
			replayer.calls,
		)
	}
	if replayer.request.MaximumWindowCount !=
		options.MaximumWindowCount ||
		replayer.request.DatasetLimit !=
			options.DatasetLimit ||
		!replayer.request.GeneratedAt.Equal(
			operationTestNow(),
		) {
		t.Fatalf(
			"unexpected replay request: %#v",
			replayer.request,
		)
	}
	if report.Mode != "replay" ||
		report.MaterializedRecordCount != 2 ||
		report.ReplayWindowCount != 2 ||
		report.MaximumWindowCount !=
			options.MaximumWindowCount ||
		len(report.Records) != 2 ||
		report.Records[0].Total != 2 ||
		report.Records[1].Total != 3 ||
		report.ReadSummary != nil {
		t.Fatalf(
			"unexpected replay report: %#v",
			report,
		)
	}
}

func TestCommandOperationPropagatesErrorsAndContext(
	t *testing.T,
) {
	sentinel := errors.New(
		"materialization failed",
	)
	materializer := &materializerStub{
		err: sentinel,
	}
	operation, err := newCommandOperation(
		materializer,
		&replayRunnerStub{},
		func() time.Time {
			return operationTestNow()
		},
	)
	if err != nil {
		t.Fatalf(
			"compose command operation: %v",
			err,
		)
	}

	_, err = operation.Execute(
		context.Background(),
		operationTestOptions(
			operationModeMaterialize,
		),
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf(
			"error = %v, want sentinel",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()
	_, err = operation.Execute(
		ctx,
		operationTestOptions(
			operationModeMaterialize,
		),
	)
	if !errors.Is(
		err,
		context.Canceled,
	) {
		t.Fatalf(
			"error = %v, want context canceled",
			err,
		)
	}
}

func TestNewCommandOperationRejectsMissingDependencies(
	t *testing.T,
) {
	clock := func() time.Time {
		return operationTestNow()
	}

	tests := []struct {
		name         string
		materializer commandMaterializer
		replayer     commandReplayRunner
		clock        func() time.Time
		expected     error
	}{
		{
			name:     "materializer",
			replayer: &replayRunnerStub{},
			clock:    clock,
			expected: errMaterializerRequired,
		},
		{
			name:         "replay runner",
			materializer: &materializerStub{},
			clock:        clock,
			expected:     errReplayRunnerRequired,
		},
		{
			name:         "clock",
			materializer: &materializerStub{},
			replayer:     &replayRunnerStub{},
			expected:     errClockRequired,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				operation, err :=
					newCommandOperation(
						test.materializer,
						test.replayer,
						test.clock,
					)
				if operation != nil {
					t.Fatalf(
						"operation = %#v, want nil",
						operation,
					)
				}
				if !errors.Is(
					err,
					test.expected,
				) {
					t.Fatalf(
						"error = %v, want %v",
						err,
						test.expected,
					)
				}
			},
		)
	}
}

func operationTestOptions(
	mode operationMode,
) commandOptions {
	startTime := time.Date(
		2026,
		time.July,
		14,
		8,
		0,
		0,
		0,
		time.UTC,
	)

	return commandOptions{
		Mode: mode,

		StartTime: startTime,
		EndTime: startTime.Add(
			2 * time.Hour,
		),
		AsOfTime: time.Date(
			2026,
			time.July,
			15,
			10,
			0,
			0,
			0,
			time.UTC,
		),

		Granularity: historicalcontract.GranularityHour,
		MetricName:  historicalcontract.MetricNameFlightCount,
		Scope: historicalcontract.Scope{
			Type: historicalcontract.ScopeTypeGlobal,
		},

		DatasetLimit:       5000,
		MaximumBucketCount: 100,
		MaximumWindowCount: 20,
	}
}

func operationTestNow() time.Time {
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

func operationTestRecord(
	startTime time.Time,
	endTime time.Time,
	total float64,
) historicalaggregate.Record {
	result := historicalcontract.Result{
		SchemaVersion: historicalcontract.SchemaVersionV1,
		Status:        historicalcontract.SeriesStatusComplete,
		Metric: historicalcontract.Metric{
			Name:        historicalcontract.MetricNameFlightCount,
			Unit:        "flights",
			Aggregation: historicalcontract.AggregationCount,
		},
		Scope: historicalcontract.Scope{
			Type: historicalcontract.ScopeTypeGlobal,
		},
		Window: historicalcontract.TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			AsOfTime: operationTestOptions(
				operationModeMaterialize,
			).AsOfTime,
		},
		Granularity: historicalcontract.GranularityHour,
		Summary: historicalcontract.Summary{
			PointCount: 1,
			Total:      total,
		},
		Confidence: historicalcontract.Confidence{
			Level: historicalcontract.
				ConfidenceLevelHigh,
		},
	}

	return historicalaggregate.Record{
		ID: "historical-aggregate-record-test-" +
			startTime.Format("150405"),
		InputFingerprint: "sha256:test-fingerprint",
		Result:           result,
		StoredAt:         operationTestNow(),
	}
}
