package projectionread

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type rowScanner interface {
	Scan(...any) error
}

type rowIterator interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

type postgresClient interface {
	QueryRow(
		context.Context,
		string,
		...any,
	) rowScanner

	Query(
		context.Context,
		string,
		...any,
	) (rowIterator, error)
}

type pgxPoolClient struct {
	pool *pgxpool.Pool
}

func (
	client pgxPoolClient,
) QueryRow(
	ctx context.Context,
	query string,
	args ...any,
) rowScanner {
	return client.pool.QueryRow(
		ctx,
		query,
		args...,
	)
}

func (
	client pgxPoolClient,
) Query(
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
		return nil,
			err
	}

	return rows,
		nil
}
