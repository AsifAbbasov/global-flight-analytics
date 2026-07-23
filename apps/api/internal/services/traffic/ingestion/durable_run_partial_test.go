package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
)

// These methods extend the shared test repository after the production port
// gained active-run ownership and explicit partial completion.
func (repository *testIngestionRunRepository) UpdateRunningSource(
	_ context.Context,
	_ string,
	sourceName string,
) error {
	repository.lastSourceName = sourceName
	return nil
}

func (repository *testIngestionRunRepository) DeleteRunning(
	context.Context,
	string,
) error {
	return nil
}

func (repository *testIngestionRunRepository) MarkPartial(
	_ context.Context,
	_ string,
	_ time.Time,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
	_ string,
) error {
	repository.failedCount++
	repository.lastRecordsReceived = recordsReceived
	repository.lastRecordsInserted = recordsInserted
	repository.lastRecordsUpdated = recordsUpdated
	return nil
}

func TestLoadAndProcessCreatesRunBeforeProviderCall(
	t *testing.T,
) {
	repository := &durableRunRepositoryStub{}
	provider := &orderedRegionalProvider{
		sourceName: "airplanes.live",
		states: []flightstate.FlightState{
			{ICAO24: "ABC123"},
		},
		beforeLoad: func() {
			if repository.createCount != 1 {
				t.Fatalf(
					"provider called before durable run creation: create_count=%d",
					repository.createCount,
				)
			}
		},
	}

	service := New(Config{
		Provider:               provider,
		ProcessingService:      &configuredProcessingService{},
		IngestionRunRepository: repository,
		Now:                    fixedIngestionTime,
	})

	result, err := service.LoadAndProcessByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err != nil {
		t.Fatalf("load and process: %v", err)
	}
	if result.IngestionRunID != "run-durable" {
		t.Fatalf(
			"ingestion run id = %q, want run-durable",
			result.IngestionRunID,
		)
	}
	if provider.callCount != 1 || repository.successCount != 1 {
		t.Fatalf(
			"provider_calls=%d success_count=%d, want 1 and 1",
			provider.callCount,
			repository.successCount,
		)
	}
}

func TestLoadAndProcessStopsBeforeProviderWhenRunCreationFails(
	t *testing.T,
) {
	createErr := errors.New("database unavailable")
	repository := &durableRunRepositoryStub{
		createErr: createErr,
	}
	provider := &orderedRegionalProvider{
		sourceName: "airplanes.live",
	}

	service := New(Config{
		Provider:               provider,
		ProcessingService:      &configuredProcessingService{},
		IngestionRunRepository: repository,
		Now:                    fixedIngestionTime,
	})

	_, err := service.LoadAndProcessByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if !errors.Is(err, createErr) {
		t.Fatalf("expected create error, got %v", err)
	}
	if provider.callCount != 0 {
		t.Fatalf(
			"provider calls = %d, want 0",
			provider.callCount,
		)
	}
}

func TestLoadAndProcessMarksPartialAfterObservationPersistence(
	t *testing.T,
) {
	processingErr := errors.New("derived trajectory write failed")
	repository := &durableRunRepositoryStub{}
	service := New(Config{
		Provider: &orderedRegionalProvider{
			sourceName: "airplanes.live",
			states: []flightstate.FlightState{
				{ICAO24: "ABC123"},
				{ICAO24: "DEF456"},
			},
		},
		ProcessingService: &configuredProcessingService{
			result: trafficapplication.ProcessAndStoreResult{
				StoredFlightStateCount: 2,
			},
			err: processingErr,
		},
		IngestionRunRepository: repository,
		Now:                    fixedIngestionTime,
	})

	result, err := service.LoadAndProcessByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if !errors.Is(err, processingErr) {
		t.Fatalf("expected processing error, got %v", err)
	}
	if result.ProcessingResult.StoredFlightStateCount != 2 {
		t.Fatalf(
			"stored count = %d, want 2",
			result.ProcessingResult.StoredFlightStateCount,
		)
	}
	if repository.partialCount != 1 || repository.failedCount != 0 {
		t.Fatalf(
			"partial=%d failed=%d, want 1 and 0",
			repository.partialCount,
			repository.failedCount,
		)
	}
}

