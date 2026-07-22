package reconciliation

import (
	"context"
	"time"
)

type DerivationWriter interface {
	MarkPendingDerivation(ctx context.Context, task PendingDerivation) error
}

type TaskClaimer interface {
	ClaimNextAvailable(ctx context.Context) (Task, error)
}

type TaskTransitionWriter interface {
	MarkCompleted(ctx context.Context, taskID string, attemptCount int) (TaskStatus, error)
	MarkRetry(ctx context.Context, taskID string, attemptCount int, nextAttemptAt time.Time, lastError string) error
	MarkFailed(ctx context.Context, taskID string, attemptCount int, lastError string) (TaskStatus, error)
}

type StaleTaskRequeuer interface {
	RequeueStaleProcessing(ctx context.Context, staleBefore time.Time) (int64, error)
}

type Repository interface {
	DerivationWriter
	TaskClaimer
	TaskTransitionWriter
	StaleTaskRequeuer
}
