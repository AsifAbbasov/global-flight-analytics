package opensky

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

type failingOpenSkyResponseObserver struct {
	err error
}

func (observer failingOpenSkyResponseObserver) ObserveProviderResponse(
	string,
	int,
	http.Header,
	time.Duration,
) error {
	return observer.err
}

func TestServerFailureClassificationSurvivesObserverFailure(
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

	client := mustOpenSkyResilienceClient(
		t,
		server.URL,
		failingOpenSkyResponseObserver{
			err: errors.New("observer unavailable"),
		},
	)

	_, err := client.GetStates(
		context.Background(),
		StatesRequest{},
	)
	if !errors.Is(err, integrationcommon.ErrProviderServer) {
		t.Fatalf("expected provider server error, got %v", err)
	}
}

func TestSuccessfulPayloadSurvivesObserverFailure(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = writer.Write(
					[]byte(`{"time":1720094400,"states":[]}`),
				)
			},
		),
	)
	defer server.Close()

	client := mustOpenSkyResilienceClient(
		t,
		server.URL,
		failingOpenSkyResponseObserver{
			err: errors.New("observer unavailable"),
		},
	)

	result, err := client.GetStates(
		context.Background(),
		StatesRequest{},
	)
	if err != nil {
		t.Fatalf("get states with failed observer: %v", err)
	}
	if !result.ProviderTime.Equal(
		time.Unix(1720094400, 0).UTC(),
	) {
		t.Fatalf("provider time = %s", result.ProviderTime)
	}
	if len(result.States) != 0 {
		t.Fatalf("states = %d, want 0", len(result.States))
	}
}

func TestStatesResponseRejectsDeclaredOversize(
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
					strconv.FormatInt(maxStatesResponseBytes+1, 10),
				)
				writer.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()

	client := mustOpenSkyResilienceClient(t, server.URL, nil)
	_, err := client.GetStates(
		context.Background(),
		StatesRequest{},
	)
	if !errors.Is(err, integrationcommon.ErrProviderResponseTooLarge) {
		t.Fatalf("expected oversized response error, got %v", err)
	}
}

func mustOpenSkyResilienceClient(
	t *testing.T,
	baseURL string,
	observer integrationcommon.ProviderResponseObserver,
) *Client {
	t.Helper()

	config := DefaultConfig()
	config.BaseURL = baseURL
	config.HTTPClient = &http.Client{Timeout: time.Second}
	config.PollingInterval = 10 * time.Second

	client, err := NewClientWithResponseObserver(config, observer)
	if err != nil {
		t.Fatalf("create OpenSky client: %v", err)
	}
	return client
}
