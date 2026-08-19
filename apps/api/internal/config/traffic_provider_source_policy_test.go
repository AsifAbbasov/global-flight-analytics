package config

import (
	"errors"
	"testing"
)

func TestLoadTrafficProviderConfigRejectsOpenSkyWithoutOperationalAgreement(t *testing.T) {
	t.Setenv(trafficProviderEnvironmentVariable, string(TrafficProviderOpenSky))
	t.Setenv(openSkyOperationalAgreementConfirmedEnvironmentVariable, "false")
	t.Setenv(openSkyClientIDEnvironmentVariable, "")
	t.Setenv(openSkyClientSecretEnvironmentVariable, "")

	_, err := LoadTrafficProviderConfig()
	if !errors.Is(err, ErrOpenSkyOperationalAgreementRequired) {
		t.Fatalf("expected operational agreement error, got %v", err)
	}
}

func TestLoadTrafficProviderConfigRejectsAutoWithoutOperationalAgreement(t *testing.T) {
	t.Setenv(trafficProviderEnvironmentVariable, string(TrafficProviderAuto))
	t.Setenv(openSkyOperationalAgreementConfirmedEnvironmentVariable, "false")
	t.Setenv(openSkyClientIDEnvironmentVariable, "")
	t.Setenv(openSkyClientSecretEnvironmentVariable, "")

	_, err := LoadTrafficProviderConfig()
	if !errors.Is(err, ErrOpenSkyOperationalAgreementRequired) {
		t.Fatalf("expected operational agreement error, got %v", err)
	}
}

func TestLoadTrafficProviderConfigAcceptsConfirmedOpenSkyAgreement(t *testing.T) {
	t.Setenv(trafficProviderEnvironmentVariable, string(TrafficProviderOpenSky))
	t.Setenv(openSkyOperationalAgreementConfirmedEnvironmentVariable, "true")
	t.Setenv(openSkyClientIDEnvironmentVariable, "client")
	t.Setenv(openSkyClientSecretEnvironmentVariable, "secret")

	config, err := LoadTrafficProviderConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !config.OpenSkyOperationalAgreementConfirmed {
		t.Fatal("agreement confirmation was not preserved")
	}
}

func TestLoadTrafficProviderConfigRejectsMalformedAgreementFlag(t *testing.T) {
	t.Setenv(trafficProviderEnvironmentVariable, string(TrafficProviderAirplanesLive))
	t.Setenv(openSkyOperationalAgreementConfirmedEnvironmentVariable, "yes")

	_, err := LoadTrafficProviderConfig()
	if !errors.Is(err, ErrOpenSkyOperationalAgreementInvalid) {
		t.Fatalf("expected malformed agreement flag error, got %v", err)
	}
}
