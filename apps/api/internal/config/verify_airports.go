package config

import "fmt"

func LoadVerifyAirportsConfig() (
	VerifyAirportsConfig,
	error,
) {
	databaseURL, err := requiredTrimmedStringEnvironmentVariable(
		databaseURLEnvironmentVariable,
	)
	if err != nil {
		return VerifyAirportsConfig{}, fmt.Errorf(
			"load database url: %w",
			err,
		)
	}

	return VerifyAirportsConfig{
		DatabaseURL: databaseURL,
	}, nil
}
