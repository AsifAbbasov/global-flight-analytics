package main

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/reconciliationworker"
)

type reconciliationTaskRunnerFunction func(
	context.Context,
) (reconciliationworker.RunResult, error)

func (function reconciliationTaskRunnerFunction) RunOnce(
	ctx context.Context,
) (reconciliationworker.RunResult, error) {
	return function(ctx)
}

func TestRunReconciliationBatchRejectsNilContextBeforeClaim(
	t *testing.T,
) {
	runCount := 0

	summary, err := runReconciliationBatch(
		nil,
		10,
		reconciliationTaskRunnerFunction(
			func(
				context.Context,
			) (reconciliationworker.RunResult, error) {
				runCount++

				return reconciliationworker.RunResult{}, nil
			},
		),
		nil,
	)

	if !errors.Is(
		err,
		errReconciliationBatchContextRequired,
	) {
		t.Fatalf(
			"error = %v, want context required",
			err,
		)
	}
	if runCount != 0 {
		t.Fatalf(
			"task claims = %d, want 0",
			runCount,
		)
	}
	if summary != (reconciliationBatchSummary{}) {
		t.Fatalf(
			"summary = %+v, want empty summary",
			summary,
		)
	}
}

func TestRunReconciliationBatchStopsCleanlyBeforeClaimOnCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	cancel()

	runCount := 0
	summary, err := runReconciliationBatch(
		ctx,
		10,
		reconciliationTaskRunnerFunction(
			func(
				context.Context,
			) (reconciliationworker.RunResult, error) {
				runCount++

				return reconciliationworker.RunResult{}, nil
			},
		),
		nil,
	)
	if err != nil {
		t.Fatalf(
			"expected clean cancellation, got %v",
			err,
		)
	}

	if runCount != 0 {
		t.Fatalf(
			"expected no task claim after cancellation, got %d calls",
			runCount,
		)
	}

	if summary != (reconciliationBatchSummary{}) {
		t.Fatalf(
			"expected empty summary, got %+v",
			summary,
		)
	}
}

func TestRunReconciliationBatchStopsCleanlyWhenRunOnceObservesCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	runCount := 0
	summary, err := runReconciliationBatch(
		ctx,
		10,
		reconciliationTaskRunnerFunction(
			func(
				context.Context,
			) (reconciliationworker.RunResult, error) {
				runCount++
				cancel()

				return reconciliationworker.RunResult{}, context.Canceled
			},
		),
		nil,
	)
	if err != nil {
		t.Fatalf(
			"expected clean cancellation, got %v",
			err,
		)
	}

	if runCount != 1 {
		t.Fatalf(
			"expected one interrupted task claim, got %d",
			runCount,
		)
	}

	if summary != (reconciliationBatchSummary{}) {
		t.Fatalf(
			"expected empty summary, got %+v",
			summary,
		)
	}
}

func TestRunReconciliationBatchPreservesIndependentDeadlineFailure(
	t *testing.T,
) {
	ctx := context.Background()

	summary, err := runReconciliationBatch(
		ctx,
		10,
		reconciliationTaskRunnerFunction(
			func(
				context.Context,
			) (reconciliationworker.RunResult, error) {
				return reconciliationworker.RunResult{}, context.DeadlineExceeded
			},
		),
		nil,
	)
	if !errors.Is(
		err,
		context.DeadlineExceeded,
	) {
		t.Fatalf(
			"expected independent deadline failure, got %v",
			err,
		)
	}

	if summary != (reconciliationBatchSummary{}) {
		t.Fatalf(
			"expected empty summary, got %+v",
			summary,
		)
	}
}

func TestRunReconciliationBatchPreservesResultAccounting(
	t *testing.T,
) {
	results := []reconciliationworker.RunResult{
		{
			TaskFound:   true,
			FinalStatus: "completed",
		},
		{
			TaskFound:   true,
			FinalStatus: "pending",
		},
		{
			TaskFound:          true,
			FinalStatus:        "pending",
			PersistedItemCount: 1,
		},
		{
			TaskFound:   true,
			FinalStatus: "failed",
		},
		{},
	}

	nextResult := 0
	observedCount := 0
	summary, err := runReconciliationBatch(
		context.Background(),
		10,
		reconciliationTaskRunnerFunction(
			func(
				context.Context,
			) (reconciliationworker.RunResult, error) {
				result := results[nextResult]
				nextResult++

				return result, nil
			},
		),
		func(
			reconciliationworker.RunResult,
		) {
			observedCount++
		},
	)
	if err != nil {
		t.Fatalf(
			"run reconciliation batch: %v",
			err,
		)
	}

	expected := reconciliationBatchSummary{
		ProcessedCount:        4,
		CompletedCount:        1,
		RetryCount:            1,
		FailedCount:           1,
		RequeuedBySignalCount: 1,
	}
	if summary != expected {
		t.Fatalf(
			"expected summary %+v, got %+v",
			expected,
			summary,
		)
	}

	if observedCount != 4 {
		t.Fatalf(
			"expected four observed task results, got %d",
			observedCount,
		)
	}
}
