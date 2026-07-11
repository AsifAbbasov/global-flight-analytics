package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
)

type cycleTestTrafficProvider struct {
	states []flightstate.FlightState
	err    error
	calls  int
}

func (
	provider *cycleTestTrafficProvider,
) SourceName() string {
	return "airplanes.live"
}

func (
	provider *cycleTestTrafficProvider,
) LoadByPoint(
	context.Context,
	float64,
	float64,
	int,
) ([]flightstate.FlightState, error) {
	provider.calls++

	states := make(
		[]flightstate.FlightState,
		len(provider.states),
	)
	copy(
		states,
		provider.states,
	)

	return states, provider.err
}

type cycleTestProcessingService struct {
	received []flightstate.FlightState
	result   trafficapplication.ProcessAndStoreResult
	err      error
}

func (
	service *cycleTestProcessingService,
) ProcessAndStore(
	_ context.Context,
	states []flightstate.FlightState,
) (trafficapplication.ProcessAndStoreResult, error) {
	service.received = append(
		[]flightstate.FlightState(nil),
		states...,
	)

	return service.result, service.err
}

type cycleTestRunRepository struct {
	run ingestionrun.Run

	markSuccessCalled bool
	markFailedCalled  bool
}

func (
	repository *cycleTestRunRepository,
) CreateRunning(
	context.Context,
	string,
	string,
	time.Time,
) (ingestionrun.Run, error) {
	return repository.run, nil
}

func (
	repository *cycleTestRunRepository,
) MarkSuccess(
	context.Context,
	string,
	time.Time,
	int,
	int,
	int,
) error {
	repository.markSuccessCalled = true

	return nil
}

func (
	repository *cycleTestRunRepository,
) MarkFailed(
	context.Context,
	string,
	time.Time,
	int,
	int,
	int,
	string,
) error {
	repository.markFailedCalled = true

	return nil
}

func TestIngestionCycleRunsOrchestratedProviderDirectly(
	t *testing.T,
) {
	provider := &cycleTestTrafficProvider{
		states: []flightstate.FlightState{
			{
				ICAO24: "4k001",
			},
		},
	}

	processingService := &cycleTestProcessingService{
		result: trafficapplication.ProcessAndStoreResult{
			StoredFlightStateCount: 1,
			StoredAt: time.Date(
				2026,
				time.July,
				11,
				21,
				0,
				0,
				0,
				time.UTC,
			),
		},
	}

	runRepository := &cycleTestRunRepository{
		run: ingestionrun.Run{
			ID: "run-1",
		},
	}

	cycle, err := newIngestionCycle(
		ingestionCycleConfig{
			TrafficProvider:        provider,
			ProcessingService:      processingService,
			IngestionRunRepository: runRepository,
			Latitude:               40.4093,
			Longitude:              49.8671,
			Radius:                 100,
		},
	)
	if err != nil {
		t.Fatalf(
			"create ingestion cycle: %v",
			err,
		)
	}

	if err := cycle.Run(
		context.Background(),
	); err != nil {
		t.Fatalf(
			"run ingestion cycle: %v",
			err,
		)
	}

	if provider.calls != 1 {
		t.Fatalf(
			"expected one provider call, got %d",
			provider.calls,
		)
	}

	if len(processingService.received) != 1 {
		t.Fatalf(
			"expected one processed state, got %d",
			len(processingService.received),
		)
	}

	if processingService.received[0].IngestionRunID != "run-1" {
		t.Fatalf(
			"expected ingestion run id run-1, got %q",
			processingService.received[0].IngestionRunID,
		)
	}

	if !runRepository.markSuccessCalled {
		t.Fatal(
			"expected ingestion run to be marked successful",
		)
	}

	if runRepository.markFailedCalled {
		t.Fatal(
			"did not expect ingestion run to be marked failed",
		)
	}
}

func TestIngestionCycleReturnsProviderFailure(
	t *testing.T,
) {
	providerFailure := errors.New(
		"provider unavailable",
	)

	provider := &cycleTestTrafficProvider{
		err: providerFailure,
	}

	processingService := &cycleTestProcessingService{}
	runRepository := &cycleTestRunRepository{
		run: ingestionrun.Run{
			ID: "run-2",
		},
	}

	cycle, err := newIngestionCycle(
		ingestionCycleConfig{
			TrafficProvider:        provider,
			ProcessingService:      processingService,
			IngestionRunRepository: runRepository,
		},
	)
	if err != nil {
		t.Fatalf(
			"create ingestion cycle: %v",
			err,
		)
	}

	err = cycle.Run(
		context.Background(),
	)
	if err == nil {
		t.Fatal(
			"expected ingestion cycle failure",
		)
	}

	if !errors.Is(
		err,
		providerFailure,
	) {
		t.Fatalf(
			"expected provider failure, got %v",
			err,
		)
	}

	if !runRepository.markFailedCalled {
		t.Fatal(
			"expected ingestion run to be marked failed",
		)
	}

	if runRepository.markSuccessCalled {
		t.Fatal(
			"did not expect ingestion run to be marked successful",
		)
	}
}

func TestNewIngestionCycleValidatesDependencies(
	t *testing.T,
) {
	provider := &cycleTestTrafficProvider{}
	processingService := &cycleTestProcessingService{}
	runRepository := &cycleTestRunRepository{}

	tests := []struct {
		name        string
		config      ingestionCycleConfig
		expectedErr error
	}{
		{
			name: "traffic provider required",
			config: ingestionCycleConfig{
				ProcessingService:      processingService,
				IngestionRunRepository: runRepository,
			},
			expectedErr: errIngestionCycleTrafficProviderRequired,
		},
		{
			name: "processing service required",
			config: ingestionCycleConfig{
				TrafficProvider:        provider,
				IngestionRunRepository: runRepository,
			},
			expectedErr: errIngestionCycleProcessingServiceRequired,
		},
		{
			name: "run repository required",
			config: ingestionCycleConfig{
				TrafficProvider:   provider,
				ProcessingService: processingService,
			},
			expectedErr: errIngestionCycleRunRepositoryRequired,
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(
				t *testing.T,
			) {
				cycle, err := newIngestionCycle(
					test.config,
				)

				if cycle != nil {
					t.Fatal(
						"expected nil ingestion cycle",
					)
				}

				if !errors.Is(
					err,
					test.expectedErr,
				) {
					t.Fatalf(
						"expected %v, got %v",
						test.expectedErr,
						err,
					)
				}
			},
		)
	}
}
