package server

import (
	"log/slog"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/observability"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
	livetraffic "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/live"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProtectionConfig struct {
	AllowedOrigins string

	BodyLimitBytes int

	ReadTimeout    time.Duration
	RequestTimeout time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration

	RateLimitMax    int
	RateLimitWindow time.Duration

	ClientIPHeader     string
	TrustedProxyRanges []string

	MutationKeyDigest     internalapikey.Digest
	MutationKeyConfigured bool
}

type Config struct {
	DatabasePool          *pgxpool.Pool
	Logger                *slog.Logger
	OpenMeteoTimeout      time.Duration
	ObservabilityRegistry *observability.Registry
	LiveTrafficStore      *livetraffic.Store
	MetricsKeyDigest      internalapikey.Digest
	MetricsKeyConfigured  bool
	Protection            ProtectionConfig
}

// STAGE-14-5-MUTATION-ENDPOINT-PROTECTION
