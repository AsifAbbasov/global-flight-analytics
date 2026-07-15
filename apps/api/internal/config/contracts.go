package config

import "time"

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
