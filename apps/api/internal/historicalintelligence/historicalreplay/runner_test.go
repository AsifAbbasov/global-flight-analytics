package historicalreplay

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcomparison"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalcontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalmaterialization"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalread"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalwindow"
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

func TestRunReturnsValidatedCompleteReplay(
	t *testing.T,
) {
	request := replayRequest(
		replayAsOfTime().Add(
			-2*time.Hour,
		),
		replayAsOfTime(),
	)
	request.GeneratedAt = time.Time{}

	materializationRequests := make(
		[]historicalmaterialization.Request,
		0,
		2,
	)
	runner := replayRunner(
		t,
		materializeFunc(
			func(
				_ context.Context,
				request historicalmaterialization.Request,
			) (
				historicalmaterialization.Outcome,
				error,
			) {
				materializationRequests = append(
					materializationRequests,
					request,
				)
				return validReplayOutcome(
					t,
					request,
				), nil
			},
		),
	)

	result, err := runner.Run(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf(
			"run replay: %v",
			err,
		)
	}
	if result.Status != StatusComplete ||
		result.PlannedWindowCount != 2 ||
		result.CompletedWindowCount != 2 ||
		len(result.Windows) != 2 ||
		result.HasFailure {
		t.Fatalf(
			"unexpected result: %#v",
			result,
		)
	}
	if result.GeneratedAt != replayNow() ||
		result.StartedAt != replayNow() ||
		result.CompletedAt != replayNow() {
		t.Fatalf(
			"unexpected replay timestamps: %#v",
			result,
		)
	}
	for _, materializationRequest := range materializationRequests {
		if materializationRequest.
			MaximumBucketCount != 1 {
			t.Fatalf(
				"per-window maximum bucket count=%d want=1",
				materializationRequest.
					MaximumBucketCount,
			)
		}
		if !materializationRequest.
			GeneratedAt.Equal(replayNow()) {
			t.Fatalf(
				"generated at=%s want=%s",
				materializationRequest.GeneratedAt,
				replayNow(),
			)
		}
	}
	if result.Windows[0].
		CurrentPeriodInputFingerprint !=
		result.Windows[1].
			PreviousPeriodInputFingerprint {
		t.Fatal(
			"adjacent period fingerprints do not form a continuous replay chain",
		)
	}
	if !validFingerprint(
		result.InputFingerprint,
	) {
		t.Fatalf(
			"input fingerprint=%q",
			result.InputFingerprint,
		)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf(
			"validate replay result: %v",
			err,
		)
	}
}

