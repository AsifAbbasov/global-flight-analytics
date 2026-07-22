package ourairports

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

func TestAirportsCSVRejectsDeclaredOversize(
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
					strconv.FormatInt(maxAirportsCSVResponseBytes+1, 10),
				)
				writer.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
		ClientConfig{
			Timeout:        time.Second,
			AirportsCSVURL: server.URL,
		},
	)
	if err != nil {
		t.Fatalf("create OurAirports client: %v", err)
	}

	_, err = client.LoadAirports(context.Background())
	if !errors.Is(err, integrationcommon.ErrProviderResponseTooLarge) {
		t.Fatalf("expected oversized response error, got %v", err)
	}
}
