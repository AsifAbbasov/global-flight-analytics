package config

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
)

type PostgresConfig struct {
	URL            string
	ConnectTimeout time.Duration
}

type APIProtectionConfig struct {
	AllowedOrigins string

	BodyLimitBytes int

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	RateLimitMax    int
	RateLimitWindow time.Duration

	MutationKeyDigest     internalapikey.Digest
	MutationKeyConfigured bool
}

type ServerConfig struct {
	Port             string
	Database         *PostgresConfig
	OpenMeteoTimeout time.Duration
	APIProtection    APIProtectionConfig
}

type IngestConfig struct {
	Database PostgresConfig

	TrafficIngestionLatitude  float64
	TrafficIngestionLongitude float64
	TrafficIngestionRadius    int

	AirplanesLiveTimeout time.Duration

	TrajectoryMaxTimeGap        time.Duration
	TrajectoryMaxGroundSpeedMPS float64
}

type ImportAirportsConfig struct {
	Database PostgresConfig

	OurAirportsTimeout      time.Duration
	OurAirportsCountryCodes []string
}

type MigrationConfig struct {
	Database PostgresConfig

	MigrationsDir    string
	MigrationTimeout time.Duration
}

type HistoricalMaterializationConfig struct {
	Database PostgresConfig

	OperationTimeout time.Duration
}

type VerifyAirportsConfig struct {
	DatabaseURL string
}

// STAGE-14-5-MUTATION-ENDPOINT-PROTECTION
