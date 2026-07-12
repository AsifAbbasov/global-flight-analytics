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

type responseObserverStub struct {
	callCount    int
	providerName string
	statusCode   int
	retryAfter   string
	latency      time.Duration

	transportFailureCallCount int
	transportFailureProvider  string
	transportFailure          error
	transportFailureLatency   time.Duration

	responseFailureCallCount int
	responseFailureProvider  string
	responseFailure          error
	responseFailureLatency   time.Duration
}

func (stub *responseObserverStub) ObserveProviderResponse(
	providerName string,
	statusCode int,
	headers http.Header,
	latency time.Duration,
) error {
	stub.callCount++
	stub.providerName = providerName
	stub.statusCode = statusCode
	stub.retryAfter = headers.Get(
		"Retry-After",
	)
	stub.latency = latency

	return nil
}

func (stub *responseObserverStub) ObserveProviderTransportFailure(
	providerName string,
	requestErr error,
	latency time.Duration,
) error {
	stub.transportFailureCallCount++
	stub.transportFailureProvider = providerName
	stub.transportFailure = requestErr
	stub.transportFailureLatency = latency

	return nil
}

func (stub *responseObserverStub) ObserveProviderResponseFailure(
	providerName string,
	responseErr error,
	latency time.Duration,
) error {
	stub.responseFailureCallCount++
	stub.responseFailureProvider = providerName
	stub.responseFailure = responseErr
	stub.responseFailureLatency = latency

	return nil
}

type responseObserverRoundTripper func(
	request *http.Request,
) (*http.Response, error)

func (roundTripper responseObserverRoundTripper) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return roundTripper(request)
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

	if observer.latency <= 0 {
		t.Fatalf(
			"expected positive response latency, got %s",
			observer.latency,
		)
	}
}

func TestTransportFailureIsObserved(
	t *testing.T,
) {
	expectedFailure := errors.New(
		"network unavailable",
	)
	observer := &responseObserverStub{}

	client, err := NewClientWithResponseObserver(
		integrationcommon.HTTPClientConfig{
			BaseURL:   "https://example.test",
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

	client.httpClient.Transport = responseObserverRoundTripper(
		func(
			*http.Request,
		) (*http.Response, error) {
			return nil, expectedFailure
		},
	)

	_, err = client.GetByPoint(
		context.Background(),
		40.4093,
		49.8671,
		250,
	)
	if err == nil {
		t.Fatal(
			"expected transport failure",
		)
	}

	if observer.transportFailureCallCount != 1 {
		t.Fatalf(
			"expected one transport failure observation, got %d",
			observer.transportFailureCallCount,
		)
	}
	if observer.transportFailureProvider != sourceName {
		t.Fatalf(
			"expected provider %s, got %s",
			sourceName,
			observer.transportFailureProvider,
		)
	}
	if !errors.Is(
		observer.transportFailure,
		expectedFailure,
	) {
		t.Fatalf(
			"expected observed failure %v, got %v",
			expectedFailure,
			observer.transportFailure,
		)
	}
	if observer.transportFailureLatency < 0 {
		t.Fatalf(
			"expected non-negative failure latency, got %s",
			observer.transportFailureLatency,
		)
	}
}

func TestMalformedSuccessResponseIsObservedAsFailure(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.Header().Set(
					"Content-Type",
					"application/json",
				)
				writer.WriteHeader(http.StatusOK)
				_, _ = writer.Write([]byte("{"))
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
		t.Fatal("expected malformed response failure")
	}

	if observer.callCount != 0 {
		t.Fatalf(
			"successful response observations = %d, want 0",
			observer.callCount,
		)
	}
	if observer.responseFailureCallCount != 1 {
		t.Fatalf(
			"response failure observations = %d, want 1",
			observer.responseFailureCallCount,
		)
	}
	if observer.responseFailureProvider != sourceName {
		t.Fatalf(
			"response failure provider = %q, want %q",
			observer.responseFailureProvider,
			sourceName,
		)
	}
	if observer.responseFailure == nil {
		t.Fatal("expected observed response error")
	}
	if observer.responseFailureLatency < 0 {
		t.Fatalf(
			"response failure latency = %s, want non-negative",
			observer.responseFailureLatency,
		)
	}
}
