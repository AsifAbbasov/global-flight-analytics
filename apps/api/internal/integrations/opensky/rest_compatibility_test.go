package opensky

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseStateVectorAcceptsBaseSeventeenFieldResponse(t *testing.T) {
	raw := json.RawMessage(`[
		"abc123", "AHY101", "Azerbaijan", 1760000000, 1760000005,
		49.8671, 40.4093, 9753.6, false, 231.5, 92.0, 2.5,
		null, 9906.0, "7700", false, 0
	]`)

	state, err := ParseStateVector(raw)
	if err != nil {
		t.Fatalf("parse base state vector: %v", err)
	}
	if state.CategoryAvailable {
		t.Fatal("base state vector must not claim that aircraft category was returned")
	}
	if state.Category != AircraftCategoryNoInformation {
		t.Fatalf("category = %d, want no information", state.Category)
	}
}

func TestParseStateVectorPreservesExtendedCategory(t *testing.T) {
	raw := json.RawMessage(`[
		"abc123", "AHY101", "Azerbaijan", 1760000000, 1760000005,
		49.8671, 40.4093, 9753.6, false, 231.5, 92.0, 2.5,
		null, 9906.0, "7700", false, 0, 6
	]`)

	state, err := ParseStateVector(raw)
	if err != nil {
		t.Fatalf("parse extended state vector: %v", err)
	}
	if !state.CategoryAvailable {
		t.Fatal("extended state vector must record that aircraft category was returned")
	}
	if state.Category != AircraftCategoryHeavy {
		t.Fatalf("category = %d, want heavy", state.Category)
	}
}

func TestStatesRequestAddsExtendedCategoryParameter(t *testing.T) {
	var extended string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		extended = request.URL.Query().Get("extended")
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"time":1760000005,"states":[]}`))
	}))
	defer server.Close()

	config := DefaultConfig()
	config.BaseURL = server.URL
	config.HTTPClient = server.Client()
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.GetStates(context.Background(), StatesRequest{Extended: true})
	if err != nil {
		t.Fatalf("get extended states: %v", err)
	}
	if extended != "1" {
		t.Fatalf("extended query = %q, want 1", extended)
	}
}

func TestProductionProviderRequestsExtendedCategory(t *testing.T) {
	client := &capturingStatesClient{}
	provider, err := NewProvider(client)
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	_, err = provider.LoadByPoint(context.Background(), 40.4093, 49.8671, 100)
	if err != nil {
		t.Fatalf("load by point: %v", err)
	}
	if !client.request.Extended {
		t.Fatal("production OpenSky provider must request extended aircraft category")
	}
}

type capturingStatesClient struct {
	request StatesRequest
}

func (client *capturingStatesClient) GetStates(
	_ context.Context,
	request StatesRequest,
) (StatesResult, error) {
	client.request = request
	return StatesResult{States: []StateVector{}}, nil
}
