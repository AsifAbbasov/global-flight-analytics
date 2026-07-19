package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadServerConfigRejectsMissingAPIPort(
	t *testing.T,
) {
	t.Setenv(
		apiPortEnvironmentVariable,
		"",
	)
	t.Setenv(
		databaseURLEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadServerConfig()

	if err == nil {
		t.Fatal(
			"expected server configuration error, got nil",
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
		"load server port: API_PORT is required",
	) {
		t.Fatalf(
			"expected contextual api port error, got %q",
			err.Error(),
		)
	}

}

func TestLoadServerConfigWithoutDatabase(
	t *testing.T,
) {
	t.Setenv(
		apiPortEnvironmentVariable,
		" 8080 ",
	)
	t.Setenv(
		databaseURLEnvironmentVariable,
		" ",
	)
	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"",
	)
	t.Setenv(
		openMeteoTimeoutEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadServerConfig()
	if err != nil {
		t.Fatalf(
			"expected database-disabled server configuration, got error: %v",
			err,
		)
	}

	if loadedConfig.Port != "8080" {
		t.Fatalf(
			"expected trimmed api port %q, got %q",
			"8080",
			loadedConfig.Port,
		)
	}

	if loadedConfig.Database != nil {
		t.Fatalf(
			"expected nil database configuration, got %+v",
			loadedConfig.Database,
		)
	}

	if loadedConfig.OpenMeteoTimeout != 0 {
		t.Fatalf(
			"expected zero open-meteo timeout without database, got %s",
			loadedConfig.OpenMeteoTimeout,
		)
	}

}

func TestLoadServerConfigRejectsMissingDatabaseConnectTimeout(
	t *testing.T,
) {
	t.Setenv(
		apiPortEnvironmentVariable,
		"8080",
	)
	t.Setenv(
		databaseURLEnvironmentVariable,
		"postgresql://user:password@host/database",
	)
	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"",
	)
	t.Setenv(
		openMeteoTimeoutEnvironmentVariable,
		"5s",
	)

	loadedConfig, err := LoadServerConfig()

	if err == nil {
		t.Fatal(
			"expected server configuration error, got nil",
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
		"load database connect timeout: DATABASE_CONNECT_TIMEOUT is required",
	) {
		t.Fatalf(
			"expected contextual database timeout error, got %q",
			err.Error(),
		)
	}

}

func TestLoadServerConfigRejectsNonPositiveDatabaseConnectTimeout(
	t *testing.T,
) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "zero timeout",
			value: "0s",
		},
		{
			name:  "negative timeout",
			value: "-1s",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Setenv(
					apiPortEnvironmentVariable,
					"8080",
				)
				t.Setenv(
					databaseURLEnvironmentVariable,
					"postgresql://user:password@host/database",
				)
				t.Setenv(
					databaseConnectTimeoutEnvironmentVariable,
					test.value,
				)
				t.Setenv(
					openMeteoTimeoutEnvironmentVariable,
					"5s",
				)

				loadedConfig, err := LoadServerConfig()

				if err == nil {
					t.Fatal(
						"expected server configuration error, got nil",
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
					"DATABASE_CONNECT_TIMEOUT must be greater than zero",
				) {
					t.Fatalf(
						"expected positive database timeout error, got %q",
						err.Error(),
					)
				}
			},
		)
	}

}

func TestLoadServerConfigRejectsMissingOpenMeteoTimeout(
	t *testing.T,
) {
	t.Setenv(
		apiPortEnvironmentVariable,
		"8080",
	)
	t.Setenv(
		databaseURLEnvironmentVariable,
		"postgresql://user:password@host/database",
	)
	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"3s",
	)
	t.Setenv(
		openMeteoTimeoutEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadServerConfig()

	if err == nil {
		t.Fatal(
			"expected server configuration error, got nil",
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
		"load open-meteo timeout: OPEN_METEO_TIMEOUT is required",
	) {
		t.Fatalf(
			"expected contextual open-meteo timeout error, got %q",
			err.Error(),
		)
	}

}

func TestLoadServerConfigRejectsNonPositiveOpenMeteoTimeout(
	t *testing.T,
) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "zero timeout",
			value: "0s",
		},
		{
			name:  "negative timeout",
			value: "-1s",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Setenv(
					apiPortEnvironmentVariable,
					"8080",
				)
				t.Setenv(
					databaseURLEnvironmentVariable,
					"postgresql://user:password@host/database",
				)
				t.Setenv(
					databaseConnectTimeoutEnvironmentVariable,
					"3s",
				)
				t.Setenv(
					openMeteoTimeoutEnvironmentVariable,
					test.value,
				)

				loadedConfig, err := LoadServerConfig()

				if err == nil {
					t.Fatal(
						"expected server configuration error, got nil",
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
					"OPEN_METEO_TIMEOUT must be greater than zero",
				) {
					t.Fatalf(
						"expected positive open-meteo timeout error, got %q",
						err.Error(),
					)
				}
			},
		)
	}

}

func TestLoadServerConfigWithDatabase(
	t *testing.T,
) {
	t.Setenv(
		apiPortEnvironmentVariable,
		" 8080 ",
	)
	t.Setenv(
		databaseURLEnvironmentVariable,
		" postgresql://user:password@host/database ",
	)
	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		" 3s ",
	)
	t.Setenv(
		openMeteoTimeoutEnvironmentVariable,
		" 5s ",
	)
	t.Setenv(
		apiMutationKeySHA256EnvironmentVariable,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	)

	loadedConfig, err := LoadServerConfig()
	if err != nil {
		t.Fatalf(
			"expected valid database-backed server configuration, got error: %v",
			err,
		)
	}

	if loadedConfig.Port != "8080" {
		t.Fatalf(
			"expected trimmed api port %q, got %q",
			"8080",
			loadedConfig.Port,
		)
	}

	if loadedConfig.Database == nil {
		t.Fatal(
			"expected database configuration",
		)
	}

	if loadedConfig.Database.URL != "postgresql://user:password@host/database" {
		t.Fatalf(
			"expected trimmed database url, got %q",
			loadedConfig.Database.URL,
		)
	}

	if loadedConfig.Database.ConnectTimeout != 3*time.Second {
		t.Fatalf(
			"expected database connect timeout %s, got %s",
			3*time.Second,
			loadedConfig.Database.ConnectTimeout,
		)
	}

	if loadedConfig.OpenMeteoTimeout != 5*time.Second {
		t.Fatalf(
			"expected open-meteo timeout %s, got %s",
			5*time.Second,
			loadedConfig.OpenMeteoTimeout,
		)
	}

}

// STAGE-14-5-MUTATION-ENDPOINT-PROTECTION
