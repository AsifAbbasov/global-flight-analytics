package ourairports

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientConditionalRequestReturnsNotModified(
	t *testing.T,
) {
	fixedTime := time.Date(
		2026,
		time.July,
		5,
		8,
		0,
		0,
		0,
		time.UTC,
	)

	const requestETag = `"test-etag"`
	const requestLastModified = "Sun, 05 Jul 2026 01:53:55 GMT"
	const responseETag = `"test-etag"`

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				request *http.Request,
			) {
				if request.Header.Get(
					"If-None-Match",
				) != requestETag {
					t.Fatalf(
						"unexpected If-None-Match header: %q",
						request.Header.Get(
							"If-None-Match",
						),
					)
				}

				if request.Header.Get(
					"If-Modified-Since",
				) != requestLastModified {
					t.Fatalf(
						"unexpected If-Modified-Since header: %q",
						request.Header.Get(
							"If-Modified-Since",
						),
					)
				}

				writer.Header().Set(
					"ETag",
					responseETag,
				)

				writer.WriteHeader(
					http.StatusNotModified,
				)
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

	result, err := client.LoadAirportsConditional(
		context.Background(),
		ConditionalRequest{
			ETag:         requestETag,
			LastModified: requestLastModified,
		},
	)
	if err != nil {
		t.Fatalf(
			"load airports conditionally: %v",
			err,
		)
	}

	if !result.NotModified {
		t.Fatal(
			"expected not-modified result",
		)
	}

	if len(result.Airports) != 0 {
		t.Fatalf(
			"expected no parsed airports, got %d",
			len(result.Airports),
		)
	}

	if result.ETag != responseETag {
		t.Fatalf(
			"expected ETag %q, got %q",
			responseETag,
			result.ETag,
		)
	}

	if result.LastModified != requestLastModified {
		t.Fatalf(
			"expected preserved Last-Modified %q, got %q",
			requestLastModified,
			result.LastModified,
		)
	}

	if !result.CheckedAt.Equal(
		fixedTime,
	) {
		t.Fatalf(
			"expected checked time %s, got %s",
			fixedTime,
			result.CheckedAt,
		)
	}

	if !result.RetrievedAt.IsZero() {
		t.Fatalf(
			"expected zero retrieved time for 304, got %s",
			result.RetrievedAt,
		)
	}
}

func TestClientSuccessfulResponseCapturesValidators(
	t *testing.T,
) {
	const csvData = `ident,name,latitude_deg,longitude_deg,elevation_ft,iso_country,municipality,icao_code,iata_code
UBBB,Heydar Aliyev International Airport,40.4675,50.0467,10,AZ,Baku,UBBB,GYD
`

	fixedTime := time.Date(
		2026,
		time.July,
		5,
		8,
		0,
		0,
		0,
		time.UTC,
	)

	const responseETag = `"downloaded-etag"`
	const responseLastModified = "Sun, 05 Jul 2026 01:53:55 GMT"

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				writer http.ResponseWriter,
				_ *http.Request,
			) {
				writer.Header().Set(
					"ETag",
					responseETag,
				)

				writer.Header().Set(
					"Last-Modified",
					responseLastModified,
				)

				writer.WriteHeader(
					http.StatusOK,
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

	if result.NotModified {
		t.Fatal(
			"expected downloaded response",
		)
	}

	if result.ETag != responseETag {
		t.Fatalf(
			"expected ETag %q, got %q",
			responseETag,
			result.ETag,
		)
	}

	if result.LastModified != responseLastModified {
		t.Fatalf(
			"expected Last-Modified %q, got %q",
			responseLastModified,
			result.LastModified,
		)
	}

	if len(result.Airports) != 1 {
		t.Fatalf(
			"expected one parsed airport, got %d",
			len(result.Airports),
		)
	}

	if !result.RetrievedAt.Equal(
		fixedTime,
	) {
		t.Fatalf(
			"expected retrieved time %s, got %s",
			fixedTime,
			result.RetrievedAt,
		)
	}
}
