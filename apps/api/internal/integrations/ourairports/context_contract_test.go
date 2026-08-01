package ourairports

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

type contextContractRoundTripper struct {
	calls int
}

func (transport *contextContractRoundTripper) RoundTrip(
	*http.Request,
) (*http.Response, error) {
	transport.calls++

	return nil, errors.New("unexpected OurAirports HTTP request")
}

func TestClientRejectsNilContextBeforeHTTPRequests(
	t *testing.T,
) {
	client, err := NewClient(ClientConfig{
		Timeout:        time.Second,
		AirportsCSVURL: "https://example.invalid/airports.csv",
	})
	if err != nil {
		t.Fatalf(
			"create client: %v",
			err,
		)
	}

	transport := &contextContractRoundTripper{}
	client.httpClient.Transport = transport

	_, err = client.LoadAirports(nil)
	if !errors.Is(
		err,
		ErrLoadContextRequired,
	) {
		t.Fatalf(
			"LoadAirports error = %v, want load context required",
			err,
		)
	}

	_, err = client.LoadAirportsConditional(
		nil,
		ConditionalRequest{},
	)
	if !errors.Is(
		err,
		ErrLoadContextRequired,
	) {
		t.Fatalf(
			"LoadAirportsConditional error = %v, want load context required",
			err,
		)
	}

	if transport.calls != 0 {
		t.Fatalf(
			"HTTP calls = %d, want 0",
			transport.calls,
		)
	}
}