func TestRunReturnsSelfContainedPartialResult(
	t *testing.T,
) {
	sentinel := errors.New(
		"materialization failed",
	)
	callCount := 0
	runner := replayRunner(
		t,
		materializeFunc(
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
				return validReplayOutcome(
					t,
					request,
				), nil
			},
		),
	)

	result, err := runner.Run(
		context.Background(),
		replayRequest(
			replayAsOfTime().Add(
				-3*time.Hour,
			),
			replayAsOfTime(),
		),
	)
	if !errors.Is(err, sentinel) {
		t.Fatalf(
			"error=%v want sentinel",
			err,
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
	if result.Status != StatusPartial ||
		result.PlannedWindowCount != 3 ||
		result.CompletedWindowCount != 1 ||
		len(result.Windows) != 1 ||
		!result.HasFailure ||
		result.Failure.Sequence != 2 ||
		result.Failure.Code !=
			FailureCodeMaterialization {
		t.Fatalf(
			"unexpected partial result: %#v",
			result,
		)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf(
			"validate partial result: %v",
			err,
		)
	}
}

func TestRunRejectsInvalidMaterializationOutcome(
	t *testing.T,
) {
	tests := []struct {
		name   string
		mutate func(
			*historicalmaterialization.Outcome,
		)
		expected error
	}{
		{
			name: "version",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.Version = ""
			},
			expected: ErrOutcomeVersionMismatch,
		},
		{
			name: "record id",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.Record.ID = ""
			},
			expected: ErrOutcomeRecordMismatch,
		},
		{
			name: "record window",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.Record.Key.Window.EndTime =
					outcome.Record.Key.Window.
						EndTime.Add(time.Hour)
			},
			expected: ErrOutcomeRecordMismatch,
		},
		{
			name: "record metric",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.Record.Key.MetricName =
					historicalcontract.
						MetricNameActiveAircraft
			},
			expected: ErrOutcomeRecordMismatch,
		},
		{
			name: "record scope",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.Record.Key.Scope =
					historicalcontract.Scope{
						Type: historicalcontract.
							ScopeTypeAirport,
						AirportICAOCode: "UBBB",
					}
			},
			expected: ErrOutcomeRecordMismatch,
		},
		{
			name: "record granularity",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.Record.Key.Granularity =
					historicalcontract.
						GranularityDay
			},
			expected: ErrOutcomeRecordMismatch,
		},
		{
			name: "record as of time",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.Record.Key.Window.AsOfTime =
					outcome.Record.Key.Window.
						AsOfTime.Add(time.Hour)
			},
			expected: ErrOutcomeRecordMismatch,
		},
		{
			name: "record fingerprint",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.Record.InputFingerprint =
					"sha256:" +
						strings.Repeat(
							"f",
							64,
						)
			},
			expected: ErrOutcomeFingerprintMismatch,
		},
		{
			name: "invalid current result",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.CurrentResult.Status = ""
				outcome.Record.Result =
					outcome.CurrentResult.Clone()
			},
			expected: ErrOutcomeResultInvalid,
		},
		{
			name: "missing comparison",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.CurrentResult.Comparison = nil
				outcome.Record.Result =
					outcome.CurrentResult.Clone()
			},
			expected: ErrOutcomeComparisonRequired,
		},
		{
			name: "period fingerprint",
			mutate: func(
				outcome *historicalmaterialization.Outcome,
			) {
				outcome.PreviousResult.
					Provenance.InputFingerprint =
					"invalid"
			},
			expected: ErrOutcomeResultInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				runner := replayRunner(
					t,
					materializeFunc(
						func(
							_ context.Context,
							request historicalmaterialization.Request,
						) (
							historicalmaterialization.Outcome,
							error,
						) {
							outcome :=
								validReplayOutcome(
									t,
									request,
								)
							test.mutate(&outcome)
							return outcome, nil
						},
					),
				)

				result, err := runner.Run(
					context.Background(),
					replayRequest(
						replayAsOfTime().
							Add(-time.Hour),
						replayAsOfTime(),
					),
				)
				if !errors.Is(
					err,
					test.expected,
				) {
					t.Fatalf(
						"error=%v want=%v",
						err,
						test.expected,
					)
				}
				var windowErr *WindowError
				var outcomeErr *OutcomeContractError
				if !errors.As(
					err,
					&windowErr,
				) ||
					!errors.As(
						err,
						&outcomeErr,
					) {
					t.Fatalf(
						"unexpected error chain: %#v",
						err,
					)
				}
				if result.Status !=
					StatusFailed ||
					result.CompletedWindowCount != 0 ||
					result.Failure.Code !=
						FailureCodeOutcomeContract {
					t.Fatalf(
						"unexpected failed result: %#v",
						result,
					)
				}
			},
		)
	}
}

