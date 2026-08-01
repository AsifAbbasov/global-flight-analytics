package opensky

import (
	"net/http"
	"testing"
	"time"
)

func TestNewClientAppliesDefaultTimeoutToUnboundedHTTPClientClone(
	t *testing.T,
) {
	originalClient := &http.Client{}
	config := DefaultConfig()
	config.HTTPClient = originalClient

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("create OpenSky client: %v", err)
	}
	if client.config.HTTPClient == originalClient {
		t.Fatal("expected OpenSky client to own an HTTP client clone")
	}
	if originalClient.Timeout != 0 {
		t.Fatalf("constructor mutated caller timeout to %s", originalClient.Timeout)
	}
	if client.config.HTTPClient.Timeout != DefaultHTTPTimeout {
		t.Fatalf(
			"expected default timeout %s, got %s",
			DefaultHTTPTimeout,
			client.config.HTTPClient.Timeout,
		)
	}
}

func TestNewClientClonesBoundedHTTPClient(
	t *testing.T,
) {
	const timeout = 7 * time.Second
	originalClient := &http.Client{Timeout: timeout}
	config := DefaultConfig()
	config.HTTPClient = originalClient

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("create OpenSky client: %v", err)
	}
	if client.config.HTTPClient == originalClient {
		t.Fatal("expected OpenSky client to own an HTTP client clone")
	}
	if client.config.HTTPClient.Timeout != timeout {
		t.Fatalf("expected timeout %s, got %s", timeout, client.config.HTTPClient.Timeout)
	}

	originalClient.Timeout = 0
	if client.config.HTTPClient.Timeout != timeout {
		t.Fatalf(
			"caller mutation changed validated timeout to %s",
			client.config.HTTPClient.Timeout,
		)
	}
}
