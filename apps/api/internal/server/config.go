package server

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DatabasePool     *pgxpool.Pool
	Logger           *slog.Logger
	OpenMeteoTimeout time.Duration
}
