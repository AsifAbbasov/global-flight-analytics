package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadServerConfigDefaultsRequestTimeoutToWriteTimeout(
	t *testing.T,
) {
	setServerWithoutDatabaseEnvironment(t)
	clearAPIProtectionEnvironment(t)

	loadedConfig, err := LoadServerConfig()
	if err != nil {
		t.Fatalf("load server configuration: %v", err)
	}
	if loadedConfig.APIProtection.RequestTimeout !=
		loadedConfig.APIProtection.WriteTimeout {
		t.Fatalf(
			"expected request timeout to default to write timeout %s, got %s",
			loadedConfig.APIProtection.WriteTimeout,
			loadedConfig.APIProtection.RequestTimeout,
		)
	}
}

func TestLoadServerConfigUsesConfiguredRequestTimeout(
	t *testing.T,
) {
	setServerWithoutDatabaseEnvironment(t)
	clearAPIProtectionEnvironment(t)
	t.Setenv(apiRequestTimeoutEnvironmentVariable, "7s")

	loadedConfig, err := LoadServerConfig()
	if err != nil {
		t.Fatalf("load server configuration: %v", err)
	}
	if loadedConfig.APIProtection.RequestTimeout != 7*time.Second {
		t.Fatalf(
			"expected configured request timeout 7s, got %s",
			loadedConfig.APIProtection.RequestTimeout,
		)
	}
}

func TestLoadServerConfigRejectsRequestTimeoutAboveWriteTimeout(
	t *testing.T,
) {
	setServerWithoutDatabaseEnvironment(t)
	clearAPIProtectionEnvironment(t)
	t.Setenv(apiRequestTimeoutEnvironmentVariable, "20s")
	t.Setenv(apiWriteTimeoutEnvironmentVariable, "15s")

	loadedConfig, err := LoadServerConfig()
	if err == nil {
		t.Fatal("expected request timeout relationship error")
	}
	if loadedConfig.Port != "" ||
		loadedConfig.Database != nil ||
		loadedConfig.OpenMeteoTimeout != 0 ||
		loadedConfig.APIProtection.RequestTimeout != 0 {
		t.Fatalf("expected zero server configuration, got %+v", loadedConfig)
	}
	if !strings.Contains(
		err.Error(),
		"API_REQUEST_TIMEOUT must not exceed API_WRITE_TIMEOUT",
	) {
		t.Fatalf("unexpected request timeout error: %v", err)
	}
}
