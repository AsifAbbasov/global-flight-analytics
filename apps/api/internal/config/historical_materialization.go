package config

import (
	"fmt"
	"time"
)

const (
	historicalMaterializationTimeoutEnvironmentVariable = "HISTORICAL_MATERIALIZATION_TIMEOUT"

	DefaultHistoricalMaterializationTimeout = 5 * time.Minute
)

func LoadHistoricalMaterializationConfig() (
	HistoricalMaterializationConfig,
	error,
) {
	databaseURL, err :=
		requiredTrimmedStringEnvironmentVariable(
			databaseURLEnvironmentVariable,
		)
	if err != nil {
		return HistoricalMaterializationConfig{},
			fmt.Errorf(
				"load database url: %w",
				err,
			)
	}

	databaseConnectTimeout, err :=
		requiredPositiveDurationEnvironmentVariable(
			databaseConnectTimeoutEnvironmentVariable,
		)
	if err != nil {
		return HistoricalMaterializationConfig{},
			fmt.Errorf(
				"load database connect timeout: %w",
				err,
			)
	}

	operationTimeout :=
		DefaultHistoricalMaterializationTimeout
	rawOperationTimeout :=
		optionalTrimmedStringEnvironmentVariable(
			historicalMaterializationTimeoutEnvironmentVariable,
		)
	if rawOperationTimeout != "" {
		operationTimeout, err = time.ParseDuration(
			rawOperationTimeout,
		)
		if err != nil {
			return HistoricalMaterializationConfig{},
				fmt.Errorf(
					"parse %s as duration: %w",
					historicalMaterializationTimeoutEnvironmentVariable,
					err,
				)
		}
		if operationTimeout <= 0 {
			return HistoricalMaterializationConfig{},
				fmt.Errorf(
					"%s must be greater than zero",
					historicalMaterializationTimeoutEnvironmentVariable,
				)
		}
	}

	return HistoricalMaterializationConfig{
		Database: PostgresConfig{
			URL:            databaseURL,
			ConnectTimeout: databaseConnectTimeout,
		},
		OperationTimeout: operationTimeout,
	}, nil
}
