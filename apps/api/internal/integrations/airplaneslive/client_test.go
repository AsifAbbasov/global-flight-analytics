package airplaneslive

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

func TestNewClientRejectsInvalidHTTPConfiguration(
	t *testing.T,
) {
	tests := []struct {
		name        string
		config      integrationcommon.HTTPClientConfig
		expectedErr error
	}{
		{
			name: "missing base url",
			config: integrationcommon.HTTPClientConfig{
				Timeout:   time.Second,
				UserAgent: "global-flight-analytics-test",
			},
			expectedErr: integrationcommon.ErrHTTPClientBaseURLRequired,
		},
		{
			name: "zero timeout",
			config: integrationcommon.HTTPClientConfig{
				BaseURL:   "https://example.com",
				Timeout:   0,
				UserAgent: "global-flight-analytics-test",
			},
			expectedErr: integrationcommon.ErrHTTPClientTimeoutInvalid,
		},
		{
			name: "missing user agent",
			config: integrationcommon.HTTPClientConfig{
				BaseURL: "https://example.com",
				Timeout: time.Second,
			},
			expectedErr: integrationcommon.ErrHTTPClientUserAgentRequired,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				client, err := NewClient(
					test.config,
				)

				if client != nil {
					t.Fatal(
						"expected nil client for invalid configuration",
					)
				}

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

func TestGetByPointBuildsExpectedRequest(
	t *testing.T,
) {
	var receivedMethod string
	var receivedPath string

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				receivedMethod = request.Method
				receivedPath = request.URL.Path

				writer.Header().Set(
					"Content-Type",
					"application/json",
				)
				writer.WriteHeader(
					http.StatusOK,
				)

				if _, err := writer.Write(
					[]byte(`{}`),
				); err != nil {
					t.Fatalf(
						"write test response: %v",
						err,
					)
				}
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
		integrationcommon.HTTPClientConfig{
			BaseURL:   server.URL,
			Timeout:   time.Second,
			UserAgent: "global-flight-analytics-test",
		},
	)
	if err != nil {
		t.Fatalf(
			"create airplanes.live client: %v",
			err,
		)
	}

	result, err := client.GetByPoint(
		context.Background(),
		40.4093,
		49.8671,
		250,
	)
	if err != nil {
		t.Fatalf(
			"expected successful point request, got error: %v",
			err,
		)
	}

	if result == nil {
		t.Fatal(
			"expected non-nil state response",
		)
	}

	if receivedMethod != http.MethodGet {
		t.Fatalf(
			"expected HTTP method %s, got %s",
			http.MethodGet,
			receivedMethod,
		)
	}

	expectedPath := "/v2/point/40.409300/49.867100/250"

	if receivedPath != expectedPath {
		t.Fatalf(
			"expected request path %q, got %q",
			expectedPath,
			receivedPath,
		)
	}
}

func TestGetByPointRejectsInvalidInputBeforeRequest(
	t *testing.T,
) {
	requestCount := 0

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				requestCount++

				writer.Header().Set(
					"Content-Type",
					"application/json",
				)
				writer.WriteHeader(
					http.StatusOK,
				)

				if _, err := writer.Write(
					[]byte(`{}`),
				); err != nil {
					t.Fatalf(
						"write test response: %v",
						err,
					)
				}
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
		integrationcommon.HTTPClientConfig{
			BaseURL:   server.URL,
			Timeout:   time.Second,
			UserAgent: "global-flight-analytics-test",
		},
	)
	if err != nil {
		t.Fatalf(
			"create airplanes.live client: %v",
			err,
		)
	}

	tests := []struct {
		name      string
		latitude  float64
		longitude float64
		radius    int
	}{
		{
			name:      "latitude above maximum",
			latitude:  91,
			longitude: 49.8671,
			radius:    250,
		},
		{
			name:      "latitude is not finite",
			latitude:  math.NaN(),
			longitude: 49.8671,
			radius:    250,
		},
		{
			name:      "longitude above maximum",
			latitude:  40.4093,
			longitude: 181,
			radius:    250,
		},
		{
			name:      "longitude is not finite",
			latitude:  40.4093,
			longitude: math.Inf(1),
			radius:    250,
		},
		{
			name:      "radius is zero",
			latitude:  40.4093,
			longitude: 49.8671,
			radius:    0,
		},
		{
			name:      "radius is negative",
			latitude:  40.4093,
			longitude: 49.8671,
			radius:    -1,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				result, err := client.GetByPoint(
					context.Background(),
					test.latitude,
					test.longitude,
					test.radius,
				)

				if err == nil {
					t.Fatal(
						"expected validation error",
					)
				}

				if result != nil {
					t.Fatal(
						"expected nil result for invalid input",
					)
				}
			},
		)
	}

	if requestCount != 0 {
		t.Fatalf(
			"expected invalid input to produce zero HTTP requests, got %d",
			requestCount,
		)
	}
}
