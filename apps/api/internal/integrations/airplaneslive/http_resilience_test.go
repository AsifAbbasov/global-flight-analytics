package airplaneslive

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

type failingAirplanesLiveResponseObserver struct {
	err error
}

func (observer failingAirplanesLiveResponseObserver) ObserveProviderResponse(
	string,
	int,
	http.Header,
	time.Duration,
) error {
	return observer.err
}

func TestServerFailurePreservesProviderClassificationWhenObserverFails(
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
	client, err := NewClientWithResponseObserver(
		integrationcommon.HTTPClientConfig{
			BaseURL:   server.URL,
			Timeout:   time.Second,
			UserAgent: "global-flight-analytics-test",
		},
		failingAirplanesLiveResponseObserver{err: observerErr},
	)
	if err != nil {
		t.Fatalf("create airplanes.live client: %v", err)
	}

	_, err = client.GetByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if !errors.Is(err, integrationcommon.ErrProviderServer) {
		t.Fatalf("expected provider server error, got %v", err)
	}
	if !errors.Is(err, observerErr) {
		t.Fatalf("expected joined observer error, got %v", err)
	}
}

func TestStateResponseRejectsDeclaredOversize(
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
					strconv.FormatInt(maxStateResponseBytes+1, 10),
				)
				writer.WriteHeader(http.StatusOK)
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
		t.Fatalf("create airplanes.live client: %v", err)
	}

	_, err = client.GetByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if !errors.Is(err, integrationcommon.ErrProviderResponseTooLarge) {
		t.Fatalf("expected oversized response error, got %v", err)
	}
}