func TestRunValidatesGlobalRequestBeforeReplay(
	t *testing.T,
) {
	tests := []struct {
		name     string
		mutate   func(*Request)
		expected error
	}{
		{
			name: "unsupported metric",
			mutate: func(request *Request) {
				request.MetricName =
					historicalcontract.
						MetricNamePeakActivity
			},
			expected: ErrMetricUnsupported,
		},
		{
			name: "malformed scope",
			mutate: func(request *Request) {
				request.Scope.AirportICAOCode =
					"UBBB"
			},
			expected: ErrScopeUnsupported,
		},
		{
			name: "dataset limit below minimum",
			mutate: func(request *Request) {
				request.DatasetLimit = -1
			},
			expected: ErrDatasetLimitInvalid,
		},
		{
			name: "dataset limit above maximum",
			mutate: func(request *Request) {
				request.DatasetLimit =
					historicalread.
						MaximumDatasetLimit + 1
			},
			expected: ErrDatasetLimitInvalid,
		},
		{
			name: "window limit below minimum",
			mutate: func(request *Request) {
				request.MaximumWindowCount = -1
			},
			expected: ErrMaximumWindowCountInvalid,
		},
		{
			name: "window limit above maximum",
			mutate: func(request *Request) {
				request.MaximumWindowCount =
					MaximumWindowCount + 1
			},
			expected: ErrMaximumWindowCountInvalid,
		},
		{
			name: "bucket limit below minimum",
			mutate: func(request *Request) {
				request.MaximumBucketCount = -1
			},
			expected: ErrMaximumBucketCountInvalid,
		},
		{
			name: "bucket limit above maximum",
			mutate: func(request *Request) {
				request.MaximumBucketCount =
					historicalwindow.
						MaximumBucketCount + 1
			},
			expected: ErrMaximumBucketCountInvalid,
		},
		{
			name: "generated before as of time",
			mutate: func(request *Request) {
				request.GeneratedAt =
					request.AsOfTime.Add(
						-time.Second,
					)
			},
			expected: ErrGeneratedAtBeforeAsOfTime,
		},
		{
			name: "generated after replay start",
			mutate: func(request *Request) {
				request.GeneratedAt =
					replayNow().Add(time.Second)
			},
			expected: ErrGeneratedAtAfterStartTime,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				calls := 0
				runner := replayRunner(
					t,
					materializeFunc(
						func(
							context.Context,
							historicalmaterialization.Request,
						) (
							historicalmaterialization.Outcome,
							error,
						) {
							calls++
							return historicalmaterialization.
									Outcome{},
								nil
						},
					),
				)
				request := replayRequest(
					replayAsOfTime().
						Add(-time.Hour),
					replayAsOfTime(),
				)
				test.mutate(&request)

				result, err := runner.Run(
					context.Background(),
					request,
				)
				if !errors.Is(
					err,
					test.expected,
				) {
					t.Fatalf(
						"error=%v want=%v",
						err,
						test.expected,
					)
				}
				var windowErr *WindowError
				if errors.As(
					err,
					&windowErr,
				) {
					t.Fatalf(
						"global request failure became a window error: %v",
						err,
					)
				}
				if calls != 0 ||
					result.Version != "" ||
					len(result.Windows) != 0 {
					t.Fatalf(
						"calls=%d result=%#v",
						calls,
						result,
					)
				}
			},
		)
	}
}

func TestRunUsesBoundedPlanningLimits(
	t *testing.T,
) {
	calls := 0
	runner := replayRunner(
		t,
		materializeFunc(
			func(
				context.Context,
				historicalmaterialization.Request,
			) (
				historicalmaterialization.Outcome,
				error,
			) {
				calls++
				return historicalmaterialization.
						Outcome{},
					nil
			},
		),
	)
	request := replayRequest(
		replayAsOfTime().Add(
			-3*time.Hour,
		),
		replayAsOfTime(),
	)
	request.MaximumWindowCount = 2
	request.MaximumBucketCount = 10

	_, err := runner.Run(
		context.Background(),
		request,
	)
	var windowCountErr *WindowCountExceededError
	if !errors.As(err, &windowCountErr) ||
		windowCountErr.Count != 3 ||
		windowCountErr.Maximum != 2 {
		t.Fatalf(
			"unexpected replay limit error: %#v",
			err,
		)
	}
	if calls != 0 {
		t.Fatalf(
			"materializer calls=%d want=0",
			calls,
		)
	}

	request.MaximumWindowCount = 3
	request.MaximumBucketCount = 2
	_, err = runner.Run(
		context.Background(),
		request,
	)
	var bucketCountErr *historicalwindow.
		BucketCountExceededError
	if !errors.As(err, &bucketCountErr) {
		t.Fatalf(
			"expected bucket limit error, got %v",
			err,
		)
	}
	if errors.As(err, &windowCountErr) {
		t.Fatalf(
			"bucket limit was ambiguously mapped to replay window limit: %v",
			err,
		)
	}
}

