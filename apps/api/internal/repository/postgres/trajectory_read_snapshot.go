package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

const trajectoryReadRollbackTimeout = 5 * time.Second

var (
	errTrajectoryReadOperationRequired = errors.New("trajectory read operation is required")
	errTrajectoryReadRepositoryNil     = errors.New("trajectory repository is required")
)

type trajectoryReadOperation func(
	*TrajectoryRepository,
) (trajectory.FlightTrajectory, error)

func (
	repository *TrajectoryRepository,
) withTrajectoryReadSnapshot(
	ctx context.Context,
	operation trajectoryReadOperation,
) (trajectory.FlightTrajectory, error) {
	if err := requireRepositoryContext(ctx); err != nil {
		return trajectory.FlightTrajectory{}, err
	}
	if repository == nil {
		return trajectory.FlightTrajectory{}, errTrajectoryReadRepositoryNil
	}
	if operation == nil {
		return trajectory.FlightTrajectory{}, errTrajectoryReadOperationRequired
	}

	// A transaction-bound read repository already belongs to a caller-owned
	// snapshot. Reuse it rather than opening a nested transaction.
	if repository.db == nil {
		return operation(repository)
	}

	tx, err := repository.db.BeginTx(
		ctx,
		pgx.TxOptions{
			IsoLevel:   pgx.RepeatableRead,
			AccessMode: pgx.ReadOnly,
		},
	)
	if err != nil {
		return trajectory.FlightTrajectory{}, fmt.Errorf(
			"begin trajectory read snapshot: %w",
			err,
		)
	}

	committed := false
	defer func() {
		if !committed {
			rollbackTrajectoryReadSnapshot(tx)
		}
	}()

	snapshotRepository := NewTrajectoryReadRepository(tx)
	item, err := operation(snapshotRepository)
	if err != nil {
		return trajectory.FlightTrajectory{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return trajectory.FlightTrajectory{}, fmt.Errorf(
			"commit trajectory read snapshot: %w",
			err,
		)
	}

	committed = true
	return item, nil
}

func rollbackTrajectoryReadSnapshot(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		trajectoryReadRollbackTimeout,
	)
	defer cancel()

	_ = tx.Rollback(ctx)
}
