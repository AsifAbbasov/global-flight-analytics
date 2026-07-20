package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/jackc/pgx/v5"
)

func assertReconciliationAttemptOwned(
	ctx context.Context,
	tx pgx.Tx,
	taskID string,
	attemptCount int,
) error {
	const query = `
		SELECT id::text
		FROM derived_reconciliation_tasks
		WHERE id = $1
			AND status = 'processing'
			AND attempt_count = $2
		FOR UPDATE;
	`

	var ownedTaskID string

	err := tx.QueryRow(
		ctx,
		query,
		taskID,
		attemptCount,
	).Scan(
		&ownedTaskID,
	)
	if errors.Is(
		err,
		pgx.ErrNoRows,
	) {
		return reconciliation.ErrTaskTransitionRejected
	}
	if err != nil {
		return fmt.Errorf(
			"verify reconciliation trajectory attempt ownership: %w",
			err,
		)
	}

	return nil
}
func deleteExistingReconciledTrajectory(
	ctx context.Context,
	tx pgx.Tx,
	taskID string,
) error {
	if _, err := tx.Exec(
		ctx,
		`
			DELETE FROM flight_trajectories
			WHERE reconciliation_task_id = $1;
		`,
		taskID,
	); err != nil {
		return fmt.Errorf(
			"delete existing reconciled trajectory: %w",
			err,
		)
	}

	return nil
}
