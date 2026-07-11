package config

import (
	"testing"
	"time"
)

func TestLoadIngestDaemonConfigUsesDefaultInterval(
	t *testing.T,
) {
	t.Setenv(
		trafficIngestionIntervalEnvironmentVariable,
		"",
	)

	config, err := LoadIngestDaemonConfig()
	if err != nil {
		t.Fatalf(
			"load ingest daemon config: %v",
			err,
		)
	}

	if config.Interval != defaultTrafficIngestionInterval {
		t.Fatalf(
			"expected default interval %s, got %s",
			defaultTrafficIngestionInterval,
			config.Interval,
		)
	}
}

func TestLoadIngestDaemonConfigUsesConfiguredInterval(
	t *testing.T,
) {
	t.Setenv(
		trafficIngestionIntervalEnvironmentVariable,
		"25s",
	)

	config, err := LoadIngestDaemonConfig()
	if err != nil {
		t.Fatalf(
			"load ingest daemon config: %v",
			err,
		)
	}

	if config.Interval != 25*time.Second {
		t.Fatalf(
			"expected interval 25s, got %s",
			config.Interval,
		)
	}
}

func TestLoadIngestDaemonConfigRejectsInvalidDuration(
	t *testing.T,
) {
	t.Setenv(
		trafficIngestionIntervalEnvironmentVariable,
		"invalid",
	)

	_, err := LoadIngestDaemonConfig()
	if err == nil {
		t.Fatal(
			"expected invalid duration error",
		)
	}
}

func TestLoadIngestDaemonConfigRejectsNonPositiveDuration(
	t *testing.T,
) {
	for _, rawValue := range []string{
		"0s",
		"-1s",
	} {
		t.Run(
			rawValue,
			func(
				t *testing.T,
			) {
				t.Setenv(
					trafficIngestionIntervalEnvironmentVariable,
					rawValue,
				)

				_, err := LoadIngestDaemonConfig()
				if err == nil {
					t.Fatal(
						"expected non-positive duration error",
					)
				}
			},
		)
	}
}
