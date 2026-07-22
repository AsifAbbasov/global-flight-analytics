package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
)

func TestReconciliationRepositoryClaimsPendingTask(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	pending := makeLifecyclePendingDerivation(
		"ABC123",
		"initial failure",
	)

	if err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		pending,
	); err != nil {
		t.Fatalf(
			"mark pending derivation: %v",
			err,
		)
	}

	task, err := fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"claim pending derivation: %v",
			err,
		)
	}

	if task.ID == "" {
		t.Fatal(
			"expected claimed task id",
		)
	}

	if task.Status != reconciliation.TaskStatusProcessing {
		t.Fatalf(
			"expected processing status, got %s",
			task.Status,
		)
	}

	if task.AttemptCount != 1 {
		t.Fatalf(
			"expected attempt count 1, got %d",
			task.AttemptCount,
		)
	}

	if task.SignalVersion != 1 {
		t.Fatalf(
			"expected signal version 1, got %d",
			task.SignalVersion,
		)
	}

	if task.ClaimedSignalVersion != task.SignalVersion {
		t.Fatalf(
			"expected claimed signal version %d, got %d",
			task.SignalVersion,
			task.ClaimedSignalVersion,
		)
	}

	if task.ProcessingStartedAt == nil {
		t.Fatal(
			"expected processing start timestamp",
		)
	}

	_, err = fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if !errors.Is(err, reconciliation.ErrNoTaskAvailable) {
		t.Fatalf(
			"expected ErrNoTaskAvailable, got %v",
			err,
		)
	}
}

func TestReconciliationRepositoryClaimsTaskOnlyOnceConcurrently(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	if err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		makeLifecyclePendingDerivation(
			"ABC123",
			"concurrent claim",
		),
	); err != nil {
		t.Fatalf(
			"mark pending derivation: %v",
			err,
		)
	}

	type claimResult struct {
		task reconciliation.Task
		err  error
	}

	start := make(
		chan struct{},
	)
	results := make(
		chan claimResult,
		2,
	)

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	for index := 0; index < 2; index++ {
		go func() {
			defer waitGroup.Done()

			<-start

			task, err := fixture.repository.ClaimNextAvailable(
				context.Background(),
			)

			results <- claimResult{
				task: task,
				err:  err,
			}
		}()
	}

	close(start)
	waitGroup.Wait()
	close(results)

	successCount := 0
	noTaskCount := 0

	for result := range results {
		switch {
		case result.err == nil:
			successCount++

			if result.task.Status != reconciliation.TaskStatusProcessing {
				t.Fatalf(
					"expected claimed task to be processing, got %s",
					result.task.Status,
				)
			}

		case errors.Is(
			result.err,
			reconciliation.ErrNoTaskAvailable,
		):
			noTaskCount++

		default:
			t.Fatalf(
				"unexpected concurrent claim error: %v",
				result.err,
			)
		}
	}

	if successCount != 1 {
		t.Fatalf(
			"expected exactly 1 successful claim, got %d",
			successCount,
		)
	}

	if noTaskCount != 1 {
		t.Fatalf(
			"expected exactly 1 empty claim, got %d",
			noTaskCount,
		)
	}
}

