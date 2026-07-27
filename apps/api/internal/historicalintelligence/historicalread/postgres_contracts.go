package historicalread

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Pool *pgxpool.Pool
}

type rowIterator interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

type postgresClient interface {
	Query(context.Context, string, ...any) (rowIterator, error)
}

type managedSnapshot interface {
	postgresClient
	Commit(context.Context) error
	Rollback(context.Context) error
}

type snapshotBeginner interface {
	BeginSnapshot(context.Context) (managedSnapshot, error)
}

type poolSnapshotBeginner struct {
	pool *pgxpool.Pool
}

func (beginner poolSnapshotBeginner) BeginSnapshot(
	ctx context.Context,
) (managedSnapshot, error) {
	transaction, err := beginner.pool.BeginTx(
		ctx,
		pgx.TxOptions{
			IsoLevel:   pgx.RepeatableRead,
			AccessMode: pgx.ReadOnly,
		},
	)
	if err != nil {
		return nil, err
	}

	return transactionClient{transaction: transaction}, nil
}

type transactionClient struct {
	transaction pgx.Tx
}

func (client transactionClient) Query(
	ctx context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	rows, err := client.transaction.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (client transactionClient) Commit(ctx context.Context) error {
	return client.transaction.Commit(ctx)
}

func (client transactionClient) Rollback(ctx context.Context) error {
	return client.transaction.Rollback(ctx)
}
