package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrReconciliationRepositoryPoolRequired = errors.New(
	"reconciliation repository pool is required",
)

type ReconciliationRepository struct {
	db *pgxpool.Pool
}

var _ reconciliation.Repository = (*ReconciliationRepository)(nil)

func NewReconciliationRepository(
	db *pgxpool.Pool,
) *ReconciliationRepository {
	return &ReconciliationRepository{
		db: db,
	}
}

func (repository *ReconciliationRepository) MarkPendingDerivation(
	ctx context.Context,
	task reconciliation.PendingDerivation,
) error {
	if err := repository.validatePool(); err != nil {
		return err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	normalized := task.Normalize()
	if err := normalized.Validate(); err != nil {
		return err
	}

	const query = `
		INSERT INTO derived_reconciliation_tasks AS existing (
			deduplication_key,
			ingestion_run_id,
			icao24,
			derivation_type,
			status,
			observed_from,
			observed_to,
			last_error,
			next_attempt_at,
			signal_version
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			'pending',
			$5,
			$6,
			$7,
			now(),
			1
		)
		ON CONFLICT (deduplication_key)
		DO UPDATE SET
			status = CASE
				WHEN existing.status = 'processing'
					THEN existing.status
				ELSE 'pending'
			END,
			observed_from = EXCLUDED.observed_from,
			observed_to = EXCLUDED.observed_to,
			last_error = EXCLUDED.last_error,
			next_attempt_at = CASE
				WHEN existing.status = 'processing'
					THEN existing.next_attempt_at
				ELSE now()
			END,
			signal_version = existing.signal_version + 1,
			processing_started_at = CASE
				WHEN existing.status = 'processing'
					THEN existing.processing_started_at
				ELSE NULL
			END,
			claimed_signal_version = CASE
				WHEN existing.status = 'processing'
					THEN existing.claimed_signal_version
				ELSE NULL
			END,
			updated_at = now(),
			completed_at = NULL;
	`

	_, err := repository.db.Exec(
		ctx,
		query,
		normalized.DeduplicationKey(),
		nullableUUID(normalized.IngestionRunID),
		normalized.ICAO24,
		string(normalized.DerivationType),
		normalized.ObservedFrom,
		normalized.ObservedTo,
		normalized.LastError,
	)
	if err != nil {
		return fmt.Errorf(
			"insert pending derived reconciliation task: %w",
			err,
		)
	}

	return nil
}

func (repository *ReconciliationRepository) ClaimNextAvailable(
	ctx context.Context,
) (reconciliation.Task, error) {
	if err := repository.validatePool(); err != nil {
		return reconciliation.Task{}, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		WITH candidate AS (
			SELECT id
			FROM derived_reconciliation_tasks
			WHERE status = 'pending'
				AND next_attempt_at <= now()
			ORDER BY
				next_attempt_at ASC,
				created_at ASC,
				id ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE derived_reconciliation_tasks AS task
		SET
			status = 'processing',
			attempt_count = task.attempt_count + 1,
			processing_started_at = now(),
			claimed_signal_version = task.signal_version,
			updated_at = now(),
			completed_at = NULL
		FROM candidate
		WHERE task.id = candidate.id
		RETURNING
			task.id::text,
			task.deduplication_key,
			COALESCE(task.ingestion_run_id::text, ''),
			task.icao24,
			task.derivation_type,
			task.status,
			task.observed_from,
			task.observed_to,
			task.attempt_count,
			task.signal_version,
			COALESCE(task.claimed_signal_version, 0),
			task.last_error,
			task.next_attempt_at,
			task.processing_started_at,
			task.created_at,
			task.updated_at,
			task.completed_at;
	`

	task, err := scanReconciliationTask(
		repository.db.QueryRow(
			ctx,
			query,
		),
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return reconciliation.Task{},
			reconciliation.ErrNoTaskAvailable
	}
	if err != nil {
		return reconciliation.Task{}, fmt.Errorf(
			"claim next available reconciliation task: %w",
			err,
		)
	}

	return task, nil
}

// MarkCompleted completes the claimed attempt. If a newer persistence failure
// signalled the same task while it was processing, the task is atomically
// returned to pending instead of losing that newer signal.
func (repository *ReconciliationRepository) MarkCompleted(
	ctx context.Context,
	taskID string,
	attemptCount int,
) (reconciliation.TaskStatus, error) {
	if err := repository.validatePool(); err != nil {
		return "", err
	}

	normalizedTaskID := reconciliation.NormalizeTaskID(taskID)
	if normalizedTaskID == "" {
		return "", reconciliation.ErrTaskIDRequired
	}

	if attemptCount <= 0 {
		return "", reconciliation.ErrAttemptCountInvalid
	}

	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		UPDATE derived_reconciliation_tasks
		SET
			status = CASE
				WHEN signal_version > claimed_signal_version
					THEN 'pending'
				ELSE 'completed'
			END,
			last_error = CASE
				WHEN signal_version > claimed_signal_version
					THEN last_error
				ELSE ''
			END,
			next_attempt_at = CASE
				WHEN signal_version > claimed_signal_version
					THEN now()
				ELSE next_attempt_at
			END,
			processing_started_at = NULL,
			claimed_signal_version = NULL,
			updated_at = now(),
			completed_at = CASE
				WHEN signal_version > claimed_signal_version
					THEN NULL
				ELSE now()
			END
		WHERE id = $1
			AND status = 'processing'
			AND attempt_count = $2
		RETURNING status;
	`

	var status string

	err := repository.db.QueryRow(
		ctx,
		query,
		normalizedTaskID,
		attemptCount,
	).Scan(
		&status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", reconciliation.ErrTaskTransitionRejected
	}
	if err != nil {
		return "", fmt.Errorf(
			"mark reconciliation task completed: %w",
			err,
		)
	}

	return reconciliation.TaskStatus(status), nil
}

func (repository *ReconciliationRepository) MarkRetry(
	ctx context.Context,
	taskID string,
	attemptCount int,
	nextAttemptAt time.Time,
	lastError string,
) error {
	if err := repository.validatePool(); err != nil {
		return err
	}

	normalizedTaskID := reconciliation.NormalizeTaskID(taskID)
	if normalizedTaskID == "" {
		return reconciliation.ErrTaskIDRequired
	}

	if attemptCount <= 0 {
		return reconciliation.ErrAttemptCountInvalid
	}

	if nextAttemptAt.IsZero() {
		return reconciliation.ErrNextAttemptAtRequired
	}

	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		UPDATE derived_reconciliation_tasks
		SET
			status = 'pending',
			last_error = CASE
				WHEN signal_version > claimed_signal_version
					THEN last_error
				ELSE $3
			END,
			next_attempt_at = CASE
				WHEN signal_version > claimed_signal_version
					THEN now()
				ELSE $4
			END,
			processing_started_at = NULL,
			claimed_signal_version = NULL,
			updated_at = now(),
			completed_at = NULL
		WHERE id = $1
			AND status = 'processing'
			AND attempt_count = $2;
	`

	commandTag, err := repository.db.Exec(
		ctx,
		query,
		normalizedTaskID,
		attemptCount,
		reconciliation.NormalizeLastError(lastError),
		nextAttemptAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf(
			"mark reconciliation task for retry: %w",
			err,
		)
	}

	if commandTag.RowsAffected() == 0 {
		return reconciliation.ErrTaskTransitionRejected
	}

	return nil
}

// MarkFailed records a terminal failure for the claimed attempt. A newer
// persistence signal wins over the terminal decision and requeues the task.
func (repository *ReconciliationRepository) MarkFailed(
	ctx context.Context,
	taskID string,
	attemptCount int,
	lastError string,
) (reconciliation.TaskStatus, error) {
	if err := repository.validatePool(); err != nil {
		return "", err
	}

	normalizedTaskID := reconciliation.NormalizeTaskID(taskID)
	if normalizedTaskID == "" {
		return "", reconciliation.ErrTaskIDRequired
	}

	if attemptCount <= 0 {
		return "", reconciliation.ErrAttemptCountInvalid
	}

	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		UPDATE derived_reconciliation_tasks
		SET
			status = CASE
				WHEN signal_version > claimed_signal_version
					THEN 'pending'
				ELSE 'failed'
			END,
			last_error = CASE
				WHEN signal_version > claimed_signal_version
					THEN last_error
				ELSE $3
			END,
			next_attempt_at = CASE
				WHEN signal_version > claimed_signal_version
					THEN now()
				ELSE next_attempt_at
			END,
			processing_started_at = NULL,
			claimed_signal_version = NULL,
			updated_at = now(),
			completed_at = NULL
		WHERE id = $1
			AND status = 'processing'
			AND attempt_count = $2
		RETURNING status;
	`

	var status string

	err := repository.db.QueryRow(
		ctx,
		query,
		normalizedTaskID,
		attemptCount,
		reconciliation.NormalizeLastError(lastError),
	).Scan(
		&status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", reconciliation.ErrTaskTransitionRejected
	}
	if err != nil {
		return "", fmt.Errorf(
			"mark reconciliation task failed: %w",
			err,
		)
	}

	return reconciliation.TaskStatus(status), nil
}

// RequeueStaleProcessing recovers tasks abandoned by a crashed worker.
// The caller chooses the lease boundary and must use a value old enough that a
// healthy worker cannot still own the task.
func (repository *ReconciliationRepository) RequeueStaleProcessing(
	ctx context.Context,
	staleBefore time.Time,
) (int64, error) {
	if err := repository.validatePool(); err != nil {
		return 0, err
	}

	if staleBefore.IsZero() {
		return 0, reconciliation.ErrStaleBeforeRequired
	}

	if ctx == nil {
		ctx = context.Background()
	}

	const query = `
		UPDATE derived_reconciliation_tasks
		SET
			status = 'pending',
			next_attempt_at = now(),
			processing_started_at = NULL,
			claimed_signal_version = NULL,
			updated_at = now(),
			completed_at = NULL
		WHERE status = 'processing'
			AND processing_started_at <= $1;
	`

	commandTag, err := repository.db.Exec(
		ctx,
		query,
		staleBefore.UTC(),
	)
	if err != nil {
		return 0, fmt.Errorf(
			"requeue stale processing reconciliation tasks: %w",
			err,
		)
	}

	return commandTag.RowsAffected(), nil
}

func (repository *ReconciliationRepository) validatePool() error {
	if repository == nil || repository.db == nil {
		return ErrReconciliationRepositoryPoolRequired
	}

	return nil
}

func scanReconciliationTask(
	row pgx.Row,
) (reconciliation.Task, error) {
	var task reconciliation.Task
	var derivationType string
	var status string
	var processingStartedAt pgtype.Timestamptz
	var completedAt pgtype.Timestamptz

	err := row.Scan(
		&task.ID,
		&task.DeduplicationKey,
		&task.IngestionRunID,
		&task.ICAO24,
		&derivationType,
		&status,
		&task.ObservedFrom,
		&task.ObservedTo,
		&task.AttemptCount,
		&task.SignalVersion,
		&task.ClaimedSignalVersion,
		&task.LastError,
		&task.NextAttemptAt,
		&processingStartedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
		&completedAt,
	)
	if err != nil {
		return reconciliation.Task{}, err
	}

	task.DerivationType = reconciliation.DerivationType(
		derivationType,
	)
	task.Status = reconciliation.TaskStatus(
		status,
	)
	task.ObservedFrom = task.ObservedFrom.UTC()
	task.ObservedTo = task.ObservedTo.UTC()
	task.NextAttemptAt = task.NextAttemptAt.UTC()
	task.CreatedAt = task.CreatedAt.UTC()
	task.UpdatedAt = task.UpdatedAt.UTC()

	if processingStartedAt.Valid {
		value := processingStartedAt.Time.UTC()
		task.ProcessingStartedAt = &value
	}

	if completedAt.Valid {
		value := completedAt.Time.UTC()
		task.CompletedAt = &value
	}

	return task, nil
}
