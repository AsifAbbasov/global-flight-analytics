package postgres

import (
	"context"
	"fmt"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

func (repository *TrajectoryRepository) SaveTrajectory(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) error {
	return repository.saveTrajectory(
		ctx,
		"",
		0,
		item,
	)
}
func (repository *TrajectoryRepository) SaveReconciledTrajectory(
	ctx context.Context,
	taskID string,
	attemptCount int,
	item trajectory.FlightTrajectory,
) error {
	normalizedTaskID := reconciliation.NormalizeTaskID(
		taskID,
	)
	if normalizedTaskID == "" {
		return reconciliation.ErrTaskIDRequired
	}

	if attemptCount <= 0 {
		return reconciliation.ErrAttemptCountInvalid
	}

	return repository.saveTrajectory(
		ctx,
		normalizedTaskID,
		attemptCount,
		item,
	)
}
func (repository *TrajectoryRepository) saveTrajectory(
	ctx context.Context,
	reconciliationTaskID string,
	attemptCount int,
	item trajectory.FlightTrajectory,
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := validateTrajectoryRelationalIntegrity(item); err != nil {
		return fmt.Errorf(
			"validate trajectory relational integrity: %w",
			err,
		)
	}

	if err := validatePersistedFlightIdentity(item); err != nil {
		return fmt.Errorf(
			"validate flight identity metadata: %w",
			err,
		)
	}

	tx, err := repository.db.BeginTx(
		ctx,
		pgx.TxOptions{},
	)
	if err != nil {
		return err
	}

	committed := false

	defer func() {
		if !committed {
			_ = tx.Rollback(
				ctx,
			)
		}
	}()

	if reconciliationTaskID != "" {
		if err := assertReconciliationAttemptOwned(
			ctx,
			tx,
			reconciliationTaskID,
			attemptCount,
		); err != nil {
			return err
		}
	}

	if reconciliationTaskID != "" {
		if err := deleteExistingReconciledTrajectory(
			ctx,
			tx,
			reconciliationTaskID,
		); err != nil {
			return err
		}
	}

	var trajectoryID string

	if reconciliationTaskID == "" {
		trajectoryID, err = repository.insertFlightTrajectory(
			ctx,
			tx,
			item,
		)
	} else {
		trajectoryID, err = repository.insertReconciledFlightTrajectory(
			ctx,
			tx,
			reconciliationTaskID,
			item,
		)
	}
	if err != nil {
		return fmt.Errorf(
			"insert flight trajectory: %w",
			err,
		)
	}

	segmentIDs, err := repository.insertTrajectorySegments(
		ctx,
		tx,
		trajectoryID,
		item.Segments,
	)
	if err != nil {
		return fmt.Errorf(
			"insert trajectory segments: %w",
			err,
		)
	}

	if err := repository.insertCoverageGaps(
		ctx,
		tx,
		trajectoryID,
		segmentIDs,
		item.CoverageGaps,
	); err != nil {
		return fmt.Errorf(
			"insert coverage gaps: %w",
			err,
		)
	}

	if err := tx.Commit(
		ctx,
	); err != nil {
		return err
	}

	committed = true

	return nil
}