func TestLoadAndProcessMarksFailedBeforeObservationPersistence(
	t *testing.T,
) {
	processingErr := errors.New("flight state transaction failed")
	repository := &durableRunRepositoryStub{}
	service := New(Config{
		Provider: &orderedRegionalProvider{
			sourceName: "airplanes.live",
			states: []flightstate.FlightState{
				{ICAO24: "ABC123"},
			},
		},
		ProcessingService: &configuredProcessingService{
			err: processingErr,
		},
		IngestionRunRepository: repository,
		Now:                    fixedIngestionTime,
	})

	_, err := service.LoadAndProcessByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if !errors.Is(err, processingErr) {
		t.Fatalf("expected processing error, got %v", err)
	}
	if repository.partialCount != 0 || repository.failedCount != 1 {
		t.Fatalf(
			"partial=%d failed=%d, want 0 and 1",
			repository.partialCount,
			repository.failedCount,
		)
	}
}

func TestLocalProviderDenialRemovesProvisionalRun(
	t *testing.T,
) {
	repository := &durableRunRepositoryStub{}
	service := New(Config{
		Provider: &orderedRegionalProvider{
			sourceName: "airplanes.live",
			err: &ingestionorchestrator.AccessDeniedError{
				Provider: providerpolicy.ProviderAirplanesLive,
				Reason: providerbudget.
					DecisionReasonFixedWindowExhausted,
			},
		},
		ProcessingService:      &configuredProcessingService{},
		IngestionRunRepository: repository,
		Now:                    fixedIngestionTime,
	})

	result, err := service.LoadAndProcessByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err == nil {
		t.Fatal("expected local denial")
	}
	if result.IngestionRunID != "" {
		t.Fatalf(
			"ingestion run id = %q, want empty after deletion",
			result.IngestionRunID,
		)
	}
	if repository.createCount != 1 ||
		repository.deleteCount != 1 ||
		repository.failedCount != 0 {
		t.Fatalf(
			"create=%d delete=%d failed=%d, want 1, 1, 0",
			repository.createCount,
			repository.deleteCount,
			repository.failedCount,
		)
	}
}

type orderedRegionalProvider struct {
	sourceName string
	states     []flightstate.FlightState
	err        error
	beforeLoad func()
	callCount  int
}

func (provider *orderedRegionalProvider) SourceName() string {
	return provider.sourceName
}

func (provider *orderedRegionalProvider) LoadByPoint(
	context.Context,
	float64,
	float64,
	int,
) ([]flightstate.FlightState, error) {
	provider.callCount++
	if provider.beforeLoad != nil {
		provider.beforeLoad()
	}
	return append(
		[]flightstate.FlightState(nil),
		provider.states...,
	), provider.err
}

type configuredProcessingService struct {
	result trafficapplication.ProcessAndStoreResult
	err    error
}

func (service *configuredProcessingService) ProcessAndStore(
	_ context.Context,
	states []flightstate.FlightState,
) (trafficapplication.ProcessAndStoreResult, error) {
	result := service.result
	if service.err == nil && result.StoredFlightStateCount == 0 {
		result.StoredFlightStateCount = len(states)
	}
	return result, service.err
}

type durableRunRepositoryStub struct {
	createErr         error
	createCount       int
	updateSourceCount int
	deleteCount       int
	successCount      int
	partialCount      int
	failedCount       int
	lastSourceName    string
}

func (repository *durableRunRepositoryStub) CreateRunning(
	_ context.Context,
	sourceName string,
	_ string,
	startedAt time.Time,
) (ingestionrun.Run, error) {
	repository.createCount++
	if repository.createErr != nil {
		return ingestionrun.Run{}, repository.createErr
	}
	repository.lastSourceName = sourceName
	return ingestionrun.Run{
		ID:         "run-durable",
		SourceName: sourceName,
		StartedAt:  startedAt,
		Status:     ingestionrun.StatusRunning,
	}, nil
}

func (repository *durableRunRepositoryStub) UpdateRunningSource(
	_ context.Context,
	_ string,
	sourceName string,
) error {
	repository.updateSourceCount++
	repository.lastSourceName = sourceName
	return nil
}

func (repository *durableRunRepositoryStub) DeleteRunning(
	context.Context,
	string,
) error {
	repository.deleteCount++
	return nil
}

func (repository *durableRunRepositoryStub) MarkSuccess(
	context.Context,
	string,
	time.Time,
	int,
	int,
	int,
) error {
	repository.successCount++
	return nil
}

func (repository *durableRunRepositoryStub) MarkPartial(
	context.Context,
	string,
	time.Time,
	int,
	int,
	int,
	string,
) error {
	repository.partialCount++
	return nil
}

func (repository *durableRunRepositoryStub) MarkFailed(
	context.Context,
	string,
	time.Time,
	int,
	int,
	int,
	string,
) error {
	repository.failedCount++
	return nil
}