func TestReconciliationRepositoryMarksClaimedTaskCompleted(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	claimed := markAndClaimLifecycleTask(
		t,
		fixture,
		makeLifecyclePendingDerivation(
			"ABC123",
			"completion",
		),
	)

	status, err := fixture.repository.MarkCompleted(
		context.Background(),
		claimed.ID,
		claimed.AttemptCount,
	)
	if err != nil {
		t.Fatalf(
			"mark task completed: %v",
			err,
		)
	}

	if status != reconciliation.TaskStatusCompleted {
		t.Fatalf(
			"expected completed status, got %s",
			status,
		)
	}

	var storedStatus string
	var completed bool
	var processingCleared bool
	var claimedVersionCleared bool
	var lastError string

	err = fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT
				status,
				completed_at IS NOT NULL,
				processing_started_at IS NULL,
				claimed_signal_version IS NULL,
				last_error
			FROM derived_reconciliation_tasks
			WHERE id = $1;
		`,
		claimed.ID,
	).Scan(
		&storedStatus,
		&completed,
		&processingCleared,
		&claimedVersionCleared,
		&lastError,
	)
	if err != nil {
		t.Fatalf(
			"load completed task: %v",
			err,
		)
	}

	if storedStatus != string(reconciliation.TaskStatusCompleted) {
		t.Fatalf(
			"expected stored completed status, got %s",
			storedStatus,
		)
	}

	if !completed || !processingCleared || !claimedVersionCleared {
		t.Fatalf(
			"unexpected completed lifecycle metadata: completed=%t processing_cleared=%t claimed_version_cleared=%t",
			completed,
			processingCleared,
			claimedVersionCleared,
		)
	}

	if lastError != "" {
		t.Fatalf(
			"expected completed task error to be cleared, got %q",
			lastError,
		)
	}
}

func TestReconciliationRepositoryRequeuesCompletedAttemptWhenNewSignalArrives(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	pending := makeLifecyclePendingDerivation(
		"ABC123",
		"first failure",
	)

	firstClaim := markAndClaimLifecycleTask(
		t,
		fixture,
		pending,
	)

	pending.LastError = "new failure while processing"

	if err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		pending,
	); err != nil {
		t.Fatalf(
			"signal task while processing: %v",
			err,
		)
	}

	status, err := fixture.repository.MarkCompleted(
		context.Background(),
		firstClaim.ID,
		firstClaim.AttemptCount,
	)
	if err != nil {
		t.Fatalf(
			"complete signalled task attempt: %v",
			err,
		)
	}

	if status != reconciliation.TaskStatusPending {
		t.Fatalf(
			"expected task to return to pending, got %s",
			status,
		)
	}

	secondClaim, err := fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"claim requeued task: %v",
			err,
		)
	}

	if secondClaim.ID != firstClaim.ID {
		t.Fatalf(
			"expected requeued task id %s, got %s",
			firstClaim.ID,
			secondClaim.ID,
		)
	}

	if secondClaim.AttemptCount != 2 {
		t.Fatalf(
			"expected attempt count 2, got %d",
			secondClaim.AttemptCount,
		)
	}

	if secondClaim.SignalVersion != 2 {
		t.Fatalf(
			"expected signal version 2, got %d",
			secondClaim.SignalVersion,
		)
	}

	if secondClaim.LastError != "new failure while processing" {
		t.Fatalf(
			"expected latest failure message, got %q",
			secondClaim.LastError,
		)
	}
}

func TestReconciliationRepositorySchedulesRetry(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	claimed := markAndClaimLifecycleTask(
		t,
		fixture,
		makeLifecyclePendingDerivation(
			"ABC123",
			"retry",
		),
	)

	nextAttemptAt := time.Now().UTC().Add(
		time.Hour,
	)

	err := fixture.repository.MarkRetry(
		context.Background(),
		claimed.ID,
		claimed.AttemptCount,
		nextAttemptAt,
		"temporary database failure",
	)
	if err != nil {
		t.Fatalf(
			"mark task for retry: %v",
			err,
		)
	}

	_, err = fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if !errors.Is(err, reconciliation.ErrNoTaskAvailable) {
		t.Fatalf(
			"expected delayed task to be unavailable, got %v",
			err,
		)
	}

	_, err = fixture.pool.Exec(
		context.Background(),
		`
			UPDATE derived_reconciliation_tasks
			SET next_attempt_at = now() - interval '1 second'
			WHERE id = $1;
		`,
		claimed.ID,
	)
	if err != nil {
		t.Fatalf(
			"make retry task available: %v",
			err,
		)
	}

	secondClaim, err := fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"claim retry task: %v",
			err,
		)
	}

	if secondClaim.AttemptCount != 2 {
		t.Fatalf(
			"expected attempt count 2, got %d",
			secondClaim.AttemptCount,
		)
	}

	if secondClaim.LastError != "temporary database failure" {
		t.Fatalf(
			"expected retry error message, got %q",
			secondClaim.LastError,
		)
	}
}

func TestReconciliationRepositoryNewSignalOverridesDelayedRetry(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	pending := makeLifecyclePendingDerivation(
		"ABC123",
		"first failure",
	)

	claimed := markAndClaimLifecycleTask(
		t,
		fixture,
		pending,
	)

	pending.LastError = "new failure while processing"

	if err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		pending,
	); err != nil {
		t.Fatalf(
			"signal task while processing: %v",
			err,
		)
	}

	err := fixture.repository.MarkRetry(
		context.Background(),
		claimed.ID,
		claimed.AttemptCount,
		time.Now().UTC().Add(time.Hour),
		"old attempt retry error",
	)
	if err != nil {
		t.Fatalf(
			"mark signalled task for retry: %v",
			err,
		)
	}

	secondClaim, err := fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"claim immediately requeued task: %v",
			err,
		)
	}

	if secondClaim.LastError != "new failure while processing" {
		t.Fatalf(
			"expected latest signal error, got %q",
			secondClaim.LastError,
		)
	}
}

func TestReconciliationRepositoryMarksClaimedTaskFailed(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	claimed := markAndClaimLifecycleTask(
		t,
		fixture,
		makeLifecyclePendingDerivation(
			"ABC123",
			"terminal",
		),
	)

	status, err := fixture.repository.MarkFailed(
		context.Background(),
		claimed.ID,
		claimed.AttemptCount,
		"unsupported derivation input",
	)
	if err != nil {
		t.Fatalf(
			"mark task failed: %v",
			err,
		)
	}

	if status != reconciliation.TaskStatusFailed {
		t.Fatalf(
			"expected failed status, got %s",
			status,
		)
	}

	_, err = fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if !errors.Is(err, reconciliation.ErrNoTaskAvailable) {
		t.Fatalf(
			"expected failed task to stay unavailable, got %v",
			err,
		)
	}

	var storedStatus string
	var lastError string
	var processingCleared bool

	err = fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT
				status,
				last_error,
				processing_started_at IS NULL
			FROM derived_reconciliation_tasks
			WHERE id = $1;
		`,
		claimed.ID,
	).Scan(
		&storedStatus,
		&lastError,
		&processingCleared,
	)
	if err != nil {
		t.Fatalf(
			"load failed task: %v",
			err,
		)
	}

	if storedStatus != string(reconciliation.TaskStatusFailed) {
		t.Fatalf(
			"expected failed status, got %s",
			storedStatus,
		)
	}

	if lastError != "unsupported derivation input" {
		t.Fatalf(
			"unexpected failed task error: %q",
			lastError,
		)
	}

	if !processingCleared {
		t.Fatal(
			"expected failed task processing timestamp to be cleared",
		)
	}
}

