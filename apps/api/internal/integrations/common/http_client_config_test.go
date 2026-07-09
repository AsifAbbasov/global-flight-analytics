package integrations

import (
	"errors"
	"testing"
	"time"
)

func TestHTTPClientConfigValidateAcceptsValidConfiguration(
	t *testing.T,
) {
	config := HTTPClientConfig{
		BaseURL:   "https://example.com",
		Timeout:   time.Second,
		UserAgent: "global-flight-analytics-test",
	}

	if err := config.Validate(); err != nil {
		t.Fatalf(
			"validate http client config: %v",
			err,
		)
	}
}

func TestHTTPClientConfigValidateRejectsInvalidConfiguration(
	t *testing.T,
) {
	tests := []struct {
		name        string
		config      HTTPClientConfig
		expectedErr error
	}{
		{
			name: "base url is empty",
			config: HTTPClientConfig{
				Timeout:   time.Second,
				UserAgent: "test-agent",
			},
			expectedErr: ErrHTTPClientBaseURLRequired,
		},
		{
			name: "base url is whitespace",
			config: HTTPClientConfig{
				BaseURL:   "   ",
				Timeout:   time.Second,
				UserAgent: "test-agent",
			},
			expectedErr: ErrHTTPClientBaseURLRequired,
		},
		{
			name: "base url is relative",
			config: HTTPClientConfig{
				BaseURL:   "/v2",
				Timeout:   time.Second,
				UserAgent: "test-agent",
			},
			expectedErr: ErrHTTPClientBaseURLInvalid,
		},
		{
			name: "base url has unsupported scheme",
			config: HTTPClientConfig{
				BaseURL:   "ftp://example.com",
				Timeout:   time.Second,
				UserAgent: "test-agent",
			},
			expectedErr: ErrHTTPClientBaseURLInvalid,
		},
		{
			name: "timeout is zero",
			config: HTTPClientConfig{
				BaseURL:   "https://example.com",
				Timeout:   0,
				UserAgent: "test-agent",
			},
			expectedErr: ErrHTTPClientTimeoutInvalid,
		},
		{
			name: "timeout is negative",
			config: HTTPClientConfig{
				BaseURL:   "https://example.com",
				Timeout:   -time.Second,
				UserAgent: "test-agent",
			},
			expectedErr: ErrHTTPClientTimeoutInvalid,
		},
		{
			name: "user agent is empty",
			config: HTTPClientConfig{
				BaseURL: "https://example.com",
				Timeout: time.Second,
			},
			expectedErr: ErrHTTPClientUserAgentRequired,
		},
		{
			name: "user agent is whitespace",
			config: HTTPClientConfig{
				BaseURL:   "https://example.com",
				Timeout:   time.Second,
				UserAgent: "   ",
			},
			expectedErr: ErrHTTPClientUserAgentRequired,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				err := test.config.Validate()

				if !errors.Is(
					err,
					test.expectedErr,
				) {
					t.Fatalf(
						"expected error %v, got %v",
						test.expectedErr,
						err,
					)
				}
			},
		)
	}
}
