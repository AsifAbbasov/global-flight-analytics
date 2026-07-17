package config

import (
	"errors"
	"testing"
	"time"
)

func TestLoadTrafficProviderConfigDefaultsToAirplanesLive(
	t *testing.T,
) {
	t.Setenv(trafficProviderEnvironmentVariable, "")
	t.Setenv(openSkyClientIDEnvironmentVariable, "")
	t.Setenv(openSkyClientSecretEnvironmentVariable, "")
	t.Setenv(openSkyTimeoutEnvironmentVariable, "")
	t.Setenv(openSkyPollingIntervalEnvironmentVariable, "")

	config, err := LoadTrafficProviderConfig()
	if err != nil {
		t.Fatalf("load traffic provider config: %v", err)
	}
	if config.Provider != TrafficProviderAirplanesLive {
		t.Fatalf("provider = %q, want %q", config.Provider, TrafficProviderAirplanesLive)
	}
	if config.OpenSkyPollingInterval != 10*time.Second {
		t.Fatalf("polling interval = %s, want 10s", config.OpenSkyPollingInterval)
	}
}

func TestLoadTrafficProviderConfigAcceptsAutomaticFallback(
	t *testing.T,
) {
	t.Setenv(trafficProviderEnvironmentVariable, string(TrafficProviderAuto))
	t.Setenv(openSkyClientIDEnvironmentVariable, "")
	t.Setenv(openSkyClientSecretEnvironmentVariable, "")
	t.Setenv(openSkyPollingIntervalEnvironmentVariable, "")

	config, err := LoadTrafficProviderConfig()
	if err != nil {
		t.Fatalf("load automatic traffic provider config: %v", err)
	}
	if config.Provider != TrafficProviderAuto {
		t.Fatalf("provider = %q, want %q", config.Provider, TrafficProviderAuto)
	}
	if config.OpenSkyPollingInterval != 10*time.Second {
		t.Fatalf("polling interval = %s, want 10s", config.OpenSkyPollingInterval)
	}
}

func TestLoadTrafficProviderConfigUsesAuthenticatedMinimum(
	t *testing.T,
) {
	t.Setenv(trafficProviderEnvironmentVariable, string(TrafficProviderOpenSky))
	t.Setenv(openSkyClientIDEnvironmentVariable, "client")
	t.Setenv(openSkyClientSecretEnvironmentVariable, "secret")
	t.Setenv(openSkyPollingIntervalEnvironmentVariable, "")

	config, err := LoadTrafficProviderConfig()
	if err != nil {
		t.Fatalf("load authenticated traffic provider config: %v", err)
	}
	if config.OpenSkyPollingInterval != 5*time.Second {
		t.Fatalf("polling interval = %s, want 5s", config.OpenSkyPollingInterval)
	}
}

func TestLoadTrafficProviderConfigRejectsCredentialHalfPair(
	t *testing.T,
) {
	t.Setenv(openSkyClientIDEnvironmentVariable, "client")
	t.Setenv(openSkyClientSecretEnvironmentVariable, "")

	_, err := LoadTrafficProviderConfig()
	if !errors.Is(err, ErrOpenSkyCredentialPairRequired) {
		t.Fatalf("expected credential pair error, got %v", err)
	}
}

func TestLoadTrafficProviderConfigRejectsAnonymousFiveSecondPolling(
	t *testing.T,
) {
	t.Setenv(trafficProviderEnvironmentVariable, string(TrafficProviderOpenSky))
	t.Setenv(openSkyClientIDEnvironmentVariable, "")
	t.Setenv(openSkyClientSecretEnvironmentVariable, "")
	t.Setenv(openSkyPollingIntervalEnvironmentVariable, "5s")

	_, err := LoadTrafficProviderConfig()
	if err == nil {
		t.Fatal("expected anonymous polling interval validation error")
	}
}
