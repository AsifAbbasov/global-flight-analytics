package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(
	databaseURL string,
	connectTimeout time.Duration,
) (*pgxpool.Pool, error) {
	if connectTimeout <= 0 {
		return nil, fmt.Errorf(
			"postgres connect timeout must be greater than zero",
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		connectTimeout,
	)
	defer cancel()

	pool, err := pgxpool.New(
		ctx,
		databaseURL,
	)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(
		ctx,
	); err != nil {
		pool.Close()

		return nil, err
	}

	return pool, nil
}
