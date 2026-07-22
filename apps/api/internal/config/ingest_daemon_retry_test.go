package config

import (
	"testing"
)

func TestLoadIngestDaemonConfigUsesDefaultMaximumBackoff(
	t *testing.T,
) {
	t.Setenv(
		trafficIngestionIntervalEnvironmentVariable,
		"",
	)
	t.Setenv(
		trafficIngestionMaxBackoffEnvironmentVariable,
		"",
	)

	config, err := LoadIngestDaemonConfig()
	if err != nil {
		t.Fatalf(
			"load ingest daemon configuration: %v",
			err,
		)
	}
	if config.MaxBackoff !=
		defaultTrafficIngestionMaxBackoff {
		t.Fatalf(
			"maximum backoff = %s, want %s",
			config.MaxBackoff,
			defaultTrafficIngestionMaxBackoff,
		)
	}
}

func TestLoadIngestDaemonConfigRejectsMaximumBackoffBelowInterval(
	t *testing.T,
) {
	t.Setenv(
		trafficIngestionIntervalEnvironmentVariable,
		"30s",
	)
	t.Setenv(
		trafficIngestionMaxBackoffEnvironmentVariable,
		"10s",
	)

	_, err := LoadIngestDaemonConfig()
	if err == nil {
		t.Fatal(
			"expected maximum backoff validation error",
		)
	}
}
