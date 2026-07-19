package config

import "fmt"

const (
	apiPortEnvironmentVariable          = "API_PORT"
	openMeteoTimeoutEnvironmentVariable = "OPEN_METEO_TIMEOUT"
)

func LoadServerConfig() (
	ServerConfig,
	error,
) {
	port, err := requiredTrimmedStringEnvironmentVariable(
		apiPortEnvironmentVariable,
	)
	if err != nil {
		return ServerConfig{}, fmt.Errorf(
			"load server port: %w",
			err,
		)
	}

	apiProtection, err := loadAPIProtectionConfig()
	if err != nil {
		return ServerConfig{}, fmt.Errorf(
			"load api protection configuration: %w",
			err,
		)
	}

	databaseURL := optionalTrimmedStringEnvironmentVariable(
		databaseURLEnvironmentVariable,
	)

	if databaseURL == "" {
		return ServerConfig{
			Port:          port,
			APIProtection: apiProtection,
		}, nil
	}

	databaseConnectTimeout, err := requiredPositiveDurationEnvironmentVariable(
		databaseConnectTimeoutEnvironmentVariable,
	)
	if err != nil {
		return ServerConfig{}, fmt.Errorf(
			"load database connect timeout: %w",
			err,
		)
	}

	openMeteoTimeout, err := requiredPositiveDurationEnvironmentVariable(
		openMeteoTimeoutEnvironmentVariable,
	)
	if err != nil {
		return ServerConfig{}, fmt.Errorf(
			"load open-meteo timeout: %w",
			err,
		)
	}

	if !apiProtection.MutationKeyConfigured {
		return ServerConfig{}, fmt.Errorf(
			"%s is required when DATABASE_URL is configured",
			apiMutationKeySHA256EnvironmentVariable,
		)
	}

	return ServerConfig{
		Port: port,
		Database: &PostgresConfig{
			URL:            databaseURL,
			ConnectTimeout: databaseConnectTimeout,
		},
		OpenMeteoTimeout: openMeteoTimeout,
		APIProtection:    apiProtection,
	}, nil
}

// STAGE-14-5-MUTATION-ENDPOINT-PROTECTION
