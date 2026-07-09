package config

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRequiredTrimmedStringEnvironmentVariable(
	t *testing.T,
) {
	const environmentVariableName = "TEST_REQUIRED_TRIMMED_STRING"

	tests := []struct {
		name          string
		value         string
		expectedValue string
		expectError   bool
	}{
		{
			name:          "trims surrounding whitespace",
			value:         "  value  ",
			expectedValue: "value",
		},
		{
			name:        "rejects empty value",
			value:       "",
			expectError: true,
		},
		{
			name:        "rejects whitespace only value",
			value:       "   ",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Setenv(
					environmentVariableName,
					test.value,
				)

				value, err := requiredTrimmedStringEnvironmentVariable(
					environmentVariableName,
				)

				if test.expectError {
					if err == nil {
						t.Fatal(
							"expected validation error, got nil",
						)
					}

					return
				}

				if err != nil {
					t.Fatalf(
						"expected valid value, got error: %v",
						err,
					)
				}

				if value != test.expectedValue {
					t.Fatalf(
						"expected value %q, got %q",
						test.expectedValue,
						value,
					)
				}
			},
		)
	}
}

func TestOptionalTrimmedStringEnvironmentVariable(
	t *testing.T,
) {
	const environmentVariableName = "TEST_OPTIONAL_TRIMMED_STRING"

	t.Setenv(
		environmentVariableName,
		"  optional value  ",
	)

	value := optionalTrimmedStringEnvironmentVariable(
		environmentVariableName,
	)

	if value != "optional value" {
		t.Fatalf(
			"expected trimmed optional value %q, got %q",
			"optional value",
			value,
		)
	}
}

func TestRequiredFiniteFloat64EnvironmentVariable(
	t *testing.T,
) {
	const environmentVariableName = "TEST_REQUIRED_FINITE_FLOAT64"

	tests := []struct {
		name          string
		value         string
		expectedValue float64
		expectError   bool
	}{
		{
			name:          "accepts positive finite value",
			value:         "40.4093",
			expectedValue: 40.4093,
		},
		{
			name:          "accepts negative finite value",
			value:         "-49.8671",
			expectedValue: -49.8671,
		},
		{
			name:          "accepts zero",
			value:         "0",
			expectedValue: 0,
		},
		{
			name:        "rejects missing value",
			value:       "",
			expectError: true,
		},
		{
			name:        "rejects invalid number",
			value:       "not-a-number",
			expectError: true,
		},
		{
			name:        "rejects not a number",
			value:       "NaN",
			expectError: true,
		},
		{
			name:        "rejects positive infinity",
			value:       "+Inf",
			expectError: true,
		},
		{
			name:        "rejects negative infinity",
			value:       "-Inf",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Setenv(
					environmentVariableName,
					test.value,
				)

				value, err := requiredFiniteFloat64EnvironmentVariable(
					environmentVariableName,
				)

				if test.expectError {
					if err == nil {
						t.Fatal(
							"expected validation error, got nil",
						)
					}

					return
				}

				if err != nil {
					t.Fatalf(
						"expected valid finite value, got error: %v",
						err,
					)
				}

				if math.Abs(value-test.expectedValue) > 1e-12 {
					t.Fatalf(
						"expected value %f, got %f",
						test.expectedValue,
						value,
					)
				}
			},
		)
	}
}

func TestRequiredNonNegativeFiniteFloat64EnvironmentVariable(
	t *testing.T,
) {
	const environmentVariableName = "TEST_REQUIRED_NON_NEGATIVE_FLOAT64"

	tests := []struct {
		name          string
		value         string
		expectedValue float64
		expectError   bool
	}{
		{
			name:          "accepts zero",
			value:         "0",
			expectedValue: 0,
		},
		{
			name:          "accepts positive finite value",
			value:         "420.5",
			expectedValue: 420.5,
		},
		{
			name:        "rejects negative finite value",
			value:       "-1",
			expectError: true,
		},
		{
			name:        "rejects not a number",
			value:       "NaN",
			expectError: true,
		},
		{
			name:        "rejects infinity",
			value:       "+Inf",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Setenv(
					environmentVariableName,
					test.value,
				)

				value, err := requiredNonNegativeFiniteFloat64EnvironmentVariable(
					environmentVariableName,
				)

				if test.expectError {
					if err == nil {
						t.Fatal(
							"expected validation error, got nil",
						)
					}

					return
				}

				if err != nil {
					t.Fatalf(
						"expected valid non-negative value, got error: %v",
						err,
					)
				}

				if math.Abs(value-test.expectedValue) > 1e-12 {
					t.Fatalf(
						"expected value %f, got %f",
						test.expectedValue,
						value,
					)
				}
			},
		)
	}
}

func TestRequiredIntegerEnvironmentVariable(
	t *testing.T,
) {
	const environmentVariableName = "TEST_REQUIRED_INTEGER"

	tests := []struct {
		name          string
		value         string
		expectedValue int
		expectError   bool
	}{
		{
			name:          "accepts positive integer",
			value:         "250",
			expectedValue: 250,
		},
		{
			name:          "accepts negative integer",
			value:         "-10",
			expectedValue: -10,
		},
		{
			name:          "accepts zero",
			value:         "0",
			expectedValue: 0,
		},
		{
			name:          "trims surrounding whitespace",
			value:         "  42  ",
			expectedValue: 42,
		},
		{
			name:        "rejects missing value",
			value:       "",
			expectError: true,
		},
		{
			name:        "rejects non-integer value",
			value:       "42.5",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Setenv(
					environmentVariableName,
					test.value,
				)

				value, err := requiredIntegerEnvironmentVariable(
					environmentVariableName,
				)

				if test.expectError {
					if err == nil {
						t.Fatal(
							"expected validation error, got nil",
						)
					}

					return
				}

				if err != nil {
					t.Fatalf(
						"expected valid integer, got error: %v",
						err,
					)
				}

				if value != test.expectedValue {
					t.Fatalf(
						"expected value %d, got %d",
						test.expectedValue,
						value,
					)
				}
			},
		)
	}
}

