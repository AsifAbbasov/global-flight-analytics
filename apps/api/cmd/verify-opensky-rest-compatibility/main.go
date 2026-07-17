package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/opensky"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
	fmt.Println("OpenSky REST compatibility verification: PASS")
}

func run() error {
	base := json.RawMessage(`[
		"abc123", "AHY101", "Azerbaijan", 1760000000, 1760000005,
		49.8671, 40.4093, 9753.6, false, 231.5, 92.0, 2.5,
		null, 9906.0, "7700", false, 0
	]`)
	baseState, err := opensky.ParseStateVector(base)
	if err != nil {
		return fmt.Errorf("parse base state vector: %w", err)
	}
	if baseState.CategoryAvailable || baseState.Category != opensky.AircraftCategoryNoInformation {
		return fmt.Errorf("base state vector category contract is invalid")
	}

	extended := json.RawMessage(`[
		"abc123", "AHY101", "Azerbaijan", 1760000000, 1760000005,
		49.8671, 40.4093, 9753.6, false, 231.5, 92.0, 2.5,
		null, 9906.0, "7700", false, 0, 6
	]`)
	extendedState, err := opensky.ParseStateVector(extended)
	if err != nil {
		return fmt.Errorf("parse extended state vector: %w", err)
	}
	if !extendedState.CategoryAvailable || extendedState.Category != opensky.AircraftCategoryHeavy {
		return fmt.Errorf("extended state vector category contract is invalid")
	}

	var requestedExtended string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestedExtended = request.URL.Query().Get("extended")
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"time":1760000005,"states":[]}`))
	}))
	defer server.Close()

	config := opensky.DefaultConfig()
	config.BaseURL = server.URL
	config.HTTPClient = server.Client()
	client, err := opensky.NewClient(config)
	if err != nil {
		return fmt.Errorf("create OpenSky client: %w", err)
	}
	if _, err := client.GetStates(context.Background(), opensky.StatesRequest{Extended: true}); err != nil {
		return fmt.Errorf("request extended states: %w", err)
	}
	if requestedExtended != "1" {
		return fmt.Errorf("extended query parameter = %q, want 1", requestedExtended)
	}

	return nil
}