func TestRunRejectsNilAndCanceledContext(
	t *testing.T,
) {
	runner := replayRunner(
		t,
		materializeFunc(
			func(
				context.Context,
				historicalmaterialization.Request,
			) (
				historicalmaterialization.Outcome,
				error,
			) {
				t.Fatal(
					"materializer must not be called",
				)
				return historicalmaterialization.
						Outcome{},
					nil
			},
		),
	)
	request := replayRequest(
		replayAsOfTime().Add(-time.Hour),
		replayAsOfTime(),
	)
	if _, err := runner.Run(
		nil,
		request,
	); !errors.Is(err, ErrContextRequired) {
		t.Fatalf(
			"nil context error=%v",
			err,
		)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()
	if _, err := runner.Run(
		ctx,
		request,
	); !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"canceled context error=%v",
			err,
		)
	}
}

func TestRunReturnsPartialContextCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	callCount := 0
	runner := replayRunner(
		t,
		materializeFunc(
			func(
				_ context.Context,
				request historicalmaterialization.Request,
			) (
				historicalmaterialization.Outcome,
				error,
			) {
				callCount++
				outcome := validReplayOutcome(
					t,
					request,
				)
				cancel()
				return outcome, nil
			},
		),
	)

	result, err := runner.Run(
		ctx,
		replayRequest(
			replayAsOfTime().Add(
				-2*time.Hour,
			),
			replayAsOfTime(),
		),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"error=%v want context canceled",
			err,
		)
	}
	if result.Status != StatusPartial ||
		result.CompletedWindowCount != 1 ||
		result.Failure.Code !=
			FailureCodeContextCanceled ||
		result.Failure.Sequence != 2 {
		t.Fatalf(
			"unexpected canceled result: %#v",
			result,
		)
	}
}

func TestRunRejectsOverlappingReadContinuityMismatch(
	t *testing.T,
) {
	callCount := 0
	runner := replayRunner(
		t,
		materializeFunc(
			func(
				_ context.Context,
				request historicalmaterialization.Request,
			) (
				historicalmaterialization.Outcome,
				error,
			) {
				callCount++
				outcome := validReplayOutcome(
					t,
					request,
				)
				if callCount == 2 {
					outcome.PreviousResult.
						Provenance.InputFingerprint =
						"sha256:" +
							strings.Repeat(
								"f",
								64,
							)
				}
				return outcome, nil
			},
		),
	)

	result, err := runner.Run(
		context.Background(),
		replayRequest(
			replayAsOfTime().Add(
				-2*time.Hour,
			),
			replayAsOfTime(),
		),
	)
	if !errors.Is(
		err,
		ErrReplayContinuityMismatch,
	) {
		t.Fatalf(
			"error=%v want continuity mismatch",
			err,
		)
	}
	if result.Status != StatusPartial ||
		result.CompletedWindowCount != 1 ||
		result.Failure.Code !=
			FailureCodeContinuityMismatch {
		t.Fatalf(
			"unexpected continuity result: %#v",
			result,
		)
	}
}

