package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	trafficIngestionIntervalEnvironmentVariable = "TRAFFIC_INGESTION_INTERVAL"

	defaultTrafficIngestionInterval = 10 * time.Second
)

type IngestDaemonConfig struct {
	Interval time.Duration
}

func LoadIngestDaemonConfig() (
	IngestDaemonConfig,
	error,
) {
	rawValue := strings.TrimSpace(
		os.Getenv(
			trafficIngestionIntervalEnvironmentVariable,
		),
	)

	if rawValue == "" {
		return IngestDaemonConfig{
			Interval: defaultTrafficIngestionInterval,
		}, nil
	}

	interval, err := time.ParseDuration(
		rawValue,
	)
	if err != nil {
		return IngestDaemonConfig{}, fmt.Errorf(
			"load traffic ingestion interval: %s must be a duration: %w",
			trafficIngestionIntervalEnvironmentVariable,
			err,
		)
	}

	if interval <= 0 {
		return IngestDaemonConfig{}, fmt.Errorf(
			"load traffic ingestion interval: %s must be greater than zero",
			trafficIngestionIntervalEnvironmentVariable,
		)
	}

	return IngestDaemonConfig{
		Interval: interval,
	}, nil
}
