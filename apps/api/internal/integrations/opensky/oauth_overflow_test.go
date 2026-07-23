package opensky

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOAuthTokenResponseRejectsExpiresInDurationOverflow(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(writer http.ResponseWriter, _ *http.Request) {
				writer.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintf(
					writer,
					`{"access_token":"token","expires_in":%d,"token_type":"Bearer"}`,
					maxTokenLifetimeSeconds+1,
				)
			},
		),
	)
	defer server.Close()

	manager, err := NewTokenManager(
		server.Client(),
		server.URL,
		"client-id",
		"client-secret",
	)
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}

	_, err = manager.Token(context.Background())
	if !errors.Is(err, ErrTokenResponseInvalid) {
		t.Fatalf("error = %v, want %v", err, ErrTokenResponseInvalid)
	}
}
