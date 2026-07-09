package config

import (
	"strings"
	"testing"
)

func TestLoadVerifyAirportsConfig(
	t *testing.T,
) {
	t.Setenv(
		databaseURLEnvironmentVariable,
		"  postgresql://user:password@host/database  ",
	)

	loadedConfig, err := LoadVerifyAirportsConfig()
	if err != nil {
		t.Fatalf(
			"expected valid verify-airports configuration, got error: %v",
			err,
		)
	}

	if loadedConfig.DatabaseURL != "postgresql://user:password@host/database" {
		t.Fatalf(
			"expected trimmed database url %q, got %q",
			"postgresql://user:password@host/database",
			loadedConfig.DatabaseURL,
		)
	}
}

func TestLoadVerifyAirportsConfigRejectsMissingDatabaseURL(
	t *testing.T,
) {
	t.Setenv(
		databaseURLEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadVerifyAirportsConfig()

	if err == nil {
		t.Fatal(
			"expected verify-airports configuration error, got nil",
		)
	}

	if loadedConfig != (VerifyAirportsConfig{}) {
		t.Fatalf(
			"expected zero verify-airports configuration, got %+v",
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

func TestLoadVerifyAirportsConfigRejectsWhitespaceOnlyDatabaseURL(
	t *testing.T,
) {
	t.Setenv(
		databaseURLEnvironmentVariable,
		"   ",
	)

	loadedConfig, err := LoadVerifyAirportsConfig()

	if err == nil {
		t.Fatal(
			"expected verify-airports configuration error, got nil",
		)
	}

	if loadedConfig != (VerifyAirportsConfig{}) {
		t.Fatalf(
			"expected zero verify-airports configuration, got %+v",
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
