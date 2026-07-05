package ourairports

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientLoadAirports(t *testing.T) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,50.0467,10,AZ,Baku,UBBB,GYD
`

	fixedTime := time.Date(
		2026,
		time.July,
		4,
		14,
		30,
		0,
		0,
		time.UTC,
	)

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				if request.Method != http.MethodGet {
					t.Fatalf(
						"expected GET request, got %s",
						request.Method,
					)
				}

				if request.UserAgent() !=
					"global-flight-analytics-airports-import" {
					t.Fatalf(
						"unexpected User-Agent: %s",
						request.UserAgent(),
					)
				}

				writer.Header().Set(
					"Content-Type",
					"text/csv",
				)

				_, err := writer.Write(
					[]byte(csvData),
				)
				if err != nil {
					t.Fatalf(
						"write test response: %v",
						err,
					)
				}
			},
		),
	)
	defer server.Close()

	client, err := NewClient(
		ClientConfig{
			Timeout:        time.Second,
			AirportsCSVURL: server.URL,
			Now: func() time.Time {
				return fixedTime
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	result, err := client.LoadAirports(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"load airports: %v",
			err,
		)
	}

	if len(result.Airports) != 1 {
		t.Fatalf(
			"expected 1 airport, got %d",
			len(result.Airports),
		)
	}

	if result.Airports[0].SourceIdent != "UBBB" {
		t.Fatalf(
			"expected source ident UBBB, got %s",
			result.Airports[0].SourceIdent,
		)
	}

	if !result.RetrievedAt.Equal(fixedTime) {
		t.Fatalf(
			"expected retrieved time %s, got %s",
			fixedTime,
			result.RetrievedAt,
		)
	}

	if !result.Airports[0].LastSyncedAt.Equal(
		fixedTime,
	) {
		t.Fatalf(
			"expected airport synced time %s, got %s",
			fixedTime,
			result.Airports[0].LastSyncedAt,
		)
	}
}

func TestNewClientRejectsMissingTimeout(
	t *testing.T,
) {
	client, err := NewClient(
		ClientConfig{},
	)

	if client != nil {
		t.Fatal(
			"expected nil client",
		)
	}

	if !errors.Is(
		err,
		ErrClientTimeoutRequired,
	) {
		t.Fatalf(
			"expected timeout error, got %v",
			err,
		)
	}
}

func TestClientLoadAirportsRejectsHTTPFailure(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				writer.WriteHeader(
					http.StatusServiceUnavailable,
				)
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
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	_, err = client.LoadAirports(
		context.Background(),
	)
	if err == nil {
		t.Fatal(
			"expected HTTP status error",
		)
	}
}
