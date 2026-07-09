package config

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

func requiredTrimmedStringEnvironmentVariable(
	name string,
) (string, error) {
	value := optionalTrimmedStringEnvironmentVariable(
		name,
	)

	if value == "" {
		return "", fmt.Errorf(
			"%s is required",
			name,
		)
	}

	return value, nil
}

func optionalTrimmedStringEnvironmentVariable(
	name string,
) string {
	return strings.TrimSpace(
		os.Getenv(
			name,
		),
	)
}

func requiredFiniteFloat64EnvironmentVariable(
	name string,
) (float64, error) {
	value, err := requiredTrimmedStringEnvironmentVariable(
		name,
	)
	if err != nil {
		return 0, err
	}

	parsed, err := strconv.ParseFloat(
		value,
		64,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"parse %s as float64: %w",
			name,
			err,
		)
	}

	if math.IsNaN(parsed) ||
		math.IsInf(parsed, 0) {
		return 0, fmt.Errorf(
			"%s must be a finite value",
			name,
		)
	}

	return parsed, nil
}

func requiredNonNegativeFiniteFloat64EnvironmentVariable(
	name string,
) (float64, error) {
	value, err := requiredFiniteFloat64EnvironmentVariable(
		name,
	)
	if err != nil {
		return 0, err
	}

	if value < 0 {
		return 0, fmt.Errorf(
			"%s must be non-negative",
			name,
		)
	}

	return value, nil
}

func requiredIntegerEnvironmentVariable(
	name string,
) (int, error) {
	value, err := requiredTrimmedStringEnvironmentVariable(
		name,
	)
	if err != nil {
		return 0, err
	}

	parsed, err := strconv.Atoi(
		value,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"parse %s as integer: %w",
			name,
			err,
		)
	}

	return parsed, nil
}

func requiredPositiveDurationEnvironmentVariable(
	name string,
) (time.Duration, error) {
	value, err := requiredParsedDurationEnvironmentVariable(
		name,
	)
	if err != nil {
		return 0, err
	}

	if value <= 0 {
		return 0, fmt.Errorf(
			"%s must be greater than zero",
			name,
		)
	}

	return value, nil
}

func requiredNonNegativeDurationEnvironmentVariable(
	name string,
) (time.Duration, error) {
	value, err := requiredParsedDurationEnvironmentVariable(
		name,
	)
	if err != nil {
		return 0, err
	}

	if value < 0 {
		return 0, fmt.Errorf(
			"%s must be non-negative",
			name,
		)
	}

	return value, nil
}

func requiredParsedDurationEnvironmentVariable(
	name string,
) (time.Duration, error) {
	value, err := requiredTrimmedStringEnvironmentVariable(
		name,
	)
	if err != nil {
		return 0, err
	}

	parsed, err := time.ParseDuration(
		value,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"parse %s as duration: %w",
			name,
			err,
		)
	}

	return parsed, nil
}

func requiredCountryCodesEnvironmentVariable(
	name string,
) ([]string, error) {
	value, err := requiredTrimmedStringEnvironmentVariable(
		name,
	)
	if err != nil {
		return nil, err
	}

	rawCountryCodes := strings.Split(
		value,
		",",
	)

	countryCodes := make(
		[]string,
		0,
		len(rawCountryCodes),
	)

	seenCountryCodes := make(
		map[string]struct{},
		len(rawCountryCodes),
	)

	for _, rawCountryCode := range rawCountryCodes {
		countryCode := strings.ToUpper(
			strings.TrimSpace(
				rawCountryCode,
			),
		)

		if countryCode == "" {
			continue
		}

		if _, exists := seenCountryCodes[countryCode]; exists {
			continue
		}

		seenCountryCodes[countryCode] = struct{}{}

		countryCodes = append(
			countryCodes,
			countryCode,
		)
	}

	if len(countryCodes) == 0 {
		return nil, fmt.Errorf(
			"%s must contain at least one country code",
			name,
		)
	}

	return countryCodes, nil
}
