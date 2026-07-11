package reconciliation

import (
	"context"
	"time"
)

type Repository interface {
	MarkPendingDerivation(
		ctx context.Context,
		task PendingDerivation,
	) error
	ClaimNextAvailable(
		ctx context.Context,
	) (Task, error)
	MarkCompleted(
		ctx context.Context,
		taskID string,
		attemptCount int,
	) (TaskStatus, error)
	MarkRetry(
		ctx context.Context,
		taskID string,
		attemptCount int,
		nextAttemptAt time.Time,
		lastError string,
	) error
	MarkFailed(
		ctx context.Context,
		taskID string,
		attemptCount int,
		lastError string,
	) (TaskStatus, error)
	RequeueStaleProcessing(
		ctx context.Context,
		staleBefore time.Time,
	) (int64, error)
}