func TestRunReturnsFailedResultWhenNoWindowExists(
	t *testing.T,
) {
	runner := replayRunner(
		t,
		materializeFunc(
			func(
				context.Context,
				historicalmaterialization.Request,
			) (
				historicalmaterialization.Outcome,
				error,
			) {
				t.Fatal(
					"materializer must not be called",
				)
				return historicalmaterialization.
						Outcome{},
					nil
			},
		),
	)
	start := replayAsOfTime().
		Add(-30 * time.Minute)
	result, err := runner.Run(
		context.Background(),
		replayRequest(
			start,
			replayAsOfTime(),
		),
	)
	if !errors.Is(
		err,
		ErrNoReplayWindow,
	) {
		t.Fatalf(
			"error=%v want no replay window",
			err,
		)
	}
	if result.Status != StatusFailed ||
		result.PlannedWindowCount != 0 ||
		result.CompletedWindowCount != 0 ||
		result.Failure.Code !=
			FailureCodeNoReplayWindow ||
		result.Failure.Sequence != 0 {
		t.Fatalf(
			"unexpected failed result: %#v",
			result,
		)
	}
	if err := result.Validate(); err != nil {
		t.Fatalf(
			"validate no-window result: %v",
			err,
		)
	}
}

func TestRunIsDeterministicAndCloneIsolated(
	t *testing.T,
) {
	runner := replayRunner(
		t,
		materializeFunc(
			func(
				_ context.Context,
				request historicalmaterialization.Request,
			) (
				historicalmaterialization.Outcome,
				error,
			) {
				return validReplayOutcome(
					t,
					request,
				), nil
			},
		),
	)
	request := replayRequest(
		replayAsOfTime().Add(
			-2*time.Hour,
		),
		replayAsOfTime(),
	)

	first, err := runner.Run(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf(
			"first replay: %v",
			err,
		)
	}
	second, err := runner.Run(
		context.Background(),
		request,
	)
	if err != nil {
		t.Fatalf(
			"second replay: %v",
			err,
		)
	}
	if first.InputFingerprint !=
		second.InputFingerprint ||
		first.Windows[0].Record.ID !=
			second.Windows[0].Record.ID {
		t.Fatalf(
			"replay is not deterministic: first=%#v second=%#v",
			first,
			second,
		)
	}

	cloned := first.Clone()
	cloned.Windows[0].Record.Result.Points[0].
		Value = 999
	cloned.Plan.Buckets[0].Key = "mutated"
	if first.Windows[0].Record.Result.
		Points[0].Value == 999 ||
		first.Plan.Buckets[0].Key ==
			"mutated" {
		t.Fatal(
			"clone mutation leaked into original result",
		)
	}
}

func TestResultValidateRejectsTampering(
	t *testing.T,
) {
	runner := replayRunner(
		t,
		materializeFunc(
			func(
				_ context.Context,
				request historicalmaterialization.Request,
			) (
				historicalmaterialization.Outcome,
				error,
			) {
				return validReplayOutcome(
					t,
					request,
				), nil
			},
		),
	)
	result, err := runner.Run(
		context.Background(),
		replayRequest(
			replayAsOfTime().Add(
				-2*time.Hour,
			),
			replayAsOfTime(),
		),
	)
	if err != nil {
		t.Fatalf(
			"run replay: %v",
			err,
		)
	}

	tests := []struct {
		name   string
		mutate func(*Result)
	}{
		{
			name: "fingerprint",
			mutate: func(result *Result) {
				result.InputFingerprint =
					"sha256:" +
						strings.Repeat(
							"f",
							64,
						)
			},
		},
		{
			name: "completed count",
			mutate: func(result *Result) {
				result.CompletedWindowCount--
			},
		},
		{
			name: "record identity",
			mutate: func(result *Result) {
				result.Windows[0].Record.
					Key.MetricName =
					historicalcontract.
						MetricNameActiveAircraft
			},
		},
		{
			name: "continuity",
			mutate: func(result *Result) {
				result.Windows[1].
					PreviousPeriodInputFingerprint =
					"sha256:" +
						strings.Repeat(
							"f",
							64,
						)
			},
		},
	}
	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				mutated := result.Clone()
				test.mutate(&mutated)
				if err := mutated.Validate(); err == nil {
					t.Fatal(
						"tampered result validated",
					)
				}
			},
		)
	}
}

