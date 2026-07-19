package historicalaggregate

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConfig struct {
	Pool *pgxpool.Pool
	Now  func() time.Time
}

type Executor interface {
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
