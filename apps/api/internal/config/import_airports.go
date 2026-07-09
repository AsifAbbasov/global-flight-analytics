package config

import "fmt"

const (
	ourAirportsTimeoutEnvironmentVariable = "OURAIRPORTS_TIMEOUT"

	ourAirportsCountryCodesEnvironmentVariable = "OURAIRPORTS_COUNTRY_CODES"
)

func LoadImportAirportsConfig() (
	ImportAirportsConfig,
	error,
) {
	databaseURL, err := requiredTrimmedStringEnvironmentVariable(
		databaseURLEnvironmentVariable,
	)
	if err != nil {
		return ImportAirportsConfig{}, fmt.Errorf(
			"load database url: %w",
			err,
		)
	}

	databaseConnectTimeout, err := requiredPositiveDurationEnvironmentVariable(
		databaseConnectTimeoutEnvironmentVariable,
	)
	if err != nil {
		return ImportAirportsConfig{}, fmt.Errorf(
			"load database connect timeout: %w",
			err,
		)
	}

	ourAirportsTimeout, err := requiredPositiveDurationEnvironmentVariable(
		ourAirportsTimeoutEnvironmentVariable,
	)
	if err != nil {
		return ImportAirportsConfig{}, fmt.Errorf(
			"load OurAirports timeout: %w",
			err,
		)
	}

	ourAirportsCountryCodes, err := requiredCountryCodesEnvironmentVariable(
		ourAirportsCountryCodesEnvironmentVariable,
	)
	if err != nil {
		return ImportAirportsConfig{}, fmt.Errorf(
			"load OurAirports country codes: %w",
			err,
		)
	}

	return ImportAirportsConfig{
		Database: PostgresConfig{
			URL:            databaseURL,
			ConnectTimeout: databaseConnectTimeout,
		},
		OurAirportsTimeout:      ourAirportsTimeout,
		OurAirportsCountryCodes: ourAirportsCountryCodes,
	}, nil
}
