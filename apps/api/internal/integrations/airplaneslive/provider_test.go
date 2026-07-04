package airplaneslive

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

func TestLoadByPointReturnsCanonicalFlightStates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)

			_, _ = writer.Write([]byte(`{
				"now": 1760000000,
				"messages": 1,
				"total": 1,
				"aircraft": [
					{
						"hex": "abc123",
						"flight": " AHY101 ",
						"lat": 40.4093,
						"lon": 49.8671,
						"alt_baro": 32000,
						"alt_geom": 32500,
						"gs": 450,
						"track": 92,
						"baro_rate": 500,
						"seen": 5,
						"type": "adsb_icao",
						"r": "4K-AZ01",
						"t": "A320"
					}
				]
			}`))
		},
	))
	defer server.Close()

	client := NewClient(integrationcommon.HTTPClientConfig{
		BaseURL:   server.URL,
		Timeout:   time.Second,
		UserAgent: "global-flight-analytics-test",
	})

	provider := NewProvider(client)

	states, err := provider.LoadByPoint(
		context.Background(),
		40.4093,
		49.8671,
		250,
	)
	if err != nil {
		t.Fatalf(
			"expected successful provider point load, got error: %v",
			err,
		)
	}

	if len(states) != 1 {
		t.Fatalf(
			"expected 1 canonical flight state, got %d",
			len(states),
		)
	}

	state := states[0]

	if state.ICAO24 != "ABC123" {
		t.Fatalf(
			"expected normalized ICAO24 ABC123, got %q",
			state.ICAO24,
		)
	}

	if state.Callsign != "AHY101" {
		t.Fatalf(
			"expected trimmed callsign AHY101, got %q",
			state.Callsign,
		)
	}

	if state.Latitude != 40.4093 {
		t.Fatalf(
			"expected latitude 40.4093, got %f",
			state.Latitude,
		)
	}

	if state.Longitude != 49.8671 {
		t.Fatalf(
			"expected longitude 49.8671, got %f",
			state.Longitude,
		)
	}

	if state.HeadingDegrees != 92 {
		t.Fatalf(
			"expected heading 92 degrees, got %f",
			state.HeadingDegrees,
		)
	}

	if state.OnGround {
		t.Fatal("expected aircraft state to be airborne")
	}

	expectedObservedAt := time.Unix(
		1759999995,
		0,
	).UTC()

	if !state.ObservedAt.Equal(expectedObservedAt) {
		t.Fatalf(
			"expected observed_at %s, got %s",
			expectedObservedAt,
			state.ObservedAt,
		)
	}

	if state.SourceName != "airplanes.live" {
		t.Fatalf(
			"expected source name airplanes.live, got %q",
			state.SourceName,
		)
	}
}

func TestLoadByPointWrapsClientValidationError(t *testing.T) {
	client := NewClient(integrationcommon.HTTPClientConfig{
		BaseURL:   "http://127.0.0.1",
		Timeout:   time.Second,
		UserAgent: "global-flight-analytics-test",
	})

	provider := NewProvider(client)

	states, err := provider.LoadByPoint(
		context.Background(),
		40.4093,
		49.8671,
		0,
	)

	if err == nil {
		t.Fatal("expected provider error for invalid radius")
	}

	if states != nil {
		t.Fatal("expected nil states for invalid radius")
	}

	if !strings.Contains(
		err.Error(),
		"load airplanes live traffic by point",
	) {
		t.Fatalf(
			"expected provider error context, got %q",
			err.Error(),
		)
	}
}
