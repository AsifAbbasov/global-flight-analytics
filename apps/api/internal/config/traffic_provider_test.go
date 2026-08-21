package config

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/adsblol"
)

func resetTrafficProviderEnvironment(t *testing.T) {
	t.Helper()

	for _, name := range []string{
		trafficProviderEnvironmentVariable,
		adsbLOLBaseURLEnvironmentVariable,
		adsbLOLTimeoutEnvironmentVariable,
		airplanesLiveTimeoutEnvironmentVariable,
		airplanesLiveAccessApprovedEnvironmentVariable,
		openSkyBaseURLEnvironmentVariable,
		openSkyTokenURLEnvironmentVariable,
		openSkyClientIDEnvironmentVariable,
		openSkyClientSecretEnvironmentVariable,
		openSkyTimeoutEnvironmentVariable,
		openSkyPollingIntervalEnvironmentVariable,
		openSkyOperationalAgreementConfirmedEnvironmentVariable,
	} {
		t.Setenv(name, "")
	}
}

func TestLoadTrafficProviderConfigDefaultsToADSBLOL(
	t *testing.T,
) {
	resetTrafficProviderEnvironment(t)

	config, err := LoadTrafficProviderConfig()
	if err != nil {
		t.Fatalf("load traffic provider config: %v", err)
	}
	if config.Provider != TrafficProviderADSBLOL {
		t.Fatalf("provider = %q, want %q", config.Provider, TrafficProviderADSBLOL)
	}
	if config.ADSBLOL.BaseURL != adsblol.BaseURL {
		t.Fatalf("ADSB.lol base URL = %q, want %q", config.ADSBLOL.BaseURL, adsblol.BaseURL)
	}
	if config.ADSBLOL.Timeout != 15*time.Second {
		t.Fatalf("ADSB.lol timeout = %s, want 15s", config.ADSBLOL.Timeout)
	}
	if config.AirplanesLive.Timeout != 15*time.Second {
		t.Fatalf("airplanes.live timeout = %s, want 15s", config.AirplanesLive.Timeout)
	}
	if config.AirplanesLive.AccessApproved {
		t.Fatal("airplanes.live access must default fail-closed")
	}
	if config.OpenSky.OperationalAgreementConfirmed {
		t.Fatal("OpenSky operational agreement must default fail-closed")
	}
}

func TestTrafficProviderConfigAutomaticCandidatesFailClosed(
	t *testing.T,
) {
	config := TrafficProviderConfig{}

	got := config.AutomaticCandidates()
	want := []TrafficProvider{TrafficProviderADSBLOL}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("automatic candidates = %+v, want %+v", got, want)
	}
}

func TestTrafficProviderConfigAutomaticCandidatesPreferApprovedAirplanesLive(
	t *testing.T,
) {
	config := TrafficProviderConfig{
		AirplanesLive: AirplanesLiveProviderConfig{
			AccessApproved: true,
		},
		OpenSky: OpenSkyProviderConfig{
			OperationalAgreementConfirmed: true,
		},
	}

	got := config.AutomaticCandidates()
	want := []TrafficProvider{
		TrafficProviderADSBLOL,
		TrafficProviderAirplanesLive,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("automatic candidates = %+v, want %+v", got, want)
	}
}

func TestTrafficProviderConfigAutomaticCandidatesUsesApprovedOpenSky(
	t *testing.T,
) {
	config := TrafficProviderConfig{
		OpenSky: OpenSkyProviderConfig{
			OperationalAgreementConfirmed: true,
		},
	}

	got := config.AutomaticCandidates()
	want := []TrafficProvider{
		TrafficProviderADSBLOL,
		TrafficProviderOpenSky,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("automatic candidates = %+v, want %+v", got, want)
	}
}

func TestTrafficProviderConfigRequireEligibleRejectsRestrictedProviders(
	t *testing.T,
) {
	config := TrafficProviderConfig{}

	if !errors.Is(
		config.RequireEligible(TrafficProviderAirplanesLive),
		ErrAirplanesLiveAccessApprovalRequired,
	) {
		t.Fatal("expected airplanes.live access approval error")
	}
	if !errors.Is(
		config.RequireEligible(TrafficProviderOpenSky),
		ErrOpenSkyOperationalAgreementRequired,
	) {
		t.Fatal("expected OpenSky operational agreement error")
	}
	if err := config.RequireEligible(TrafficProviderADSBLOL); err != nil {
		t.Fatalf("ADSB.lol should be eligible: %v", err)
	}
}

func TestLoadTrafficProviderConfigParsesAccessApprovals(
	t *testing.T,
) {
	resetTrafficProviderEnvironment(t)
	t.Setenv(airplanesLiveAccessApprovedEnvironmentVariable, "true")
	t.Setenv(openSkyOperationalAgreementConfirmedEnvironmentVariable, "true")

	config, err := LoadTrafficProviderConfig()
	if err != nil {
		t.Fatalf("load access-approved provider config: %v", err)
	}
	if !config.AirplanesLive.AccessApproved {
		t.Fatal("expected airplanes.live access approval")
	}
	if !config.OpenSky.OperationalAgreementConfirmed {
		t.Fatal("expected OpenSky operational agreement confirmation")
	}
}

func TestLoadTrafficProviderConfigRejectsInvalidApprovalBoolean(
	t *testing.T,
) {
	resetTrafficProviderEnvironment(t)
	t.Setenv(openSkyOperationalAgreementConfirmedEnvironmentVariable, "not-a-bool")

	_, err := LoadTrafficProviderConfig()
	if err == nil {
		t.Fatal("expected invalid approval boolean error")
	}
}

func TestLoadTrafficProviderConfigUsesAuthenticatedMinimum(
	t *testing.T,
) {
	resetTrafficProviderEnvironment(t)
	t.Setenv(trafficProviderEnvironmentVariable, string(TrafficProviderOpenSky))
	t.Setenv(openSkyClientIDEnvironmentVariable, "client")
	t.Setenv(openSkyClientSecretEnvironmentVariable, "secret")

	config, err := LoadTrafficProviderConfig()
	if err != nil {
		t.Fatalf("load authenticated traffic provider config: %v", err)
	}
	if config.OpenSky.PollingInterval != 5*time.Second {
		t.Fatalf("polling interval = %s, want 5s", config.OpenSky.PollingInterval)
	}
}

func TestLoadTrafficProviderConfigRejectsCredentialHalfPair(
	t *testing.T,
) {
	resetTrafficProviderEnvironment(t)
	t.Setenv(openSkyClientIDEnvironmentVariable, "client")

	_, err := LoadTrafficProviderConfig()
	if !errors.Is(err, ErrOpenSkyCredentialPairRequired) {
		t.Fatalf("expected credential pair error, got %v", err)
	}
}

func TestLoadTrafficProviderConfigRejectsAnonymousFiveSecondPolling(
	t *testing.T,
) {
	resetTrafficProviderEnvironment(t)
	t.Setenv(trafficProviderEnvironmentVariable, string(TrafficProviderOpenSky))
	t.Setenv(openSkyPollingIntervalEnvironmentVariable, "5s")

	_, err := LoadTrafficProviderConfig()
	if err == nil {
		t.Fatal("expected anonymous polling interval validation error")
	}
}
