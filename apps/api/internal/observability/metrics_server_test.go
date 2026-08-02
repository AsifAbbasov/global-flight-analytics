package observability

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
)

func TestMetricsServerServesAuthorizedPrometheusOutput(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server, err := StartMetricsServer(
		ctx,
		MetricsServerConfig{
			Address:  "127.0.0.1:0",
			Registry: NewRegistry(BuildInfo{}),
			Authorization: AuthorizationConfig{
				ExpectedDigest: internalapikey.DigestCandidate(testMetricsKey),
				Configured:     true,
			},
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
	)
	if err != nil {
		t.Fatalf("start metrics server: %v", err)
	}
	defer func() {
		shutdownContext, shutdownCancel := context.WithTimeout(
			context.Background(),
			time.Second,
		)
		defer shutdownCancel()
		if closeErr := server.Close(shutdownContext); closeErr != nil {
			t.Fatalf("close metrics server: %v", closeErr)
		}
	}()

	request, err := http.NewRequest(
		http.MethodGet,
		"http://"+server.Address()+"/metrics",
		nil,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set(internalapikey.HeaderName, testMetricsKey)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if !strings.Contains(string(body), "global_flight_analytics_build_info") {
		t.Fatalf("unexpected metrics body: %s", body)
	}
}
