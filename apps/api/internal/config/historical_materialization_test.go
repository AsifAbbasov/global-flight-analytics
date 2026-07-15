package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadHistoricalMaterializationConfig(
	t *testing.T,
) {
	setValidHistoricalMaterializationEnvironment(
		t,
	)
	t.Setenv(
		historicalMaterializationTimeoutEnvironmentVariable,
		"  7m  ",
	)

	loaded, err :=
		LoadHistoricalMaterializationConfig()
	if err != nil {
		t.Fatalf(
			"load Historical Materialization configuration: %v",
			err,
		)
	}

	if loaded.Database.URL !=
		"postgresql://user:password@host/database" {
		t.Fatalf(
			"database url = %q",
			loaded.Database.URL,
		)
	}
	if loaded.Database.ConnectTimeout !=
		3*time.Second {
		t.Fatalf(
			"database connect timeout = %s",
			loaded.Database.ConnectTimeout,
		)
	}
	if loaded.OperationTimeout !=
		7*time.Minute {
		t.Fatalf(
			"operation timeout = %s",
			loaded.OperationTimeout,
		)
	}
}

func TestLoadHistoricalMaterializationConfigUsesDefaultTimeout(
	t *testing.T,
) {
	setValidHistoricalMaterializationEnvironment(
		t,
	)
	t.Setenv(
		historicalMaterializationTimeoutEnvironmentVariable,
		"",
	)

	loaded, err :=
		LoadHistoricalMaterializationConfig()
	if err != nil {
		t.Fatalf(
			"load Historical Materialization configuration: %v",
			err,
		)
	}
	if loaded.OperationTimeout !=
		DefaultHistoricalMaterializationTimeout {
		t.Fatalf(
			"operation timeout = %s, want %s",
			loaded.OperationTimeout,
			DefaultHistoricalMaterializationTimeout,
		)
	}
}

func TestLoadHistoricalMaterializationConfigRejectsInvalidEnvironment(
	t *testing.T,
) {
	tests := []struct {
		name          string
		variable      string
		value         string
		errorFragment string
	}{
		{
			name:          "database url required",
			variable:      databaseURLEnvironmentVariable,
			value:         "",
			errorFragment: "load database url: DATABASE_URL is required",
		},
		{
			name:          "database connect timeout required",
			variable:      databaseConnectTimeoutEnvironmentVariable,
			value:         "",
			errorFragment: "load database connect timeout: DATABASE_CONNECT_TIMEOUT is required",
		},
		{
			name:          "operation timeout parses",
			variable:      historicalMaterializationTimeoutEnvironmentVariable,
			value:         "invalid",
			errorFragment: "parse HISTORICAL_MATERIALIZATION_TIMEOUT as duration",
		},
		{
			name:          "operation timeout positive",
			variable:      historicalMaterializationTimeoutEnvironmentVariable,
			value:         "0s",
			errorFragment: "HISTORICAL_MATERIALIZATION_TIMEOUT must be greater than zero",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				setValidHistoricalMaterializationEnvironment(
					t,
				)
				t.Setenv(
					test.variable,
					test.value,
				)

				loaded, err :=
					LoadHistoricalMaterializationConfig()
				if err == nil {
					t.Fatal(
						"expected configuration error",
					)
				}
				if loaded !=
					(HistoricalMaterializationConfig{}) {
					t.Fatalf(
						"loaded non-zero configuration: %#v",
						loaded,
					)
				}
				if !strings.Contains(
					err.Error(),
					test.errorFragment,
				) {
					t.Fatalf(
						"error = %q, want fragment %q",
						err.Error(),
						test.errorFragment,
					)
				}
			},
		)
	}
}

func setValidHistoricalMaterializationEnvironment(
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
		historicalMaterializationTimeoutEnvironmentVariable,
		"",
	)
}
