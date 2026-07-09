package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadMigrationConfig(
	t *testing.T,
) {
	setValidMigrationEnvironment(
		t,
	)

	loadedConfig, err := LoadMigrationConfig()
	if err != nil {
		t.Fatalf(
			"expected valid migration configuration, got error: %v",
			err,
		)
	}

	if loadedConfig.Database.URL != "postgresql://user:password@host/database" {
		t.Fatalf(
			"expected database url %q, got %q",
			"postgresql://user:password@host/database",
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

	if loadedConfig.MigrationTimeout != 30*time.Second {
		t.Fatalf(
			"expected migration timeout %s, got %s",
			30*time.Second,
			loadedConfig.MigrationTimeout,
		)
	}
}

func TestLoadMigrationConfigRejectsMissingDatabaseURL(
	t *testing.T,
) {
	setValidMigrationEnvironment(
		t,
	)

	t.Setenv(
		databaseURLEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadMigrationConfig()

	if err == nil {
		t.Fatal(
			"expected migration configuration error, got nil",
		)
	}

	if loadedConfig != (MigrationConfig{}) {
		t.Fatalf(
			"expected zero migration configuration, got %+v",
			loadedConfig,
		)
	}

	if !strings.Contains(
		err.Error(),
		"load database url: DATABASE_URL is required",
	) {
		t.Fatalf(
			"expected contextual database url error, got %q",
			err.Error(),
		)
	}
}

func TestLoadMigrationConfigRejectsMissingDatabaseConnectTimeout(
	t *testing.T,
) {
	setValidMigrationEnvironment(
		t,
	)

	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadMigrationConfig()

	if err == nil {
		t.Fatal(
			"expected migration configuration error, got nil",
		)
	}

	if loadedConfig != (MigrationConfig{}) {
		t.Fatalf(
			"expected zero migration configuration, got %+v",
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

func TestLoadMigrationConfigRejectsInvalidDatabaseConnectTimeout(
	t *testing.T,
) {
	setValidMigrationEnvironment(
		t,
	)

	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"invalid-duration",
	)

	loadedConfig, err := LoadMigrationConfig()

	if err == nil {
		t.Fatal(
			"expected migration configuration error, got nil",
		)
	}

	if loadedConfig != (MigrationConfig{}) {
		t.Fatalf(
			"expected zero migration configuration, got %+v",
			loadedConfig,
		)
	}

	if !strings.Contains(
		err.Error(),
		"load database connect timeout: parse DATABASE_CONNECT_TIMEOUT as duration",
	) {
		t.Fatalf(
			"expected database timeout parse error, got %q",
			err.Error(),
		)
	}
}

func TestLoadMigrationConfigRejectsNonPositiveDatabaseConnectTimeout(
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
				setValidMigrationEnvironment(
					t,
				)

				t.Setenv(
					databaseConnectTimeoutEnvironmentVariable,
					test.value,
				)

				loadedConfig, err := LoadMigrationConfig()

				if err == nil {
					t.Fatal(
						"expected migration configuration error, got nil",
					)
				}

				if loadedConfig != (MigrationConfig{}) {
					t.Fatalf(
						"expected zero migration configuration, got %+v",
						loadedConfig,
					)
				}

				if !strings.Contains(
					err.Error(),
					"load database connect timeout: DATABASE_CONNECT_TIMEOUT must be greater than zero",
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

func TestLoadMigrationConfigRejectsMissingMigrationTimeout(
	t *testing.T,
) {
	setValidMigrationEnvironment(
		t,
	)

	t.Setenv(
		migrationTimeoutEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadMigrationConfig()

	if err == nil {
		t.Fatal(
			"expected migration configuration error, got nil",
		)
	}

	if loadedConfig != (MigrationConfig{}) {
		t.Fatalf(
			"expected zero migration configuration, got %+v",
			loadedConfig,
		)
	}

	if !strings.Contains(
		err.Error(),
		"load migration timeout: MIGRATION_TIMEOUT is required",
	) {
		t.Fatalf(
			"expected contextual migration timeout error, got %q",
			err.Error(),
		)
	}
}

func TestLoadMigrationConfigRejectsInvalidMigrationTimeout(
	t *testing.T,
) {
	setValidMigrationEnvironment(
		t,
	)

	t.Setenv(
		migrationTimeoutEnvironmentVariable,
		"invalid-duration",
	)

	loadedConfig, err := LoadMigrationConfig()

	if err == nil {
		t.Fatal(
			"expected migration configuration error, got nil",
		)
	}

	if loadedConfig != (MigrationConfig{}) {
		t.Fatalf(
			"expected zero migration configuration, got %+v",
			loadedConfig,
		)
	}

	if !strings.Contains(
		err.Error(),
		"load migration timeout: parse MIGRATION_TIMEOUT as duration",
	) {
		t.Fatalf(
			"expected migration timeout parse error, got %q",
			err.Error(),
		)
	}
}

func TestLoadMigrationConfigRejectsNonPositiveMigrationTimeout(
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
				setValidMigrationEnvironment(
					t,
				)

				t.Setenv(
					migrationTimeoutEnvironmentVariable,
					test.value,
				)

				loadedConfig, err := LoadMigrationConfig()

				if err == nil {
					t.Fatal(
						"expected migration configuration error, got nil",
					)
				}

				if loadedConfig != (MigrationConfig{}) {
					t.Fatalf(
						"expected zero migration configuration, got %+v",
						loadedConfig,
					)
				}

				if !strings.Contains(
					err.Error(),
					"load migration timeout: MIGRATION_TIMEOUT must be greater than zero",
				) {
					t.Fatalf(
						"expected positive migration timeout error, got %q",
						err.Error(),
					)
				}
			},
		)
	}
}

func setValidMigrationEnvironment(
	t *testing.T,
) {
	t.Helper()

	t.Setenv(
		databaseURLEnvironmentVariable,
		"  postgresql://user:password@host/database  ",
	)

	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"  3s  ",
	)

	t.Setenv(
		migrationTimeoutEnvironmentVariable,
		"  30s  ",
	)
}
