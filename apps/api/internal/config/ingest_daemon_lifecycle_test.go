package config

import (
	"testing"
	"time"
)

func TestLoadIngestDaemonConfigUsesLifecycleDefaults(
	t *testing.T,
) {
	t.Setenv(trafficIngestionIntervalEnvironmentVariable, "")
	t.Setenv(trafficIngestionTerminalTimeoutEnvironmentVariable, "")
	t.Setenv(trafficIngestionStaleRunAfterEnvironmentVariable, "")

	config, err := LoadIngestDaemonConfig()
	if err != nil {
		t.Fatalf("load ingest daemon config: %v", err)
	}

	if config.TerminalTimeout != defaultTrafficIngestionTerminalTimeout {
		t.Fatalf(
			"terminal timeout = %s, want %s",
			config.TerminalTimeout,
			defaultTrafficIngestionTerminalTimeout,
		)
	}
	if config.StaleRunAfter != defaultTrafficIngestionStaleRunAfter {
		t.Fatalf(
			"stale run threshold = %s, want %s",
			config.StaleRunAfter,
			defaultTrafficIngestionStaleRunAfter,
		)
	}
}

func TestLoadIngestDaemonConfigUsesConfiguredLifecycleDurations(
	t *testing.T,
) {
	t.Setenv(trafficIngestionIntervalEnvironmentVariable, "20s")
	t.Setenv(trafficIngestionTerminalTimeoutEnvironmentVariable, "12s")
	t.Setenv(trafficIngestionStaleRunAfterEnvironmentVariable, "45m")

	config, err := LoadIngestDaemonConfig()
	if err != nil {
		t.Fatalf("load ingest daemon config: %v", err)
	}

	if config.Interval != 20*time.Second {
		t.Fatalf("interval = %s, want 20s", config.Interval)
	}
	if config.TerminalTimeout != 12*time.Second {
		t.Fatalf(
			"terminal timeout = %s, want 12s",
			config.TerminalTimeout,
		)
	}
	if config.StaleRunAfter != 45*time.Minute {
		t.Fatalf(
			"stale run threshold = %s, want 45m",
			config.StaleRunAfter,
		)
	}
}

func TestLoadIngestDaemonConfigRejectsInvalidLifecycleDurations(
	t *testing.T,
) {
	testCases := []struct {
		name  string
		value string
	}{
		{name: trafficIngestionTerminalTimeoutEnvironmentVariable, value: "0s"},
		{name: trafficIngestionTerminalTimeoutEnvironmentVariable, value: "invalid"},
		{name: trafficIngestionStaleRunAfterEnvironmentVariable, value: "-1m"},
		{name: trafficIngestionStaleRunAfterEnvironmentVariable, value: "invalid"},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name+"="+testCase.value,
			func(t *testing.T) {
				t.Setenv(trafficIngestionIntervalEnvironmentVariable, "")
				t.Setenv(trafficIngestionTerminalTimeoutEnvironmentVariable, "")
				t.Setenv(trafficIngestionStaleRunAfterEnvironmentVariable, "")
				t.Setenv(testCase.name, testCase.value)

				if _, err := LoadIngestDaemonConfig(); err == nil {
					t.Fatal("expected invalid lifecycle duration error")
				}
			},
		)
	}
}
