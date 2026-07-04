package ingestion

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
)

func TestLoadAndProcessByPoint(t *testing.T) {
	provider := &testRegionalProvider{
		sourceName: "test-provider",
		states: []flightstate.FlightState{
			{ICAO24: "ABC123"},
			{ICAO24: "DEF456"},
		},
	}

	processor := &testProcessingService{}

	runRepository := &testIngestionRunRepository{
		run: ingestionrun.Run{
			ID: "run-1",
		},
	}

	service := New(Config{
		Provider:               provider,
		ProcessingService:      processor,
		IngestionRunRepository: runRepository,
		Now:                    fixedIngestionTime,
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

	if result.IngestionRunID != "run-1" {
		t.Fatalf(
			"expected ingestion run id run-1, got %s",
			result.IngestionRunID,
		)
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

	for _, state := range processor.lastStates {
		if state.IngestionRunID != "run-1" {
			t.Fatalf(
				"expected ingestion run id run-1 for state %s, got %s",
				state.ICAO24,
				state.IngestionRunID,
			)
		}
	}

	if runRepository.createCount != 1 {
		t.Fatalf(
			"expected ingestion run create count 1, got %d",
			runRepository.createCount,
		)
	}

	if runRepository.successCount != 1 {
		t.Fatalf(
			"expected ingestion run success count 1, got %d",
			runRepository.successCount,
		)
	}

	if runRepository.failedCount != 0 {
		t.Fatalf(
			"expected ingestion run failed count 0, got %d",
			runRepository.failedCount,
		)
	}

	if runRepository.lastRecordsReceived != 2 {
		t.Fatalf(
			"expected 2 received records, got %d",
			runRepository.lastRecordsReceived,
		)
	}

	if runRepository.lastRecordsInserted != 2 {
		t.Fatalf(
			"expected 2 inserted records, got %d",
			runRepository.lastRecordsInserted,
		)
	}
}

func TestLoadAndProcessByPointWrapsProviderError(t *testing.T) {
	runRepository := &testIngestionRunRepository{
		run: ingestionrun.Run{
			ID: "run-1",
		},
	}

	service := New(Config{
		Provider: &testRegionalProvider{
			sourceName: "test-provider",
			err:        errors.New("provider failed"),
		},
		ProcessingService:      &testProcessingService{},
		IngestionRunRepository: runRepository,
		Now:                    fixedIngestionTime,
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

	if runRepository.createCount != 1 {
		t.Fatalf(
			"expected ingestion run create count 1, got %d",
			runRepository.createCount,
		)
	}

	if runRepository.failedCount != 1 {
		t.Fatalf(
			"expected ingestion run failed count 1, got %d",
			runRepository.failedCount,
		)
	}

	if runRepository.successCount != 0 {
		t.Fatalf(
			"expected ingestion run success count 0, got %d",
			runRepository.successCount,
		)
	}
}

func fixedIngestionTime() time.Time {
	return time.Date(
		2026,
		time.July,
		4,
		12,
		0,
		0,
		0,
		time.UTC,
	)
}

type testRegionalProvider struct {
	sourceName string
	states     []flightstate.FlightState
	err        error
	callCount  int
}

func (provider *testRegionalProvider) SourceName() string {
	return provider.sourceName
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

	service.lastStates = append(
		[]flightstate.FlightState(nil),
		states...,
	)

	return trafficapplication.ProcessAndStoreResult{
		StoredFlightStateCount: len(states),
	}, nil
}

type testIngestionRunRepository struct {
	run                 ingestionrun.Run
	createCount         int
	successCount        int
	failedCount         int
	lastRecordsReceived int
	lastRecordsInserted int
	lastRecordsUpdated  int
}

func (repository *testIngestionRunRepository) CreateRunning(
	ctx context.Context,
	sourceName string,
	regionID string,
	startedAt time.Time,
) (ingestionrun.Run, error) {
	repository.createCount++

	return repository.run, nil
}

func (repository *testIngestionRunRepository) MarkSuccess(
	ctx context.Context,
	id string,
	finishedAt time.Time,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
) error {
	repository.successCount++
	repository.lastRecordsReceived = recordsReceived
	repository.lastRecordsInserted = recordsInserted
	repository.lastRecordsUpdated = recordsUpdated

	return nil
}

func (repository *testIngestionRunRepository) MarkFailed(
	ctx context.Context,
	id string,
	finishedAt time.Time,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
	errorMessage string,
) error {
	repository.failedCount++
	repository.lastRecordsReceived = recordsReceived
	repository.lastRecordsInserted = recordsInserted
	repository.lastRecordsUpdated = recordsUpdated

	return nil
}
