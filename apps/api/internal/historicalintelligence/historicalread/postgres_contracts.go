package historicalread

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Pool *pgxpool.Pool
}

// Executor is the shared query contract implemented by pgx pools and
// transactions. It allows a historical read to observe deterministic,
// uncommitted verification evidence inside one rollback-only transaction.
type Executor interface {
	Query(
		context.Context,
		string,
		...any,
	) (pgx.Rows, error)
}

type rowIterator interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

type postgresClient interface {
	Query(
		context.Context,
		string,
		...any,
	) (rowIterator, error)
}

type executorClient struct {
	executor Executor
}

func (client executorClient) Query(
	ctx context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	rows, err := client.executor.Query(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

type pgxPoolClient struct {
	pool *pgxpool.Pool
}

func (client pgxPoolClient) Query(
	ctx context.Context,
	query string,
	args ...any,
) (rowIterator, error) {
	rows, err := client.pool.Query(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
