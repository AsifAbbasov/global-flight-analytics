package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	trafficIngestionIntervalEnvironmentVariable        = "TRAFFIC_INGESTION_INTERVAL"
	trafficIngestionTerminalTimeoutEnvironmentVariable = "TRAFFIC_INGESTION_TERMINAL_TIMEOUT"
	trafficIngestionStaleRunAfterEnvironmentVariable   = "TRAFFIC_INGESTION_STALE_RUN_AFTER"
	trafficIngestionMaxBackoffEnvironmentVariable      = "TRAFFIC_INGESTION_MAX_BACKOFF"

	defaultTrafficIngestionInterval        = 10 * time.Second
	defaultTrafficIngestionTerminalTimeout = 15 * time.Second
	defaultTrafficIngestionStaleRunAfter   = 30 * time.Minute
	defaultTrafficIngestionMaxBackoff      = 2 * time.Minute
)

type IngestDaemonConfig struct {
	Interval        time.Duration
	TerminalTimeout time.Duration
	StaleRunAfter   time.Duration
	MaxBackoff      time.Duration
}

func LoadIngestDaemonConfig() (
	IngestDaemonConfig,
	error,
) {
	interval, err := loadOptionalPositiveIngestDuration(
		trafficIngestionIntervalEnvironmentVariable,
		defaultTrafficIngestionInterval,
	)
	if err != nil {
		return IngestDaemonConfig{}, err
	}

	terminalTimeout, err := loadOptionalPositiveIngestDuration(
		trafficIngestionTerminalTimeoutEnvironmentVariable,
		defaultTrafficIngestionTerminalTimeout,
	)
	if err != nil {
		return IngestDaemonConfig{}, err
	}

	staleRunAfter, err := loadOptionalPositiveIngestDuration(
		trafficIngestionStaleRunAfterEnvironmentVariable,
		defaultTrafficIngestionStaleRunAfter,
	)
	if err != nil {
		return IngestDaemonConfig{}, err
	}

	maxBackoff, err := loadOptionalPositiveIngestDuration(
		trafficIngestionMaxBackoffEnvironmentVariable,
		defaultTrafficIngestionMaxBackoff,
	)
	if err != nil {
		return IngestDaemonConfig{}, err
	}
	if maxBackoff < interval {
		return IngestDaemonConfig{}, fmt.Errorf(
			"load ingest daemon configuration: %s must be at least %s",
			trafficIngestionMaxBackoffEnvironmentVariable,
			trafficIngestionIntervalEnvironmentVariable,
		)
	}

	return IngestDaemonConfig{
		Interval:        interval,
		TerminalTimeout: terminalTimeout,
		StaleRunAfter:   staleRunAfter,
		MaxBackoff:      maxBackoff,
	}, nil
}

func loadOptionalPositiveIngestDuration(
	name string,
	defaultValue time.Duration,
) (time.Duration, error) {
	rawValue := strings.TrimSpace(
		os.Getenv(name),
	)
	if rawValue == "" {
		return defaultValue, nil
	}

	value, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf(
			"load ingest daemon configuration: %s must be a duration: %w",
			name,
			err,
		)
	}

	if value <= 0 {
		return 0, fmt.Errorf(
			"load ingest daemon configuration: %s must be greater than zero",
			name,
		)
	}

	return value, nil
}
