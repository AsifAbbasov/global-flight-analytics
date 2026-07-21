package postgres

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

func (repository *TrajectoryRepository) SaveTrajectory(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) error {
	return repository.saveTrajectory(
		ctx,
		newLiveTrajectoryWriteRequest(item),
	)
}

func (repository *TrajectoryRepository) SaveReconciledTrajectory(
	ctx context.Context,
	taskID string,
	attemptCount int,
	item trajectory.FlightTrajectory,
) error {
	request, err := newReconciledTrajectoryWriteRequest(
		taskID,
		attemptCount,
		item,
	)
	if err != nil {
		return err
	}
	return repository.saveTrajectory(ctx, request)
}

func (repository *TrajectoryRepository) saveTrajectory(
	ctx context.Context,
	request trajectoryWriteRequest,
) error {
	if err := requireRepositoryContext(ctx); err != nil {
		return err
	}
	if err := request.validate(); err != nil {
		return err
	}

	item := request.item
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

	tx, err := repository.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			rollbackRepositoryTransaction(tx)
		}
	}()

	if request.isReconciled() {
		if err := assertReconciliationAttemptOwned(
			ctx,
			tx,
			request.reconciliationTaskID,
			request.attemptCount,
		); err != nil {
			return err
		}
		if err := deleteExistingReconciledTrajectory(
			ctx,
			tx,
			request.reconciliationTaskID,
		); err != nil {
			return err
		}
	}

	var trajectoryID string
	switch request.mode {
	case trajectoryWriteModeLive:
		trajectoryID, err = repository.insertFlightTrajectory(
			ctx,
			tx,
			item,
		)
	case trajectoryWriteModeReconciled:
		trajectoryID, err = repository.insertReconciledFlightTrajectory(
			ctx,
			tx,
			request.reconciliationTaskID,
			item,
		)
	default:
		return fmt.Errorf(
			"%w: got %d",
			ErrTrajectoryWriteModeInvalid,
			request.mode,
		)
	}
	if err != nil {
		return fmt.Errorf("insert flight trajectory: %w", err)
	}

	segmentIDs, err := repository.insertTrajectorySegments(
		ctx,
		tx,
		trajectoryID,
		item.Segments,
	)
	if err != nil {
		return fmt.Errorf("insert trajectory segments: %w", err)
	}

	if err := repository.insertCoverageGaps(
		ctx,
		tx,
		trajectoryID,
		segmentIDs,
		item.CoverageGaps,
	); err != nil {
		return fmt.Errorf("insert coverage gaps: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}
