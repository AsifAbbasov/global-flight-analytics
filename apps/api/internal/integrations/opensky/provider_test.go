package opensky

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

type statesClientStub struct {
	result  StatesResult
	err     error
	request StatesRequest
}

func (stub *statesClientStub) GetStates(
	_ context.Context,
	request StatesRequest,
) (StatesResult, error) {
	stub.request = request
	return stub.result, stub.err
}

func TestRegionalBoundingBoxUsesSharedNauticalMileRadius(
	t *testing.T,
) {
	box, err := RegionalBoundingBox(40.4093, 49.8671, 100)
	if err != nil {
		t.Fatalf("build regional bounding box: %v", err)
	}
	if box.MinimumLatitude >= 40.4093 ||
		box.MaximumLatitude <= 40.4093 ||
		box.MinimumLongitude >= 49.8671 ||
		box.MaximumLongitude <= 49.8671 {
		t.Fatalf("bounding box does not contain the requested point: %+v", box)
	}
	cost, err := box.EstimatedStateCreditCost()
	if err != nil {
		t.Fatalf("estimate state credit cost: %v", err)
	}
	if cost > maximumRegionalStateCreditCost {
		t.Fatalf("credit cost = %d, maximum = %d", cost, maximumRegionalStateCreditCost)
	}
}

func TestRegionalBoundingBoxRejectsRadiusAboveFreeRegionalBoundary(
	t *testing.T,
) {
	_, err := RegionalBoundingBox(40.4093, 49.8671, 251)
	if !errors.Is(err, ErrRegionalRadiusInvalid) {
		t.Fatalf("expected regional radius error, got %v", err)
	}
}

func TestMapStateVectorAcceptsProviderValidPosition(
	t *testing.T,
) {
	snapshot := time.Date(2026, time.July, 18, 10, 0, 15, 0, time.UTC)
	positionTime := snapshot.Add(-5 * time.Second)
	latitude := 40.4093
	longitude := 49.8671
	barometricAltitude := 10000.0
	geometricAltitude := 10200.0
	velocity := 230.0
	track := 92.0
	verticalRate := 1.5
	callsign := " AHY101 "

	mapped, usable, err := MapStateVector(StateVector{
		SnapshotTime:    snapshot,
		ICAO24:          "abc123",
		Callsign:        &callsign,
		OriginCountry:   "Azerbaijan",
		TimePosition:    &positionTime,
		LastContact:     snapshot.Add(-time.Second),
		Longitude:       &longitude,
		Latitude:        &latitude,
		BaroAltitudeM:   &barometricAltitude,
		VelocityMPS:     &velocity,
		TrueTrack:       &track,
		VerticalRateMPS: &verticalRate,
		GeoAltitudeM:    &geometricAltitude,
	})
	if err != nil {
		t.Fatalf("map state vector: %v", err)
	}
	if !usable {
		t.Fatal("expected provider-valid position to be usable")
	}
	if mapped.ICAO24 != "ABC123" || mapped.Callsign != "AHY101" {
		t.Fatalf("unexpected identity mapping: %+v", mapped)
	}
	if mapped.SourceName != sourceName {
		t.Fatalf("source name = %q, want %q", mapped.SourceName, sourceName)
	}
	if !mapped.ObservedAt.Equal(positionTime) {
		t.Fatalf("observed at = %s, want %s", mapped.ObservedAt, positionTime)
	}
	if mapped.BarometricAltitudeM != barometricAltitude ||
		mapped.GeometricAltitudeM != geometricAltitude {
		t.Fatalf("altitude mapping was not preserved: %+v", mapped)
	}
}

