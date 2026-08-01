package openmeteo

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNewUsesDefaultRequestTimeoutForUnboundedInjectedClient(
	t *testing.T,
) {
	injectedClient := &http.Client{}
	client, err := New(Config{HTTPClient: injectedClient})
	if err != nil {
		t.Fatalf("create Open-Meteo client: %v", err)
	}
	if client.httpClient != injectedClient {
		t.Fatal("expected injected HTTP client instance to be preserved")
	}
	if injectedClient.Timeout != 0 {
		t.Fatalf("constructor mutated injected timeout to %s", injectedClient.Timeout)
	}
	if client.requestTimeout != defaultHTTPTimeout {
		t.Fatalf(
			"expected default request timeout %s, got %s",
			defaultHTTPTimeout,
			client.requestTimeout,
		)
	}
}

func TestNewUsesConfiguredRequestTimeoutWithoutMutatingInjectedClient(
	t *testing.T,
) {
	const timeout = 6 * time.Second
	injectedClient := &http.Client{}
	client, err := New(Config{HTTPClient: injectedClient, Timeout: timeout})
	if err != nil {
		t.Fatalf("create Open-Meteo client: %v", err)
	}
	if client.httpClient != injectedClient {
		t.Fatal("expected injected HTTP client instance to be preserved")
	}
	if injectedClient.Timeout != 0 {
		t.Fatalf("constructor mutated injected timeout to %s", injectedClient.Timeout)
	}
	if client.requestTimeout != timeout {
		t.Fatalf("expected request timeout %s, got %s", timeout, client.requestTimeout)
	}
}

func TestNewUsesInjectedClientTimeoutWhenConfigTimeoutIsOmitted(
	t *testing.T,
) {
	const timeout = 4 * time.Second
	injectedClient := &http.Client{Timeout: timeout}
	client, err := New(Config{HTTPClient: injectedClient})
	if err != nil {
		t.Fatalf("create Open-Meteo client: %v", err)
	}
	if client.httpClient != injectedClient {
		t.Fatal("expected injected HTTP client instance to be preserved")
	}
	if client.requestTimeout != timeout {
		t.Fatalf("expected request timeout %s, got %s", timeout, client.requestTimeout)
	}
}

func TestInjectedClientRequestUsesConfiguredDeadline(
	t *testing.T,
) {
	const timeout = 10 * time.Millisecond

	injectedClient := &http.Client{
		Transport: timeoutContractRoundTripper(
			func(
				request *http.Request,
			) (*http.Response, error) {
				<-request.Context().Done()
				return nil, request.Context().Err()
			},
		),
	}

	client, err := New(
		Config{
			BaseURL:    "https://example.test",
			HTTPClient: injectedClient,
			Timeout:    timeout,
		},
	)
	if err != nil {
		t.Fatalf(
			"create Open-Meteo client: %v",
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
	if !errors.Is(
		err,
		context.DeadlineExceeded,
	) {
		t.Fatalf(
			"expected deadline exceeded, got %v",
			err,
		)
	}
}

type timeoutContractRoundTripper func(
	request *http.Request,
) (*http.Response, error)

func (
	roundTripper timeoutContractRoundTripper,
) RoundTrip(
	request *http.Request,
) (*http.Response, error) {
	return roundTripper(
		request,
	)
}
