package projectionread

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type snapshotOperation func(
	postgresClient,
	trajectoryRepository,
) (Snapshot, error)

type snapshotExecutor interface {
	Execute(
		context.Context,
		snapshotOperation,
	) (Snapshot, error)
}

type directSnapshotExecutor struct {
	client     postgresClient
	repository trajectoryRepository
}

func (
	executor directSnapshotExecutor,
) Execute(
	ctx context.Context,
	operation snapshotOperation,
) (Snapshot, error) {
	if operation == nil {
		return Snapshot{},
			fmt.Errorf(
				"Projection Intelligence snapshot operation is required",
			)
	}
	return operation(
		executor.client,
		executor.repository,
	)
}

type snapshotTransaction interface {
	Client() postgresClient
	TrajectoryRepository() trajectoryRepository
	Commit(context.Context) error
	Rollback(context.Context) error
}

type snapshotTransactionStarter interface {
	Begin(
		context.Context,
		pgx.TxOptions,
	) (snapshotTransaction, error)
}

type repeatableReadSnapshotExecutor struct {
	starter snapshotTransactionStarter
}

func (
	executor repeatableReadSnapshotExecutor,
) Execute(
	ctx context.Context,
	operation snapshotOperation,
) (Snapshot, error) {
	if executor.starter == nil {
		return Snapshot{},
			fmt.Errorf(
				"Projection Intelligence snapshot transaction starter is required",
			)
	}
	if operation == nil {
		return Snapshot{},
			fmt.Errorf(
				"Projection Intelligence snapshot operation is required",
			)
	}

	transaction, err := executor.starter.Begin(
		ctx,
		projectionSnapshotTransactionOptions(),
	)
	if err != nil {
		return Snapshot{},
			fmt.Errorf(
				"begin Projection Intelligence read snapshot: %w",
				err,
			)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		rollbackContext, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			5*time.Second,
		)
		defer cancel()
		_ = transaction.Rollback(rollbackContext)
	}()

	snapshot, err := operation(
		transaction.Client(),
		transaction.TrajectoryRepository(),
	)
	if err != nil {
		return Snapshot{}, err
	}

	if err := transaction.Commit(ctx); err != nil {
		return Snapshot{},
			fmt.Errorf(
				"commit Projection Intelligence read snapshot: %w",
				err,
			)
	}
	committed = true

	return snapshot.Clone(), nil
}

func projectionSnapshotTransactionOptions() pgx.TxOptions {
	return pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	}
}

type pgxSnapshotTransactionStarter struct {
	pool *pgxpool.Pool
}

func (
	starter pgxSnapshotTransactionStarter,
) Begin(
	ctx context.Context,
	options pgx.TxOptions,
) (snapshotTransaction, error) {
	transaction, err := starter.pool.BeginTx(
		ctx,
		options,
	)
	if err != nil {
		return nil, err
	}
	return pgxSnapshotTransaction{
		transaction: transaction,
	}, nil
}

type pgxSnapshotTransaction struct {
	transaction pgx.Tx
}

func (
	transaction pgxSnapshotTransaction,
) Client() postgresClient {
	return pgxTxClient{
		transaction: transaction.transaction,
	}
}

func (
	transaction pgxSnapshotTransaction,
) TrajectoryRepository() trajectoryRepository {
	return postgres.NewTrajectoryReadRepository(
		transaction.transaction,
	)
}

func (
	transaction pgxSnapshotTransaction,
) Commit(
	ctx context.Context,
) error {
	return transaction.transaction.Commit(ctx)
}

func (
	transaction pgxSnapshotTransaction,
) Rollback(
	ctx context.Context,
) error {
	return transaction.transaction.Rollback(ctx)
}
