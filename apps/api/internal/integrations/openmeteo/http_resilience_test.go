package openmeteo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

type failingOpenMeteoResponseObserver struct {
	err error
}

func (observer failingOpenMeteoResponseObserver) ObserveProviderResponse(
	string,
	int,
	http.Header,
	time.Duration,
) error {
	return observer.err
}

func TestServerFailurePreservesClassificationWhenObserverFails(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.WriteHeader(http.StatusInternalServerError)
			},
		),
	)
	defer server.Close()

	observerErr := errors.New("observer unavailable")
	client, err := New(
		Config{
			BaseURL:          server.URL,
			HTTPClient:       &http.Client{Timeout: time.Second},
			ResponseObserver: failingOpenMeteoResponseObserver{err: observerErr},
		},
	)
	if err != nil {
		t.Fatalf("create Open-Meteo client: %v", err)
	}

	_, err = client.GetCurrentWeather(
		context.Background(),
		CurrentWeatherRequest{
			Latitude:  40.4093,
			Longitude: 49.8671,
		},
	)
	if !errors.Is(err, integrationcommon.ErrProviderServer) {
		t.Fatalf("expected provider server error, got %v", err)
	}
	if !errors.Is(err, observerErr) {
		t.Fatalf("expected joined observer error, got %v", err)
	}
}

func TestWeatherResponseRejectsDeclaredOversize(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.Header().Set(
					"Content-Length",
					strconv.FormatInt(maxWeatherResponseBytes+1, 10),
				)
				writer.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()

	client, err := New(
		Config{
			BaseURL:    server.URL,
			HTTPClient: &http.Client{Timeout: time.Second},
		},
	)
	if err != nil {
		t.Fatalf("create Open-Meteo client: %v", err)
	}

	_, err = client.GetCurrentWeather(
		context.Background(),
		CurrentWeatherRequest{
			Latitude:  40.4093,
			Longitude: 49.8671,
		},
	)
	if !errors.Is(err, integrationcommon.ErrProviderResponseTooLarge) {
		t.Fatalf("expected oversized response error, got %v", err)
	}
}
