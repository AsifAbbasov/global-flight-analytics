package config

import "fmt"

const migrationTimeoutEnvironmentVariable = "MIGRATION_TIMEOUT"

func LoadMigrationConfig() (
	MigrationConfig,
	error,
) {
	databaseURL, err := requiredTrimmedStringEnvironmentVariable(
		databaseURLEnvironmentVariable,
	)
	if err != nil {
		return MigrationConfig{}, fmt.Errorf(
			"load database url: %w",
			err,
		)
	}

	databaseConnectTimeout, err := requiredPositiveDurationEnvironmentVariable(
		databaseConnectTimeoutEnvironmentVariable,
	)
	if err != nil {
		return MigrationConfig{}, fmt.Errorf(
			"load database connect timeout: %w",
			err,
		)
	}

	migrationTimeout, err := requiredPositiveDurationEnvironmentVariable(
		migrationTimeoutEnvironmentVariable,
	)
	if err != nil {
		return MigrationConfig{}, fmt.Errorf(
			"load migration timeout: %w",
			err,
		)
	}

	return MigrationConfig{
		Database: PostgresConfig{
			URL:            databaseURL,
			ConnectTimeout: databaseConnectTimeout,
		},
		MigrationTimeout: migrationTimeout,
	}, nil
}