func TestNewRejectsMissingMaterializerAndZeroClock(
	t *testing.T,
) {
	runner, err := New(Config{})
	if runner != nil ||
		!errors.Is(
			err,
			ErrMaterializerRequired,
		) {
		t.Fatalf(
			"runner=%#v error=%v",
			runner,
			err,
		)
	}

	runner, err = New(
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
			Now: func() time.Time {
				return time.Time{}
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create zero-clock runner: %v",
			err,
		)
	}
	_, err = runner.Run(
		context.Background(),
		replayRequest(
			replayAsOfTime().Add(-time.Hour),
			replayAsOfTime(),
		),
	)
	if !errors.Is(err, ErrClockRequired) {
		t.Fatalf(
			"zero clock error=%v",
			err,
		)
	}
}

func replayRunner(
	t *testing.T,
	materializer Materializer,
) *Runner {
	t.Helper()
	runner, err := New(
		Config{
			Materializer: materializer,
			Now:          replayNow,
		},
	)
	if err != nil {
		t.Fatalf(
			"create replay runner: %v",
			err,
		)
	}
	return runner
}

func replayRequest(
	startTime time.Time,
	endTime time.Time,
) Request {
	return Request{
		StartTime: startTime,
		EndTime:   endTime,
		AsOfTime:  replayAsOfTime(),

		Granularity: historicalcontract.
			GranularityHour,
		MetricName: historicalcontract.
			MetricNameFlightCount,
		Scope: historicalcontract.Scope{
			Type: historicalcontract.
				ScopeTypeGlobal,
		},

		DatasetLimit:       100,
		MaximumBucketCount: 100,
		MaximumWindowCount: 100,
		GeneratedAt:        replayNow(),
	}
}

func validReplayOutcome(
	t *testing.T,
	request historicalmaterialization.Request,
) historicalmaterialization.Outcome {
	t.Helper()
	currentPlan, err := historicalwindow.Build(
		context.Background(),
		historicalwindow.Request{
			StartTime: request.StartTime,
			EndTime:   request.EndTime,
			AsOfTime:  request.AsOfTime,
			Granularity: request.
				Granularity,
			MaximumBucketCount: 1,
		},
	)
	if err != nil {
		t.Fatalf(
			"build current plan: %v",
			err,
		)
	}
	if currentPlan.EffectiveWindow == nil ||
		currentPlan.PreviousWindow == nil {
		t.Fatalf(
			"incomplete current plan: %#v",
			currentPlan,
		)
	}
	previousPlan, err := historicalwindow.Build(
		context.Background(),
		historicalwindow.Request{
			StartTime: currentPlan.
				PreviousWindow.StartTime,
			EndTime: currentPlan.
				PreviousWindow.EndTime,
			AsOfTime: request.AsOfTime,
			Granularity: request.
				Granularity,
			MaximumBucketCount: 1,
		},
	)
	if err != nil {
		t.Fatalf(
			"build previous plan: %v",
			err,
		)
	}

	previous := validPeriodResult(
		t,
		*currentPlan.PreviousWindow,
		request,
	)
	current := validPeriodResult(
		t,
		*currentPlan.EffectiveWindow,
		request,
	)
	compared, err := historicalcomparison.Attach(
		current,
		previous,
	)
	if err != nil {
		t.Fatalf(
			"attach comparison: %v",
			err,
		)
	}

	record := historicalaggregate.Record{
		ID: "historical-aggregate-record-" +
			strings.TrimPrefix(
				compared.Provenance.
					InputFingerprint,
				"sha256:",
			),
		Key: historicalaggregate.ResultKey{
			SchemaVersion: compared.SchemaVersion,
			MetricName:    compared.Metric.Name,
			Scope:         compared.Scope,
			Granularity:   compared.Granularity,
			Window:        compared.Window,
		},
		InputFingerprint: compared.
			Provenance.InputFingerprint,
		Result:   compared.Clone(),
		StoredAt: request.GeneratedAt.UTC(),
	}
	return historicalmaterialization.Outcome{
		Version: historicalmaterialization.Version,
		Plan:    currentPlan.Clone(),
		PreviousPlan: previousPlan.
			Clone(),
		ReadSummaries: historicalmaterialization.
			PeriodReadSummaries{
			Previous: historicalmaterialization.
				ReadSummary{
				Window: previous.Window,
				IsolationLevel: historicalread.
					SnapshotIsolationRepeatableRead,
				DatasetLimit: request.DatasetLimit,
			},
			Current: historicalmaterialization.
				ReadSummary{
				Window: current.Window,
				IsolationLevel: historicalread.
					SnapshotIsolationRepeatableRead,
				DatasetLimit: request.DatasetLimit,
			},
		},
		CurrentPeriodResult: current.Clone(),
		CurrentResult:       compared.Clone(),
		PreviousResult:      previous.Clone(),
		Record:              record.Clone(),
	}
}