func TestRequiredPositiveDurationEnvironmentVariable(
	t *testing.T,
) {
	const environmentVariableName = "TEST_REQUIRED_POSITIVE_DURATION"

	tests := []struct {
		name          string
		value         string
		expectedValue time.Duration
		expectError   bool
	}{
		{
			name:          "accepts positive duration",
			value:         "5s",
			expectedValue: 5 * time.Second,
		},
		{
			name:          "trims surrounding whitespace",
			value:         "  90s  ",
			expectedValue: 90 * time.Second,
		},
		{
			name:        "rejects zero duration",
			value:       "0s",
			expectError: true,
		},
		{
			name:        "rejects negative duration",
			value:       "-1s",
			expectError: true,
		},
		{
			name:        "rejects invalid duration",
			value:       "invalid-duration",
			expectError: true,
		},
		{
			name:        "rejects missing duration",
			value:       "",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Setenv(
					environmentVariableName,
					test.value,
				)

				value, err := requiredPositiveDurationEnvironmentVariable(
					environmentVariableName,
				)

				if test.expectError {
					if err == nil {
						t.Fatal(
							"expected validation error, got nil",
						)
					}

					return
				}

				if err != nil {
					t.Fatalf(
						"expected valid positive duration, got error: %v",
						err,
					)
				}

				if value != test.expectedValue {
					t.Fatalf(
						"expected duration %s, got %s",
						test.expectedValue,
						value,
					)
				}
			},
		)
	}
}

func TestRequiredNonNegativeDurationEnvironmentVariable(
	t *testing.T,
) {
	const environmentVariableName = "TEST_REQUIRED_NON_NEGATIVE_DURATION"

	tests := []struct {
		name          string
		value         string
		expectedValue time.Duration
		expectError   bool
	}{
		{
			name:          "accepts zero duration",
			value:         "0s",
			expectedValue: 0,
		},
		{
			name:          "accepts positive duration",
			value:         "90s",
			expectedValue: 90 * time.Second,
		},
		{
			name:        "rejects negative duration",
			value:       "-1s",
			expectError: true,
		},
		{
			name:        "rejects invalid duration",
			value:       "invalid-duration",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Setenv(
					environmentVariableName,
					test.value,
				)

				value, err := requiredNonNegativeDurationEnvironmentVariable(
					environmentVariableName,
				)

				if test.expectError {
					if err == nil {
						t.Fatal(
							"expected validation error, got nil",
						)
					}

					return
				}

				if err != nil {
					t.Fatalf(
						"expected valid non-negative duration, got error: %v",
						err,
					)
				}

				if value != test.expectedValue {
					t.Fatalf(
						"expected duration %s, got %s",
						test.expectedValue,
						value,
					)
				}
			},
		)
	}
}

func TestRequiredCountryCodesEnvironmentVariableNormalizesAndDeduplicates(
	t *testing.T,
) {
	const environmentVariableName = "TEST_REQUIRED_COUNTRY_CODES"

	t.Setenv(
		environmentVariableName,
		" az, tr,AZ, ge ,, tr ",
	)

	countryCodes, err := requiredCountryCodesEnvironmentVariable(
		environmentVariableName,
	)
	if err != nil {
		t.Fatalf(
			"expected valid country codes, got error: %v",
			err,
		)
	}

	expectedCountryCodes := []string{
		"AZ",
		"TR",
		"GE",
	}

	if !reflect.DeepEqual(
		countryCodes,
		expectedCountryCodes,
	) {
		t.Fatalf(
			"expected country codes %v, got %v",
			expectedCountryCodes,
			countryCodes,
		)
	}
}

func TestRequiredCountryCodesEnvironmentVariableRejectsMissingValue(
	t *testing.T,
) {
	const environmentVariableName = "TEST_REQUIRED_COUNTRY_CODES_MISSING"

	t.Setenv(
		environmentVariableName,
		"",
	)

	countryCodes, err := requiredCountryCodesEnvironmentVariable(
		environmentVariableName,
	)

	if err == nil {
		t.Fatal(
			"expected validation error, got nil",
		)
	}

	if countryCodes != nil {
		t.Fatalf(
			"expected nil country codes, got %v",
			countryCodes,
		)
	}
}

func TestRequiredCountryCodesEnvironmentVariableRejectsOnlySeparators(
	t *testing.T,
) {
	const environmentVariableName = "TEST_REQUIRED_COUNTRY_CODES_SEPARATORS"

	t.Setenv(
		environmentVariableName,
		" , , ",
	)

	countryCodes, err := requiredCountryCodesEnvironmentVariable(
		environmentVariableName,
	)

	if err == nil {
		t.Fatal(
			"expected validation error, got nil",
		)
	}

	if countryCodes != nil {
		t.Fatalf(
			"expected nil country codes, got %v",
			countryCodes,
		)
	}

	if !strings.Contains(
		err.Error(),
		"must contain at least one country code",
	) {
		t.Fatalf(
			"expected country code validation error, got %q",
			err.Error(),
		)
	}
}
