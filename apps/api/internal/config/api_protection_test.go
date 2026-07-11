package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadServerConfigUsesDefaultAPIProtection(
	t *testing.T,
) {
	setServerWithoutDatabaseEnvironment(
		t,
	)
	clearAPIProtectionEnvironment(
		t,
	)

	loadedConfig, err := LoadServerConfig()
	if err != nil {
		t.Fatalf(
			"load server configuration: %v",
			err,
		)
	}

	protection := loadedConfig.APIProtection

	if protection.AllowedOrigins != defaultAPIAllowedOrigins {
		t.Fatalf(
			"expected default allowed origins %q, got %q",
			defaultAPIAllowedOrigins,
			protection.AllowedOrigins,
		)
	}

	if protection.BodyLimitBytes != defaultAPIBodyLimitBytes {
		t.Fatalf(
			"expected default body limit %d, got %d",
			defaultAPIBodyLimitBytes,
			protection.BodyLimitBytes,
		)
	}

	if protection.ReadTimeout != defaultAPIReadTimeout ||
		protection.WriteTimeout != defaultAPIWriteTimeout ||
		protection.IdleTimeout != defaultAPIIdleTimeout {
		t.Fatalf(
			"unexpected default transport timeouts: %+v",
			protection,
		)
	}

	if protection.RateLimitMax != defaultAPIRateLimitMax ||
		protection.RateLimitWindow != defaultAPIRateLimitWindow {
		t.Fatalf(
			"unexpected default rate limit: %+v",
			protection,
		)
	}
}

func TestLoadServerConfigUsesConfiguredAPIProtection(
	t *testing.T,
) {
	setServerWithoutDatabaseEnvironment(
		t,
	)

	t.Setenv(
		apiAllowedOriginsEnvironmentVariable,
		" https://web.example.com,https://admin.example.com,https://web.example.com ",
	)
	t.Setenv(
		apiBodyLimitBytesEnvironmentVariable,
		"2048",
	)
	t.Setenv(
		apiReadTimeoutEnvironmentVariable,
		"2s",
	)
	t.Setenv(
		apiWriteTimeoutEnvironmentVariable,
		"3s",
	)
	t.Setenv(
		apiIdleTimeoutEnvironmentVariable,
		"4s",
	)
	t.Setenv(
		apiRateLimitMaxEnvironmentVariable,
		"25",
	)
	t.Setenv(
		apiRateLimitWindowEnvironmentVariable,
		"30s",
	)

	loadedConfig, err := LoadServerConfig()
	if err != nil {
		t.Fatalf(
			"load server configuration: %v",
			err,
		)
	}

	protection := loadedConfig.APIProtection

	if protection.AllowedOrigins != "https://web.example.com,https://admin.example.com" {
		t.Fatalf(
			"unexpected allowed origins: %q",
			protection.AllowedOrigins,
		)
	}

	if protection.BodyLimitBytes != 2048 ||
		protection.ReadTimeout != 2*time.Second ||
		protection.WriteTimeout != 3*time.Second ||
		protection.IdleTimeout != 4*time.Second ||
		protection.RateLimitMax != 25 ||
		protection.RateLimitWindow != 30*time.Second {
		t.Fatalf(
			"unexpected configured protection: %+v",
			protection,
		)
	}
}

func TestLoadServerConfigRejectsInvalidAPIProtection(
	t *testing.T,
) {
	tests := []struct {
		name           string
		environmentKey string
		value          string
		expectedText   string
	}{
		{
			name:           "wildcard origin",
			environmentKey: apiAllowedOriginsEnvironmentVariable,
			value:          "*",
			expectedText:   "must not contain wildcard origins",
		},
		{
			name:           "origin with path",
			environmentKey: apiAllowedOriginsEnvironmentVariable,
			value:          "https://web.example.com/path",
			expectedText:   "contains invalid origin",
		},
		{
			name:           "zero body limit",
			environmentKey: apiBodyLimitBytesEnvironmentVariable,
			value:          "0",
			expectedText:   "API_BODY_LIMIT_BYTES must be greater than zero",
		},
		{
			name:           "invalid read timeout",
			environmentKey: apiReadTimeoutEnvironmentVariable,
			value:          "invalid",
			expectedText:   "parse API_READ_TIMEOUT as duration",
		},
		{
			name:           "zero write timeout",
			environmentKey: apiWriteTimeoutEnvironmentVariable,
			value:          "0s",
			expectedText:   "API_WRITE_TIMEOUT must be greater than zero",
		},
		{
			name:           "zero rate limit",
			environmentKey: apiRateLimitMaxEnvironmentVariable,
			value:          "0",
			expectedText:   "API_RATE_LIMIT_MAX must be greater than zero",
		},
		{
			name:           "negative rate window",
			environmentKey: apiRateLimitWindowEnvironmentVariable,
			value:          "-1s",
			expectedText:   "API_RATE_LIMIT_WINDOW must be greater than zero",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(
				t *testing.T,
			) {
				setServerWithoutDatabaseEnvironment(
					t,
				)
				clearAPIProtectionEnvironment(
					t,
				)
				t.Setenv(
					test.environmentKey,
					test.value,
				)

				loadedConfig, err := LoadServerConfig()
				if err == nil {
					t.Fatal(
						"expected api protection configuration error",
					)
				}

				if loadedConfig != (ServerConfig{}) {
					t.Fatalf(
						"expected zero server configuration, got %+v",
						loadedConfig,
					)
				}

				if !strings.Contains(
					err.Error(),
					test.expectedText,
				) {
					t.Fatalf(
						"expected error containing %q, got %q",
						test.expectedText,
						err.Error(),
					)
				}
			},
		)
	}
}

func setServerWithoutDatabaseEnvironment(
	t *testing.T,
) {
	t.Helper()

	t.Setenv(
		apiPortEnvironmentVariable,
		"8080",
	)
	t.Setenv(
		databaseURLEnvironmentVariable,
		"",
	)
	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"",
	)
	t.Setenv(
		openMeteoTimeoutEnvironmentVariable,
		"",
	)
}

func clearAPIProtectionEnvironment(
	t *testing.T,
) {
	t.Helper()

	for _, name := range []string{
		apiAllowedOriginsEnvironmentVariable,
		apiBodyLimitBytesEnvironmentVariable,
		apiReadTimeoutEnvironmentVariable,
		apiWriteTimeoutEnvironmentVariable,
		apiIdleTimeoutEnvironmentVariable,
		apiRateLimitMaxEnvironmentVariable,
		apiRateLimitWindowEnvironmentVariable,
	} {
		t.Setenv(
			name,
			"",
		)
	}
}
