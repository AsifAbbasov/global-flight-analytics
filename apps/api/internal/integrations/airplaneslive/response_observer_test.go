package airplaneslive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

type responseObserverStub struct {
	callCount    int
	providerName string
	statusCode   int
	retryAfter   string
}

func (stub *responseObserverStub) ObserveProviderResponse(
	providerName string,
	statusCode int,
	headers http.Header,
) error {
	stub.callCount++
	stub.providerName = providerName
	stub.statusCode = statusCode
	stub.retryAfter = headers.Get(
		"Retry-After",
	)

	return nil
}

func TestHTTPResponseMetadataIsObserved(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.Header().Set(
					"Retry-After",
					"7",
				)

				writer.WriteHeader(
					http.StatusTooManyRequests,
				)
			},
		),
	)
	defer server.Close()

	observer := &responseObserverStub{}

	client, err := NewClientWithResponseObserver(
		integrationcommon.HTTPClientConfig{
			BaseURL:   server.URL,
			Timeout:   time.Second,
			UserAgent: "global-flight-analytics-test",
		},
		observer,
	)
	if err != nil {
		t.Fatalf(
			"create airplanes.live client with response observer: %v",
			err,
		)
	}

	_, err = client.GetByPoint(
		context.Background(),
		40.4093,
		49.8671,
		250,
	)
	if err == nil {
		t.Fatal(
			"expected rate-limited request error",
		)
	}

	if observer.callCount != 1 {
		t.Fatalf(
			"expected one response observation, got %d",
			observer.callCount,
		)
	}

	if observer.providerName != sourceName {
		t.Fatalf(
			"expected provider %s, got %s",
			sourceName,
			observer.providerName,
		)
	}

	if observer.statusCode != http.StatusTooManyRequests {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTooManyRequests,
			observer.statusCode,
		)
	}

	if observer.retryAfter != "7" {
		t.Fatalf(
			"expected Retry-After value 7, got %q",
			observer.retryAfter,
		)
	}
}
