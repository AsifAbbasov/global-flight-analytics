package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoadHealthcheckConfigUsesDefaults(
	t *testing.T,
) {
	t.Setenv(
		"HEALTHCHECK_URL",
		"",
	)
	t.Setenv(
		"HEALTHCHECK_TIMEOUT",
		"",
	)
	t.Setenv(
		"API_PORT",
		"",
	)

	config, err := loadHealthcheckConfig()
	if err != nil {
		t.Fatalf(
			"load default healthcheck configuration: %v",
			err,
		)
	}

	if config.URL != "http://127.0.0.1:8080/api/v1/ready" {
		t.Fatalf(
			"unexpected default URL: %q",
			config.URL,
		)
	}

	if config.Timeout != defaultHealthcheckTimeout {
		t.Fatalf(
			"unexpected default timeout: %s",
			config.Timeout,
		)
	}
}

func TestLoadHealthcheckConfigUsesOverrides(
	t *testing.T,
) {
	t.Setenv(
		"HEALTHCHECK_URL",
		"https://api.example.com/health",
	)
	t.Setenv(
		"HEALTHCHECK_TIMEOUT",
		"5s",
	)
	t.Setenv(
		"API_PORT",
		"9090",
	)

	config, err := loadHealthcheckConfig()
	if err != nil {
		t.Fatalf(
			"load configured healthcheck: %v",
			err,
		)
	}

	if config.URL != "https://api.example.com/health" {
		t.Fatalf(
			"unexpected configured URL: %q",
			config.URL,
		)
	}

	if config.Timeout != 5*time.Second {
		t.Fatalf(
			"unexpected configured timeout: %s",
			config.Timeout,
		)
	}
}

func TestLoadHealthcheckConfigUsesConfiguredAPIPort(
	t *testing.T,
) {
	t.Setenv(
		"HEALTHCHECK_URL",
		"",
	)
	t.Setenv(
		"HEALTHCHECK_TIMEOUT",
		"",
	)
	t.Setenv(
		"API_PORT",
		"9090",
	)

	config, err := loadHealthcheckConfig()
	if err != nil {
		t.Fatalf(
			"load configured API port: %v",
			err,
		)
	}

	if config.URL != "http://127.0.0.1:9090/api/v1/ready" {
		t.Fatalf(
			"unexpected URL for configured API port: %q",
			config.URL,
		)
	}
}

func TestLoadHealthcheckConfigRejectsInvalidValues(
	t *testing.T,
) {
	tests := []struct {
		name          string
		url           string
		timeout       string
		apiPort       string
		expectedError error
	}{
		{
			name:          "invalid URL",
			url:           "file:///tmp/health",
			expectedError: errHealthcheckURLInvalid,
		},
		{
			name:          "invalid timeout",
			url:           "http://127.0.0.1:8080/health",
			timeout:       "0s",
			expectedError: errHealthcheckTimeoutInvalid,
		},
		{
			name:          "invalid API port",
			apiPort:       "70000",
			expectedError: errAPIPortInvalid,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(
				t *testing.T,
			) {
				t.Setenv(
					"HEALTHCHECK_URL",
					test.url,
				)
				t.Setenv(
					"HEALTHCHECK_TIMEOUT",
					test.timeout,
				)
				t.Setenv(
					"API_PORT",
					test.apiPort,
				)

				_, err := loadHealthcheckConfig()
				if !errors.Is(
					err,
					test.expectedError,
				) {
					t.Fatalf(
						"expected %v, got %v",
						test.expectedError,
						err,
					)
				}
			},
		)
	}
}

func TestCheckHealthAcceptsHealthyResponse(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.WriteHeader(
					http.StatusOK,
				)
			},
		),
	)
	defer server.Close()

	err := checkHealth(
		context.Background(),
		server.Client(),
		server.URL,
	)
	if err != nil {
		t.Fatalf(
			"check healthy endpoint: %v",
			err,
		)
	}
}

func TestCheckHealthRejectsUnhealthyResponse(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.WriteHeader(
					http.StatusServiceUnavailable,
				)
			},
		),
	)
	defer server.Close()

	err := checkHealth(
		context.Background(),
		server.Client(),
		server.URL,
	)
	if err == nil {
		t.Fatal(
			"expected unhealthy response error",
		)
	}

	if !strings.Contains(
		err.Error(),
		"unexpected status: 503",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}
