package airplaneslive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

func TestTypedTransportDecodesStateResponse(
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

				_, _ = writer.Write(
					[]byte(`{
						"now": 1720526400000,
						"messages": 1,
						"total": 1,
						"ac": [
							{
								"hex": "abc123",
								"flight": "AHY101",
								"lat": 40.4093,
								"lon": 49.8671,
								"alt_baro": 1000,
								"alt_geom": 1100,
								"gs": 200,
								"track": 90,
								"baro_rate": 0,
								"seen": 0
							}
						]
					}`),
				)
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
		integrationcommon.HTTPClientConfig{
			BaseURL:   server.URL,
			Timeout:   time.Second,
			UserAgent: "typed-transport-test",
		},
	)
	if err != nil {
		t.Fatalf(
			"create airplanes.live client: %v",
			err,
		)
	}

	response, err := client.GetByPoint(
		context.Background(),
		40.4093,
		49.8671,
		250,
	)
	if err != nil {
		t.Fatalf(
			"get typed state response: %v",
			err,
		)
	}

	if response == nil {
		t.Fatal(
			"expected typed state response",
		)
	}

	if len(response.Aircraft) != 1 {
		t.Fatalf(
			"expected 1 aircraft, got %d",
			len(response.Aircraft),
		)
	}

	if response.Aircraft[0].Hex != "abc123" {
		t.Fatalf(
			"expected aircraft hex abc123, got %s",
			response.Aircraft[0].Hex,
		)
	}

	if response.Aircraft[0].AltBaro.Kind !=
		BarometricAltitudeKindObserved {
		t.Fatalf(
			"expected observed barometric altitude kind, got %q",
			response.Aircraft[0].AltBaro.Kind,
		)
	}
}
