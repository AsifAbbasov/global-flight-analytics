package config

import (
	"errors"
	"testing"
	"time"
)

func TestLoadLiveTrafficCollectorConfigDisabledByDefault(t *testing.T) {
	t.Setenv(liveTrafficCollectorEnabledEnvironmentVariable, "")
	cfg, err := LoadLiveTrafficCollectorConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled {
		t.Fatal("collector must be disabled by default")
	}
	if cfg.Provider != LiveTrafficProviderADSBLOL {
		t.Fatalf("provider = %q", cfg.Provider)
	}
}

func TestLoadLiveTrafficCollectorConfigRequiresProductionContact(t *testing.T) {
	t.Setenv(liveTrafficCollectorEnabledEnvironmentVariable, "true")
	t.Setenv(adsbLOLProductionContactConfirmedEnvironmentVariable, "false")

	_, err := LoadLiveTrafficCollectorConfig()
	if !errors.Is(err, ErrADSBLOLProductionContactRequired) {
		t.Fatalf("expected contact requirement, got %v", err)
	}
}

func TestLoadLiveTrafficCollectorConfigEnabled(t *testing.T) {
	t.Setenv(liveTrafficCollectorEnabledEnvironmentVariable, "true")
	t.Setenv(adsbLOLProductionContactConfirmedEnvironmentVariable, "true")
	t.Setenv(liveTrafficLatitudeEnvironmentVariable, "40.4093")
	t.Setenv(liveTrafficLongitudeEnvironmentVariable, "49.8671")
	t.Setenv(liveTrafficRadiusEnvironmentVariable, "250")

	cfg, err := LoadLiveTrafficCollectorConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled {
		t.Fatal("expected enabled collector")
	}
	if cfg.PollInterval != 10*time.Second {
		t.Fatalf("poll interval = %s", cfg.PollInterval)
	}
	if cfg.RequestTimeout != 8*time.Second {
		t.Fatalf("request timeout = %s", cfg.RequestTimeout)
	}
	if cfg.RadiusNM != 250 {
		t.Fatalf("radius = %d", cfg.RadiusNM)
	}
	if !cfg.ADSBLOLProductionContactConfirmed {
		t.Fatal("expected confirmed contact flag")
	}
}

func TestLoadLiveTrafficCollectorConfigRejectsTooFastPolling(t *testing.T) {
	t.Setenv(liveTrafficCollectorEnabledEnvironmentVariable, "true")
	t.Setenv(adsbLOLProductionContactConfirmedEnvironmentVariable, "true")
	t.Setenv(liveTrafficLatitudeEnvironmentVariable, "40.4093")
	t.Setenv(liveTrafficLongitudeEnvironmentVariable, "49.8671")
	t.Setenv(liveTrafficRadiusEnvironmentVariable, "250")
	t.Setenv(liveTrafficPollIntervalEnvironmentVariable, "5s")

	if _, err := LoadLiveTrafficCollectorConfig(); err == nil {
		t.Fatal("expected fast-polling rejection")
	}
}

func TestLoadLiveTrafficCollectorConfigRejectsUnknownProvider(t *testing.T) {
	t.Setenv(liveTrafficCollectorEnabledEnvironmentVariable, "true")
	t.Setenv(liveTrafficProviderEnvironmentVariable, "unknown")
	t.Setenv(adsbLOLProductionContactConfirmedEnvironmentVariable, "true")

	_, err := LoadLiveTrafficCollectorConfig()
	if !errors.Is(err, ErrLiveTrafficProviderInvalid) {
		t.Fatalf("expected provider error, got %v", err)
	}
}
