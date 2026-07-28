package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalreplay"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
)

type partialReportMaterializer struct {
	calls int
}

func (materializer *partialReportMaterializer) Materialize(
	context.Context,
	historicalmaterialization.Request,
) (historicalmaterialization.Outcome, error) {
	materializer.calls++
	return historicalmaterialization.Outcome{}, nil
}

type partialReportReplayRunner struct {
	request historicalreplay.Request
	result  historicalreplay.Result
	err     error
	calls   int
}

func (runner *partialReportReplayRunner) Run(
	_ context.Context,
	request historicalreplay.Request,
) (historicalreplay.Result, error) {
	runner.calls++
	runner.request = request
	return runner.result.Clone(), runner.err
}

func TestCommandOperationPreservesPartialReplayReport(
	t *testing.T,
) {
	sentinel := errors.New(
		"second replay window failed",
	)
	start := partialReportTime()
	firstEnd := start.Add(time.Hour)
	secondEnd := firstEnd.Add(time.Hour)
	asOfTime := secondEnd
	generatedAt := asOfTime.Add(time.Minute)
	result := historicalreplay.Result{
		Version: historicalreplay.Version,
		Status:  historicalreplay.StatusPartial,
		Plan: historicalwindow.Plan{
			Version: historicalwindow.Version,
			Buckets: []historicalwindow.Bucket{
				{
					Key:       "first",
					Sequence:  1,
					StartTime: start,
					EndTime:   firstEnd,
				},
				{
					Key:       "second",
					Sequence:  2,
					StartTime: firstEnd,
					EndTime:   secondEnd,
				},
			},
		},
		PlannedWindowCount:   2,
		CompletedWindowCount: 1,
		Windows: []historicalreplay.WindowResult{
			{
				Bucket: historicalwindow.Bucket{
					Key:       "first",
					Sequence:  1,
					StartTime: start,
					EndTime:   firstEnd,
				},
				Record: partialReportRecord(
					start,
					firstEnd,
					asOfTime,
					generatedAt,
				),
			},
		},
		HasFailure: true,
		Failure: historicalreplay.Failure{
			Sequence:  2,
			StartTime: firstEnd,
			EndTime:   secondEnd,
			Code: historicalreplay.
				FailureCodeMaterialization,
			Message: sentinel.Error(),
		},
		GeneratedAt:      generatedAt,
		StartedAt:        generatedAt,
		CompletedAt:      generatedAt.Add(time.Second),
		InputFingerprint: "sha256:" + strings.Repeat("a", 64),
	}
	replayer := &partialReportReplayRunner{
		result: result,
		err:    sentinel,
	}
	materializer := &partialReportMaterializer{}
	operation, err := newCommandOperation(
		materializer,
		replayer,
		func() time.Time {
			return generatedAt
		},
	)
	if err != nil {
		t.Fatalf(
			"compose operation: %v",
			err,
		)
	}
	options := commandOptions{
		Mode:        operationModeReplay,
		StartTime:   start,
		EndTime:     secondEnd,
		AsOfTime:    asOfTime,
		Granularity: historicalcontract.GranularityHour,
		MetricName:  historicalcontract.MetricNameFlightCount,
		Scope: historicalcontract.Scope{
			Type: historicalcontract.ScopeTypeGlobal,
		},
		DatasetLimit:       100,
		MaximumBucketCount: 100,
		MaximumWindowCount: 10,
	}

	report, executeErr := operation.Execute(
		context.Background(),
		options,
	)
	if !errors.Is(executeErr, sentinel) {
		t.Fatalf(
			"execute error=%v want sentinel",
			executeErr,
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
	if report.Version != commandVersion ||
		report.Status != string(
			historicalreplay.StatusPartial,
		) ||
		report.PlannedReplayWindowCount != 2 ||
		report.CompletedReplayWindowCount != 1 ||
		report.MaterializedRecordCount != 1 ||
		len(report.Records) != 1 ||
		report.ReplayFailure == nil ||
		report.ReplayFailure.Sequence != 2 ||
		report.ReplayFailure.Code != string(
			historicalreplay.
				FailureCodeMaterialization,
		) {
		t.Fatalf(
			"partial report was not preserved: %#v",
			report,
		)
	}
}

func TestWriteCommandOutcomeEmitsPrefixBeforeFailure(
	t *testing.T,
) {
	report := commandReport{
		Version: commandVersion,
		Mode:    string(operationModeReplay),
		Status: string(
			historicalreplay.StatusPartial,
		),
		MaterializedRecordCount:    1,
		PlannedReplayWindowCount:   2,
		CompletedReplayWindowCount: 1,
		ReplayInputFingerprint:     "sha256:" + strings.Repeat("a", 64),
		Records:                    []reportRecord{{ID: "record-one"}},
		ReplayFailure: &reportReplayFailure{
			Sequence: 2,
			Code: string(
				historicalreplay.
					FailureCodeMaterialization,
			),
			Message: "failed",
		},
		StartedAt:   partialReportTime(),
		CompletedAt: partialReportTime().Add(time.Second),
	}
	sentinel := errors.New("replay failed")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := writeCommandOutcome(
		stdout,
		stderr,
		operationModeReplay,
		report,
		sentinel,
	)
	if exitCode != 1 {
		t.Fatalf(
			"exit code=%d want=1",
			exitCode,
		)
	}
	var decoded commandReport
	if err := json.Unmarshal(
		stdout.Bytes(),
		&decoded,
	); err != nil {
		t.Fatalf(
			"decode partial report: %v output=%q",
			err,
			stdout.String(),
		)
	}
	if decoded.Status != string(
		historicalreplay.StatusPartial,
	) ||
		len(decoded.Records) != 1 ||
		decoded.Records[0].ID != "record-one" ||
		decoded.ReplayFailure == nil ||
		decoded.ReplayFailure.Sequence != 2 {
		t.Fatalf(
			"decoded report lost completed prefix: %#v",
			decoded,
		)
	}
	if !strings.Contains(
		stderr.String(),
		sentinel.Error(),
	) {
		t.Fatalf(
			"stderr=%q misses replay error",
			stderr.String(),
		)
	}
}

func TestCommandOperationRejectsNilContext(
	t *testing.T,
) {
	replayer := &partialReportReplayRunner{}
	materializer := &partialReportMaterializer{}
	operation, err := newCommandOperation(
		materializer,
		replayer,
		partialReportTime,
	)
	if err != nil {
		t.Fatalf(
			"compose operation: %v",
			err,
		)
	}

	report, executeErr := operation.Execute(
		nil,
		commandOptions{},
	)
	if !errors.Is(
		executeErr,
		errCommandContextRequired,
	) ||
		report.Version != "" ||
		materializer.calls != 0 ||
		replayer.calls != 0 {
		t.Fatalf(
			"report=%#v error=%v materializer=%d replay=%d",
			report,
			executeErr,
			materializer.calls,
			replayer.calls,
		)
	}
}

func partialReportRecord(
	startTime time.Time,
	endTime time.Time,
	asOfTime time.Time,
	generatedAt time.Time,
) historicalaggregate.Record {
	return historicalaggregate.Record{
		ID:               "record-one",
		InputFingerprint: "sha256:" + strings.Repeat("b", 64),
		Result: historicalcontract.Result{
			Status: historicalcontract.SeriesStatusComplete,
			Window: historicalcontract.TimeWindow{
				StartTime: startTime,
				EndTime:   endTime,
				AsOfTime:  asOfTime,
			},
			GeneratedAt: generatedAt,
		},
		StoredAt: generatedAt,
	}
}

func partialReportTime() time.Time {
	return time.Date(
		2026,
		time.July,
		15,
		10,
		0,
		0,
		0,
		time.UTC,
	)
}
