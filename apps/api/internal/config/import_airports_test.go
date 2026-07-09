package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadImportAirportsConfig(
	t *testing.T,
) {
	setValidImportAirportsEnvironment(
		t,
	)

	loadedConfig, err := LoadImportAirportsConfig()
	if err != nil {
		t.Fatalf(
			"expected valid import-airports configuration, got error: %v",
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

	if loadedConfig.OurAirportsTimeout != 15*time.Second {
		t.Fatalf(
			"expected OurAirports timeout %s, got %s",
			15*time.Second,
			loadedConfig.OurAirportsTimeout,
		)
	}

	expectedCountryCodes := []string{
		"AZ",
		"TR",
		"GE",
	}

	if !reflect.DeepEqual(
		loadedConfig.OurAirportsCountryCodes,
		expectedCountryCodes,
	) {
		t.Fatalf(
			"expected country codes %v, got %v",
			expectedCountryCodes,
			loadedConfig.OurAirportsCountryCodes,
		)
	}
}

func TestLoadImportAirportsConfigRejectsMissingDatabaseURL(
	t *testing.T,
) {
	setValidImportAirportsEnvironment(
		t,
	)

	t.Setenv(
		databaseURLEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadImportAirportsConfig()

	if err == nil {
		t.Fatal(
			"expected import-airports configuration error, got nil",
		)
	}

	assertZeroImportAirportsConfig(
		t,
		loadedConfig,
	)

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

func TestLoadImportAirportsConfigRejectsNonPositiveDatabaseConnectTimeout(
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
				setValidImportAirportsEnvironment(
					t,
				)

				t.Setenv(
					databaseConnectTimeoutEnvironmentVariable,
					test.value,
				)

				loadedConfig, err := LoadImportAirportsConfig()

				if err == nil {
					t.Fatal(
						"expected import-airports configuration error, got nil",
					)
				}

				assertZeroImportAirportsConfig(
					t,
					loadedConfig,
				)

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

func TestLoadImportAirportsConfigRejectsInvalidDatabaseConnectTimeout(
	t *testing.T,
) {
	setValidImportAirportsEnvironment(
		t,
	)

	t.Setenv(
		databaseConnectTimeoutEnvironmentVariable,
		"invalid-duration",
	)

	loadedConfig, err := LoadImportAirportsConfig()

	if err == nil {
		t.Fatal(
			"expected import-airports configuration error, got nil",
		)
	}

	assertZeroImportAirportsConfig(
		t,
		loadedConfig,
	)

	if !strings.Contains(
		err.Error(),
		"load database connect timeout: parse DATABASE_CONNECT_TIMEOUT as duration",
	) {
		t.Fatalf(
			"expected contextual database timeout parse error, got %q",
			err.Error(),
		)
	}
}

func TestLoadImportAirportsConfigRejectsMissingOurAirportsTimeout(
	t *testing.T,
) {
	setValidImportAirportsEnvironment(
		t,
	)

	t.Setenv(
		ourAirportsTimeoutEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadImportAirportsConfig()

	if err == nil {
		t.Fatal(
			"expected import-airports configuration error, got nil",
		)
	}

	assertZeroImportAirportsConfig(
		t,
		loadedConfig,
	)

	if !strings.Contains(
		err.Error(),
		"load OurAirports timeout: OURAIRPORTS_TIMEOUT is required",
	) {
		t.Fatalf(
			"expected contextual OurAirports timeout error, got %q",
			err.Error(),
		)
	}
}

func TestLoadImportAirportsConfigRejectsNonPositiveOurAirportsTimeout(
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
				setValidImportAirportsEnvironment(
					t,
				)

				t.Setenv(
					ourAirportsTimeoutEnvironmentVariable,
					test.value,
				)

				loadedConfig, err := LoadImportAirportsConfig()

				if err == nil {
					t.Fatal(
						"expected import-airports configuration error, got nil",
					)
				}

				assertZeroImportAirportsConfig(
					t,
					loadedConfig,
				)

				if !strings.Contains(
					err.Error(),
					"load OurAirports timeout: OURAIRPORTS_TIMEOUT must be greater than zero",
				) {
					t.Fatalf(
						"expected positive OurAirports timeout error, got %q",
						err.Error(),
					)
				}
			},
		)
	}
}

func TestLoadImportAirportsConfigRejectsInvalidOurAirportsTimeout(
	t *testing.T,
) {
	setValidImportAirportsEnvironment(
		t,
	)

	t.Setenv(
		ourAirportsTimeoutEnvironmentVariable,
		"invalid-duration",
	)

	loadedConfig, err := LoadImportAirportsConfig()

	if err == nil {
		t.Fatal(
			"expected import-airports configuration error, got nil",
		)
	}

	assertZeroImportAirportsConfig(
		t,
		loadedConfig,
	)

	if !strings.Contains(
		err.Error(),
		"load OurAirports timeout: parse OURAIRPORTS_TIMEOUT as duration",
	) {
		t.Fatalf(
			"expected contextual OurAirports timeout parse error, got %q",
			err.Error(),
		)
	}
}

func TestLoadImportAirportsConfigRejectsMissingCountryCodes(
	t *testing.T,
) {
	setValidImportAirportsEnvironment(
		t,
	)

	t.Setenv(
		ourAirportsCountryCodesEnvironmentVariable,
		"",
	)

	loadedConfig, err := LoadImportAirportsConfig()

	if err == nil {
		t.Fatal(
			"expected import-airports configuration error, got nil",
		)
	}

	assertZeroImportAirportsConfig(
		t,
		loadedConfig,
	)

	if !strings.Contains(
		err.Error(),
		"load OurAirports country codes: OURAIRPORTS_COUNTRY_CODES is required",
	) {
		t.Fatalf(
			"expected contextual country codes error, got %q",
			err.Error(),
		)
	}
}

func TestLoadImportAirportsConfigRejectsOnlyCountryCodeSeparators(
	t *testing.T,
) {
	setValidImportAirportsEnvironment(
		t,
	)

	t.Setenv(
		ourAirportsCountryCodesEnvironmentVariable,
		" , , ",
	)

	loadedConfig, err := LoadImportAirportsConfig()

	if err == nil {
		t.Fatal(
			"expected import-airports configuration error, got nil",
		)
	}

	assertZeroImportAirportsConfig(
		t,
		loadedConfig,
	)

	if !strings.Contains(
		err.Error(),
		"load OurAirports country codes: OURAIRPORTS_COUNTRY_CODES must contain at least one country code",
	) {
		t.Fatalf(
			"expected country codes validation error, got %q",
			err.Error(),
		)
	}
}

func setValidImportAirportsEnvironment(
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
		ourAirportsTimeoutEnvironmentVariable,
		"  15s  ",
	)

	t.Setenv(
		ourAirportsCountryCodesEnvironmentVariable,
		" az, tr, AZ, ge, , tr ",
	)
}

func assertZeroImportAirportsConfig(
	t *testing.T,
	loadedConfig ImportAirportsConfig,
) {
	t.Helper()

	if loadedConfig.Database != (PostgresConfig{}) {
		t.Fatalf(
			"expected zero database configuration, got %+v",
			loadedConfig.Database,
		)
	}

	if loadedConfig.OurAirportsTimeout != 0 {
		t.Fatalf(
			"expected zero OurAirports timeout, got %s",
			loadedConfig.OurAirportsTimeout,
		)
	}

	if loadedConfig.OurAirportsCountryCodes != nil {
		t.Fatalf(
			"expected nil OurAirports country codes, got %v",
			loadedConfig.OurAirportsCountryCodes,
		)
	}
}