func TestMapStateVectorRejectsStalePosition(
	t *testing.T,
) {
	snapshot := time.Date(2026, time.July, 18, 10, 0, 30, 0, time.UTC)
	positionTime := snapshot.Add(-16 * time.Second)
	latitude := 40.4093
	longitude := 49.8671

	_, usable, err := MapStateVector(StateVector{
		SnapshotTime: snapshot,
		ICAO24:       "abc123",
		TimePosition: &positionTime,
		LastContact:  snapshot.Add(-time.Second),
		Longitude:    &longitude,
		Latitude:     &latitude,
	})
	if err != nil {
		t.Fatalf("map stale state vector: %v", err)
	}
	if usable {
		t.Fatal("expected stale OpenSky position to be blocked")
	}
}

func TestProviderLoadsOnlyUsableRegionalStates(
	t *testing.T,
) {
	snapshot := time.Date(2026, time.July, 18, 10, 0, 15, 0, time.UTC)
	freshPosition := snapshot.Add(-5 * time.Second)
	stalePosition := snapshot.Add(-16 * time.Second)
	latitude := 40.4093
	longitude := 49.8671

	client := &statesClientStub{
		result: StatesResult{
			States: []StateVector{
				{
					SnapshotTime: snapshot,
					ICAO24:       "abc123",
					TimePosition: &freshPosition,
					LastContact:  snapshot.Add(-time.Second),
					Longitude:    &longitude,
					Latitude:     &latitude,
				},
				{
					SnapshotTime: snapshot,
					ICAO24:       "def456",
					TimePosition: &stalePosition,
					LastContact:  snapshot.Add(-time.Second),
					Longitude:    &longitude,
					Latitude:     &latitude,
				},
			},
		},
	}
	provider, err := NewProvider(client)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	states, err := provider.LoadByPoint(context.Background(), 40.4093, 49.8671, 100)
	if err != nil {
		t.Fatalf("load regional states: %v", err)
	}
	if len(states) != 1 || states[0].ICAO24 != "ABC123" {
		t.Fatalf("unexpected mapped states: %+v", states)
	}
	if client.request.BoundingBox == nil {
		t.Fatal("expected bounded OpenSky states request")
	}
	area, err := client.request.BoundingBox.AreaSquareDegrees()
	if err != nil {
		t.Fatalf("calculate request area: %v", err)
	}
	if math.IsNaN(area) || math.IsInf(area, 0) || area <= 0 {
		t.Fatalf("invalid request area: %f", area)
	}
}

type providerResponseObserverStub struct {
	providerName string
	statusCode   int
	remaining    string
	calls        int
}

func (stub *providerResponseObserverStub) ObserveProviderResponse(
	providerName string,
	statusCode int,
	headers http.Header,
	_ time.Duration,
) error {
	stub.providerName = providerName
	stub.statusCode = statusCode
	stub.remaining = headers.Get("X-Rate-Limit-Remaining")
	stub.calls++
	return nil
}

func TestClientReportsOpenSkyHeadersAndTypedRateLimitError(
	t *testing.T,
) {
	server := httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("X-Rate-Limit-Remaining", "0")
			writer.Header().Set("X-Rate-Limit-Retry-After-Seconds", "30")
			writer.WriteHeader(http.StatusTooManyRequests)
		},
	))
	defer server.Close()

	observer := &providerResponseObserverStub{}
	clientConfig := DefaultConfig()
	clientConfig.BaseURL = server.URL
	clientConfig.HTTPClient = server.Client()
	clientConfig.PollingInterval = 10 * time.Second
	client, err := NewClientWithResponseObserver(clientConfig, observer)
	if err != nil {
		t.Fatalf("create observed OpenSky client: %v", err)
	}

	_, err = client.GetStates(context.Background(), StatesRequest{})
	if !errors.Is(err, integrationcommon.ErrProviderRateLimited) {
		t.Fatalf("expected typed rate-limit error, got %v", err)
	}
	if observer.calls != 1 ||
		observer.providerName != sourceName ||
		observer.statusCode != http.StatusTooManyRequests ||
		observer.remaining != "0" {
		t.Fatalf("unexpected provider observation: %+v", observer)
	}
}
