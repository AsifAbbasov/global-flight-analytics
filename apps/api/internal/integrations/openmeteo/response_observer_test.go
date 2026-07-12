package openmeteo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

	client, err := New(
		Config{
			BaseURL:          server.URL,
			Timeout:          time.Second,
			ResponseObserver: observer,
		},
	)
	if err != nil {
		t.Fatalf(
			"create open-meteo client: %v",
			err,
		)
	}

	_, err = client.GetCurrentWeather(
		context.Background(),
		CurrentWeatherRequest{
			Latitude:  40.4093,
			Longitude: 49.8671,
		},
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

	if observer.providerName != "open_meteo" {
		t.Fatalf(
			"expected provider open_meteo, got %s",
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

	client, err := New(
		Config{
			BaseURL: "https://example.test",
			HTTPClient: &http.Client{
				Transport: responseObserverRoundTripper(
					func(
						*http.Request,
					) (*http.Response, error) {
						return nil, expectedFailure
					},
				),
			},
			ResponseObserver: observer,
		},
	)
	if err != nil {
		t.Fatalf("create open-meteo client: %v", err)
	}

	_, err = client.GetCurrentWeather(
		context.Background(),
		CurrentWeatherRequest{
			Latitude:  40.4093,
			Longitude: 49.8671,
		},
	)
	if err == nil {
		t.Fatal("expected transport failure")
	}

	if observer.transportFailureCallCount != 1 {
		t.Fatalf(
			"transport failure observations = %d, want 1",
			observer.transportFailureCallCount,
		)
	}
	if observer.transportFailureProvider != "open_meteo" {
		t.Fatalf(
			"transport failure provider = %q, want open_meteo",
			observer.transportFailureProvider,
		)
	}
	if !errors.Is(
		observer.transportFailure,
		expectedFailure,
	) {
		t.Fatalf(
			"observed failure = %v, want %v",
			observer.transportFailure,
			expectedFailure,
		)
	}
	if observer.transportFailureLatency < 0 {
		t.Fatalf(
			"transport failure latency = %s, want non-negative",
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

	client, err := New(
		Config{
			BaseURL:          server.URL,
			Timeout:          time.Second,
			ResponseObserver: observer,
		},
	)
	if err != nil {
		t.Fatalf("create open-meteo client: %v", err)
	}

	_, err = client.GetCurrentWeather(
		context.Background(),
		CurrentWeatherRequest{
			Latitude:  40.4093,
			Longitude: 49.8671,
		},
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
	if observer.responseFailureProvider != "open_meteo" {
		t.Fatalf(
			"response failure provider = %q, want open_meteo",
			observer.responseFailureProvider,
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
