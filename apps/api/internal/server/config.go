package server

import (
	"log/slog"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
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

	ClientIPHeader     string
	TrustedProxyRanges []string

	MutationKeyDigest     internalapikey.Digest
	MutationKeyConfigured bool
}

type Config struct {
	DatabasePool     *pgxpool.Pool
	Logger           *slog.Logger
	OpenMeteoTimeout time.Duration
	Protection       ProtectionConfig
}

// STAGE-14-5-MUTATION-ENDPOINT-PROTECTION
