package airplaneslive

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

func TestTypedTransportClassifiesProviderStatusErrors(
	t *testing.T,
) {
	tests := []struct {
		name     string
		status   int
		expected error
	}{
		{
			name:     "rate limited",
			status:   http.StatusTooManyRequests,
			expected: integrationcommon.ErrProviderRateLimited,
		},
		{
			name:     "unauthorized",
			status:   http.StatusUnauthorized,
			expected: integrationcommon.ErrProviderUnauthorized,
		},
		{
			name:     "forbidden",
			status:   http.StatusForbidden,
			expected: integrationcommon.ErrProviderUnauthorized,
		},
		{
			name:     "client error",
			status:   http.StatusNotFound,
			expected: integrationcommon.ErrProviderClient,
		},
		{
			name:     "server error",
			status:   http.StatusBadGateway,
			expected: integrationcommon.ErrProviderServer,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				server := httptest.NewServer(
					http.HandlerFunc(
						func(
							writer http.ResponseWriter,
							_ *http.Request,
						) {
							writer.WriteHeader(
								test.status,
							)
						},
					),
				)
				defer server.Close()

				client, err := NewClient(
					integrationcommon.HTTPClientConfig{
						BaseURL:   server.URL,
						Timeout:   time.Second,
						UserAgent: "provider-status-error-test",
					},
				)
				if err != nil {
					t.Fatalf(
						"create airplanes.live client: %v",
						err,
					)
				}

				_, err = client.GetByPoint(
					context.Background(),
					40.4093,
					49.8671,
					250,
				)
				if !errors.Is(
					err,
					test.expected,
				) {
					t.Fatalf(
						"expected error %v, got %v",
						test.expected,
						err,
					)
				}
			},
		)
	}
}
