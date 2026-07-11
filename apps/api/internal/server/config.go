package server

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProtectionConfig struct {
	AllowedOrigins string

	BodyLimitBytes int

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	RateLimitMax    int
	RateLimitWindow time.Duration
}

type Config struct {
	DatabasePool     *pgxpool.Pool
	Logger           *slog.Logger
	OpenMeteoTimeout time.Duration
	Protection       ProtectionConfig
}
