// BACKEND_STARTUP_CONTEXT_HARDENING_V1
package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	errPostgresContextRequired = errors.New(
		"postgres context is required",
	)
	errPostgresConnectTimeoutInvalid = errors.New(
		"postgres connect timeout must be greater than zero",
	)
)

func NewPostgresPool(
	databaseURL string,
	connectTimeout time.Duration,
) (*pgxpool.Pool, error) {
	return NewPostgresPoolContext(
		context.Background(),
		databaseURL,
		connectTimeout,
	)
}

func NewPostgresPoolContext(
	ctx context.Context,
	databaseURL string,
	connectTimeout time.Duration,
) (*pgxpool.Pool, error) {
	if ctx == nil {
		return nil, errPostgresContextRequired
	}
	if connectTimeout <= 0 {
		return nil, errPostgresConnectTimeoutInvalid
	}

	connectCtx, cancel := context.WithTimeout(
		ctx,
		connectTimeout,
	)
	defer cancel()

	pool, err := pgxpool.New(
		connectCtx,
		databaseURL,
	)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(
		connectCtx,
	); err != nil {
		pool.Close()

		return nil, err
	}

	return pool, nil
}
