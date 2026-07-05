package openmeteo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
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

	client, err := New(
		Config{
			BaseURL:          server.URL,
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
}