func TestReconciliationRepositoryNewSignalOverridesTerminalFailure(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	pending := makeLifecyclePendingDerivation(
		"ABC123",
		"first failure",
	)

	claimed := markAndClaimLifecycleTask(
		t,
		fixture,
		pending,
	)

	pending.LastError = "new failure while processing"

	if err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		pending,
	); err != nil {
		t.Fatalf(
			"signal task while processing: %v",
			err,
		)
	}

	status, err := fixture.repository.MarkFailed(
		context.Background(),
		claimed.ID,
		claimed.AttemptCount,
		"old attempt terminal error",
	)
	if err != nil {
		t.Fatalf(
			"mark signalled task failed: %v",
			err,
		)
	}

	if status != reconciliation.TaskStatusPending {
		t.Fatalf(
			"expected newer signal to requeue task, got %s",
			status,
		)
	}

	secondClaim, err := fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"claim requeued task: %v",
			err,
		)
	}

	if secondClaim.LastError != "new failure while processing" {
		t.Fatalf(
			"expected latest signal error, got %q",
			secondClaim.LastError,
		)
	}
}

func TestReconciliationRepositoryRequeuesStaleProcessingTask(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	claimed := markAndClaimLifecycleTask(
		t,
		fixture,
		makeLifecyclePendingDerivation(
			"ABC123",
			"stale worker",
		),
	)

	staleStartedAt := time.Now().UTC().Add(
		-2 * time.Hour,
	)

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			UPDATE derived_reconciliation_tasks
			SET processing_started_at = $2
			WHERE id = $1;
		`,
		claimed.ID,
		staleStartedAt,
	)
	if err != nil {
		t.Fatalf(
			"make processing task stale: %v",
			err,
		)
	}

	requeuedCount, err := fixture.repository.RequeueStaleProcessing(
		context.Background(),
		time.Now().UTC().Add(-time.Hour),
	)
	if err != nil {
		t.Fatalf(
			"requeue stale processing task: %v",
			err,
		)
	}

	if requeuedCount != 1 {
		t.Fatalf(
			"expected 1 requeued stale task, got %d",
			requeuedCount,
		)
	}

	secondClaim, err := fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"claim recovered stale task: %v",
			err,
		)
	}

	if secondClaim.ID != claimed.ID {
		t.Fatalf(
			"expected recovered task id %s, got %s",
			claimed.ID,
			secondClaim.ID,
		)
	}

	if secondClaim.AttemptCount != 2 {
		t.Fatalf(
			"expected attempt count 2, got %d",
			secondClaim.AttemptCount,
		)
	}
}

func TestReconciliationRepositoryDoesNotRequeueFreshProcessingTask(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	markAndClaimLifecycleTask(
		t,
		fixture,
		makeLifecyclePendingDerivation(
			"ABC123",
			"fresh worker",
		),
	)

	requeuedCount, err := fixture.repository.RequeueStaleProcessing(
		context.Background(),
		time.Now().UTC().Add(-time.Hour),
	)
	if err != nil {
		t.Fatalf(
			"requeue stale processing tasks: %v",
			err,
		)
	}

	if requeuedCount != 0 {
		t.Fatalf(
			"expected no fresh task to be requeued, got %d",
			requeuedCount,
		)
	}
}

func TestReconciliationRepositoryRejectsCompletionFromStaleAttempt(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	firstClaim := markAndClaimLifecycleTask(
		t,
		fixture,
		makeLifecyclePendingDerivation(
			"ABC123",
			"stale ownership",
		),
	)

	staleStartedAt := time.Now().UTC().Add(
		-2 * time.Hour,
	)

	_, err := fixture.pool.Exec(
		context.Background(),
		`
			UPDATE derived_reconciliation_tasks
			SET processing_started_at = $2
			WHERE id = $1;
		`,
		firstClaim.ID,
		staleStartedAt,
	)
	if err != nil {
		t.Fatalf(
			"make first attempt stale: %v",
			err,
		)
	}

	_, err = fixture.repository.RequeueStaleProcessing(
		context.Background(),
		time.Now().UTC().Add(-time.Hour),
	)
	if err != nil {
		t.Fatalf(
			"requeue stale attempt: %v",
			err,
		)
	}

	secondClaim, err := fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"claim recovered task: %v",
			err,
		)
	}

	if secondClaim.AttemptCount != firstClaim.AttemptCount+1 {
		t.Fatalf(
			"expected next attempt count %d, got %d",
			firstClaim.AttemptCount+1,
			secondClaim.AttemptCount,
		)
	}

	_, err = fixture.repository.MarkCompleted(
		context.Background(),
		firstClaim.ID,
		firstClaim.AttemptCount,
	)
	if !errors.Is(
		err,
		reconciliation.ErrTaskTransitionRejected,
	) {
		t.Fatalf(
			"expected stale attempt completion rejection, got %v",
			err,
		)
	}

	status, err := fixture.repository.MarkCompleted(
		context.Background(),
		secondClaim.ID,
		secondClaim.AttemptCount,
	)
	if err != nil {
		t.Fatalf(
			"complete current attempt: %v",
			err,
		)
	}

	if status != reconciliation.TaskStatusCompleted {
		t.Fatalf(
			"expected current attempt to complete, got %s",
			status,
		)
	}
}

func TestReconciliationRepositoryRejectsInvalidLifecycleTransitions(
	t *testing.T,
) {
	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	pending := makeLifecyclePendingDerivation(
		"ABC123",
		"transition guard",
	)

	if err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		pending,
	); err != nil {
		t.Fatalf(
			"mark pending derivation: %v",
			err,
		)
	}

	var taskID string

	pendingDeduplicationKey, err := pending.DeduplicationKey()
	if err != nil {
		t.Fatalf("build reconciliation deduplication key: %v", err)
	}

	err = fixture.pool.QueryRow(
		context.Background(),
		`
			SELECT id::text
			FROM derived_reconciliation_tasks
			WHERE deduplication_key = $1;
		`,
		pendingDeduplicationKey,
	).Scan(
		&taskID,
	)
	if err != nil {
		t.Fatalf(
			"load pending task id: %v",
			err,
		)
	}

	_, err = fixture.repository.MarkCompleted(
		context.Background(),
		taskID,
		1,
	)
	if !errors.Is(
		err,
		reconciliation.ErrTaskTransitionRejected,
	) {
		t.Fatalf(
			"expected transition rejection, got %v",
			err,
		)
	}

	err = fixture.repository.MarkRetry(
		context.Background(),
		taskID,
		1,
		time.Now().UTC(),
		"not processing",
	)
	if !errors.Is(
		err,
		reconciliation.ErrTaskTransitionRejected,
	) {
		t.Fatalf(
			"expected retry transition rejection, got %v",
			err,
		)
	}

	_, err = fixture.repository.MarkFailed(
		context.Background(),
		taskID,
		1,
		"not processing",
	)
	if !errors.Is(
		err,
		reconciliation.ErrTaskTransitionRejected,
	) {
		t.Fatalf(
			"expected failed transition rejection, got %v",
			err,
		)
	}
}

func TestReconciliationRepositoryValidatesLifecycleInputs(
	t *testing.T,
) {
	repository := NewReconciliationRepository(
		nil,
	)

	_, err := repository.MarkCompleted(
		context.Background(),
		"",
		0,
	)
	if !errors.Is(
		err,
		ErrReconciliationRepositoryPoolRequired,
	) {
		t.Fatalf(
			"expected pool validation before task id validation, got %v",
			err,
		)
	}

	fixture := newReconciliationFixture(
		t,
	)
	defer fixture.close(
		t,
	)

	_, err = fixture.repository.MarkCompleted(
		context.Background(),
		"   ",
		1,
	)
	if !errors.Is(err, reconciliation.ErrTaskIDRequired) {
		t.Fatalf(
			"expected ErrTaskIDRequired, got %v",
			err,
		)
	}

	err = fixture.repository.MarkRetry(
		context.Background(),
		"task-id",
		1,
		time.Time{},
		"retry",
	)
	if !errors.Is(
		err,
		reconciliation.ErrNextAttemptAtRequired,
	) {
		t.Fatalf(
			"expected ErrNextAttemptAtRequired, got %v",
			err,
		)
	}

	_, err = fixture.repository.MarkFailed(
		context.Background(),
		"   ",
		1,
		"failed",
	)
	if !errors.Is(err, reconciliation.ErrTaskIDRequired) {
		t.Fatalf(
			"expected ErrTaskIDRequired, got %v",
			err,
		)
	}

	_, err = fixture.repository.MarkCompleted(
		context.Background(),
		"task-id",
		0,
	)
	if !errors.Is(
		err,
		reconciliation.ErrAttemptCountInvalid,
	) {
		t.Fatalf(
			"expected ErrAttemptCountInvalid, got %v",
			err,
		)
	}

	_, err = fixture.repository.RequeueStaleProcessing(
		context.Background(),
		time.Time{},
	)
	if !errors.Is(
		err,
		reconciliation.ErrStaleBeforeRequired,
	) {
		t.Fatalf(
			"expected ErrStaleBeforeRequired, got %v",
			err,
		)
	}
}

func makeLifecyclePendingDerivation(
	icao24 string,
	lastError string,
) reconciliation.PendingDerivation {
	observedAt := time.Date(
		2026,
		time.July,
		11,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	return reconciliation.PendingDerivation{
		ICAO24:         icao24,
		DerivationType: reconciliation.DerivationTypeTrajectory,
		ObservedFrom:   observedAt,
		ObservedTo:     observedAt.Add(time.Minute),
		LastError:      lastError,
	}
}

func markAndClaimLifecycleTask(
	t *testing.T,
	fixture *reconciliationFixture,
	pending reconciliation.PendingDerivation,
) reconciliation.Task {
	t.Helper()

	if err := fixture.repository.MarkPendingDerivation(
		context.Background(),
		pending,
	); err != nil {
		t.Fatalf(
			"mark pending lifecycle task: %v",
			err,
		)
	}

	task, err := fixture.repository.ClaimNextAvailable(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"claim lifecycle task: %v",
			err,
		)
	}

	return task
}
