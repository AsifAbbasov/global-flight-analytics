package config

import "fmt"

const (
	migrationsDirEnvironmentVariable    = "MIGRATIONS_DIR"
	migrationTimeoutEnvironmentVariable = "MIGRATION_TIMEOUT"
)

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

	migrationsDir, err := requiredTrimmedStringEnvironmentVariable(
		migrationsDirEnvironmentVariable,
	)
	if err != nil {
		return MigrationConfig{}, fmt.Errorf(
			"load migrations directory: %w",
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
		MigrationsDir:    migrationsDir,
		MigrationTimeout: migrationTimeout,
	}, nil
}
