package ingestion

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
)

func TestLoadAndProcessByPoint(t *testing.T) {
	provider := &testRegionalProvider{
		states: []flightstate.FlightState{
			{ICAO24: "ABC123"},
			{ICAO24: "DEF456"},
		},
	}

	processor := &testProcessingService{}

	service := New(Config{
		Provider:          provider,
		ProcessingService: processor,
	})

	result, err := service.LoadAndProcessByPoint(
		context.Background(),
		40.4093,
		49.8671,
		250,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.LoadedStateCount != 2 {
		t.Fatalf(
			"expected 2 loaded states, got %d",
			result.LoadedStateCount,
		)
	}

	if provider.callCount != 1 {
		t.Fatalf(
			"expected provider call count 1, got %d",
			provider.callCount,
		)
	}

	if processor.callCount != 1 {
		t.Fatalf(
			"expected processor call count 1, got %d",
			processor.callCount,
		)
	}

	if len(processor.lastStates) != 2 {
		t.Fatalf(
			"expected processor to receive 2 states, got %d",
			len(processor.lastStates),
		)
	}
}

func TestLoadAndProcessByPointWrapsProviderError(t *testing.T) {
	service := New(Config{
		Provider: &testRegionalProvider{
			err: errors.New("provider failed"),
		},
		ProcessingService: &testProcessingService{},
	})

	_, err := service.LoadAndProcessByPoint(
		context.Background(),
		40.4093,
		49.8671,
		250,
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "load regional flight states") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type testRegionalProvider struct {
	states    []flightstate.FlightState
	err       error
	callCount int
}

func (provider *testRegionalProvider) LoadByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) ([]flightstate.FlightState, error) {
	provider.callCount++

	if provider.err != nil {
		return nil, provider.err
	}

	return provider.states, nil
}

type testProcessingService struct {
	callCount  int
	lastStates []flightstate.FlightState
}

func (service *testProcessingService) ProcessAndStore(
	ctx context.Context,
	states []flightstate.FlightState,
) (trafficapplication.ProcessAndStoreResult, error) {
	service.callCount++
	service.lastStates = states

	return trafficapplication.ProcessAndStoreResult{}, nil
}
