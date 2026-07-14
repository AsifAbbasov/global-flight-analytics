package routestore

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

const PostgresExecutorVersion = "route-postgres-executor-v1"

type PostgresExecutor interface {
	QueryRow(
		ctx context.Context,
		query string,
		args ...any,
	) pgx.Row
	Query(
		ctx context.Context,
		query string,
		args ...any,
	) (pgx.Rows, error)
}

type postgresExecutorClient struct {
	executor PostgresExecutor
}

func (client postgresExecutorClient) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) rowScanner {
	return client.executor.QueryRow(
		ctx,
		query,
		args...,
	)
}

func (client postgresExecutorClient) Query(
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

func NewPostgresWithExecutor(
	executor PostgresExecutor,
	now func() time.Time,
) (*PostgresStore, error) {
	if executor == nil {
		return nil, ErrPostgresExecutorRequired
	}

	return newPostgresStore(
		postgresExecutorClient{
			executor: executor,
		},
		now,
	), nil
}
