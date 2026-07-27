package featurestore

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const PostgresVersion = "flight-feature-postgres-store-v2"

type PostgresConfig struct {
	Pool *pgxpool.Pool
	Now  func() time.Time
}
