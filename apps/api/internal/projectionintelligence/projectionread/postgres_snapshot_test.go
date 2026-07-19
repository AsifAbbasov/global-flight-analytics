package projectionread

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/jackc/pgx/v5"
)

type snapshotTransactionStub struct {
	client     postgresClient
	repository trajectoryRepository

	commitErr     error
	rollbackErr   error
	commitCalls   int
	rollbackCalls int
}

func (
	transaction *snapshotTransactionStub,
) Client() postgresClient {
	return transaction.client
}

func (
	transaction *snapshotTransactionStub,
) TrajectoryRepository() trajectoryRepository {
	return transaction.repository
}

func (
	transaction *snapshotTransactionStub,
) Commit(
	context.Context,
) error {
	transaction.commitCalls++
	return transaction.commitErr
}

func (
	transaction *snapshotTransactionStub,
) Rollback(
	context.Context,
) error {
	transaction.rollbackCalls++
	return transaction.rollbackErr
}

type snapshotTransactionStarterStub struct {
	transaction snapshotTransaction
	err         error
	options     pgx.TxOptions
	calls       int
}

func (
	starter *snapshotTransactionStarterStub,
) Begin(
	_ context.Context,
	options pgx.TxOptions,
) (snapshotTransaction, error) {
	starter.calls++
	starter.options = options
	return starter.transaction, starter.err
}

func TestRepeatableReadSnapshotExecutorCommitsOneReadOnlySnapshot(
	t *testing.T,
) {
	client := &scriptedClient{}
	repository := &trajectoryRepositoryStub{
		items: map[string]trajectory.FlightTrajectory{},
	}
	transaction := &snapshotTransactionStub{
		client:     client,
		repository: repository,
	}
	starter := &snapshotTransactionStarterStub{
		transaction: transaction,
	}
	executor := repeatableReadSnapshotExecutor{
		starter: starter,
	}
	operationCalls := 0

	_, err := executor.Execute(
		context.Background(),
		func(
			actualClient postgresClient,
			actualRepository trajectoryRepository,
		) (Snapshot, error) {
			operationCalls++
			if actualClient != client {
				t.Fatal("snapshot operation received a different PostgreSQL client")
			}
			if actualRepository != repository {
				t.Fatal("snapshot operation received a different trajectory repository")
			}
			return Snapshot{}, nil
		},
	)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if starter.calls != 1 || operationCalls != 1 ||
		transaction.commitCalls != 1 ||
		transaction.rollbackCalls != 0 {
		t.Fatalf(
			"unexpected transaction lifecycle: starter=%d operation=%d commit=%d rollback=%d",
			starter.calls,
			operationCalls,
			transaction.commitCalls,
			transaction.rollbackCalls,
		)
	}
	if starter.options.IsoLevel != pgx.RepeatableRead ||
		starter.options.AccessMode != pgx.ReadOnly {
		t.Fatalf(
			"unexpected transaction options: %#v",
			starter.options,
		)
	}
}

func TestRepeatableReadSnapshotExecutorRollsBackFailedRead(
	t *testing.T,
) {
	operationErr := errors.New("snapshot read failed")
	transaction := &snapshotTransactionStub{
		client:     &scriptedClient{},
		repository: &trajectoryRepositoryStub{},
	}
	starter := &snapshotTransactionStarterStub{
		transaction: transaction,
	}
	executor := repeatableReadSnapshotExecutor{
		starter: starter,
	}

	_, err := executor.Execute(
		context.Background(),
		func(
			postgresClient,
			trajectoryRepository,
		) (Snapshot, error) {
			return Snapshot{}, operationErr
		},
	)
	if !errors.Is(err, operationErr) {
		t.Fatalf("error = %v, want operation error", err)
	}
	if transaction.commitCalls != 0 ||
		transaction.rollbackCalls != 1 {
		t.Fatalf(
			"unexpected failed transaction lifecycle: commit=%d rollback=%d",
			transaction.commitCalls,
			transaction.rollbackCalls,
		)
	}
}

func TestRepeatableReadSnapshotExecutorRollsBackCommitFailure(
	t *testing.T,
) {
	commitErr := errors.New("commit failed")
	transaction := &snapshotTransactionStub{
		client:     &scriptedClient{},
		repository: &trajectoryRepositoryStub{},
		commitErr:  commitErr,
	}
	starter := &snapshotTransactionStarterStub{
		transaction: transaction,
	}
	executor := repeatableReadSnapshotExecutor{
		starter: starter,
	}

	_, err := executor.Execute(
		context.Background(),
		func(
			postgresClient,
			trajectoryRepository,
		) (Snapshot, error) {
			return Snapshot{}, nil
		},
	)
	if !errors.Is(err, commitErr) {
		t.Fatalf("error = %v, want commit error", err)
	}
	if transaction.commitCalls != 1 ||
		transaction.rollbackCalls != 1 {
		t.Fatalf(
			"unexpected commit failure lifecycle: commit=%d rollback=%d",
			transaction.commitCalls,
			transaction.rollbackCalls,
		)
	}
}

func TestProjectionSnapshotTransactionOptionsAreStable(
	t *testing.T,
) {
	options := projectionSnapshotTransactionOptions()
	if options.IsoLevel != pgx.RepeatableRead {
		t.Fatalf(
			"isolation = %q, want repeatable read",
			options.IsoLevel,
		)
	}
	if options.AccessMode != pgx.ReadOnly {
		t.Fatalf(
			"access mode = %q, want read only",
			options.AccessMode,
		)
	}
}