func validPeriodResult(
	t *testing.T,
	window historicalcontract.TimeWindow,
	request historicalmaterialization.Request,
) historicalcontract.Result {
	t.Helper()
	specification, exists :=
		historicalcontract.MetricSpecFor(
			request.MetricName,
		)
	if !exists {
		t.Fatalf(
			"metric specification is absent: %s",
			request.MetricName,
		)
	}
	value := float64(
		window.StartTime.Hour() + 1,
	)
	confidence := historicalcontract.Confidence{
		Score:       1,
		Level:       historicalcontract.ConfidenceLevelHigh,
		SampleCount: 1,
		Reasons: []historicalcontract.ConfidenceReason{
			{
				Code:         "complete_coverage",
				Message:      "Complete deterministic test coverage.",
				Contribution: 1,
			},
		},
	}
	points := []historicalcontract.Point{
		{
			StartTime:     window.StartTime.UTC(),
			EndTime:       window.EndTime.UTC(),
			Status:        historicalcontract.BucketStatusComplete,
			Value:         value,
			SampleCount:   1,
			CoverageRatio: 1,
			Confidence:    confidence,
		},
	}
	return historicalcontract.Result{
		SchemaVersion: historicalcontract.SchemaVersionV1,
		Status:        historicalcontract.SeriesStatusComplete,
		Metric: historicalcontract.Metric{
			Name:        specification.Name,
			Unit:        specification.Unit,
			Aggregation: specification.Aggregation,
		},
		Scope:       request.Scope,
		Window:      window,
		Granularity: request.Granularity,
		Points:      points,
		Summary: historicalcontract.Summarize(
			points,
		),
		Confidence: confidence,
		Provenance: historicalcontract.Provenance{
			BuilderVersion: "historical-replay-test-builder-v1",
			InputFingerprint: periodFingerprint(
				window,
				request.MetricName,
				request.Scope,
			),
			SourceNames: []string{
				"historical_replay_test",
			},
			LatestSourceUpdatedAt: window.EndTime.UTC(),
		},
		GeneratedAt: request.GeneratedAt.UTC(),
	}
}

func periodFingerprint(
	window historicalcontract.TimeWindow,
	metric historicalcontract.MetricName,
	scope historicalcontract.Scope,
) string {
	sum := sha256.Sum256(
		[]byte(
			fmt.Sprintf(
				"%s\n%s\n%s\n%s\n%s\n%s",
				window.StartTime.UTC().
					Format(time.RFC3339Nano),
				window.EndTime.UTC().
					Format(time.RFC3339Nano),
				window.AsOfTime.UTC().
					Format(time.RFC3339Nano),
				metric,
				scope.Type,
				scope.AirportICAOCode,
			),
		),
	)
	return "sha256:" +
		hex.EncodeToString(sum[:])
}

func replayAsOfTime() time.Time {
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

func replayNow() time.Time {
	return replayAsOfTime().Add(time.Hour)
}
