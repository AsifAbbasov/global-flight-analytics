package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
)

func TestSuccessfulRunFinalizesAfterCallerCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(context.Background())
	repository := &terminalContextRunRepository{
		run: ingestionrun.Run{ID: "run-success"},
	}
	service := New(Config{
		Provider: &terminalContextProvider{
			states: []flightstate.FlightState{{ICAO24: "ABC123"}},
		},
		ProcessingService: &terminalContextProcessingService{
			cancel: cancel,
			result: trafficapplication.ProcessAndStoreResult{
				StoredFlightStateCount: 1,
			},
		},
		IngestionRunRepository: repository,
		TerminalTimeout:        time.Second,
		Now:                    fixedIngestionTime,
	})

	_, err := service.LoadAndProcessByPoint(ctx, 40.4093, 49.8671, 100)
	if err != nil {
		t.Fatalf("load and process after caller cancellation: %v", err)
	}
	if repository.successCalls != 1 {
		t.Fatalf("success calls = %d, want 1", repository.successCalls)
	}
	if repository.successContextErr != nil {
		t.Fatalf(
			"success terminal context was already cancelled: %v",
			repository.successContextErr,
		)
	}
}

func TestFailedRunFinalizesAfterCallerCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(context.Background())
	processingErr := errors.New("processing failed")
	repository := &terminalContextRunRepository{
		run: ingestionrun.Run{ID: "run-failed"},
	}
	service := New(Config{
		Provider: &terminalContextProvider{
			states: []flightstate.FlightState{{ICAO24: "DEF456"}},
		},
		ProcessingService: &terminalContextProcessingService{
			cancel: cancel,
			err:    processingErr,
		},
		IngestionRunRepository: repository,
		TerminalTimeout:        time.Second,
		Now:                    fixedIngestionTime,
	})

	_, err := service.LoadAndProcessByPoint(ctx, 40.4093, 49.8671, 100)
	if !errors.Is(err, processingErr) {
		t.Fatalf("expected processing error, got %v", err)
	}
	if repository.failedCalls != 1 {
		t.Fatalf("failed calls = %d, want 1", repository.failedCalls)
	}
	if repository.failedContextErr != nil {
		t.Fatalf(
			"failed terminal context was already cancelled: %v",
			repository.failedContextErr,
		)
	}
}

func TestProviderFailureIsRecordedAfterCallerCancellation(
	t *testing.T,
) {
	ctx, cancel := context.WithCancel(context.Background())
	providerErr := errors.New("provider failed")
	repository := &terminalContextRunRepository{
		run: ingestionrun.Run{ID: "run-provider-failed"},
	}
	service := New(Config{
		Provider: &terminalContextProvider{
			cancel: cancel,
			err:    providerErr,
		},
		ProcessingService:      &terminalContextProcessingService{},
		IngestionRunRepository: repository,
		TerminalTimeout:        time.Second,
		Now:                    fixedIngestionTime,
	})

	_, err := service.LoadAndProcessByPoint(ctx, 40.4093, 49.8671, 100)
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
	if repository.createCalls != 1 || repository.failedCalls != 1 {
		t.Fatalf(
			"create calls = %d, failed calls = %d, want 1 and 1",
			repository.createCalls,
			repository.failedCalls,
		)
	}
	if repository.createContextErr != nil {
		t.Fatalf(
			"provider failure create context was already cancelled: %v",
			repository.createContextErr,
		)
	}
	if repository.failedContextErr != nil {
		t.Fatalf(
			"provider failure terminal context was already cancelled: %v",
			repository.failedContextErr,
		)
	}
}

type terminalContextProvider struct {
	states []flightstate.FlightState
	cancel context.CancelFunc
	err    error
}

func (provider *terminalContextProvider) SourceName() string {
	return "test-provider"
}

func (provider *terminalContextProvider) LoadByPoint(
	context.Context,
	float64,
	float64,
	int,
) ([]flightstate.FlightState, error) {
	if provider.cancel != nil {
		provider.cancel()
	}
	return append([]flightstate.FlightState(nil), provider.states...), provider.err
}

type terminalContextProcessingService struct {
	cancel context.CancelFunc
	result trafficapplication.ProcessAndStoreResult
	err    error
}

func (service *terminalContextProcessingService) ProcessAndStore(
	context.Context,
	[]flightstate.FlightState,
) (trafficapplication.ProcessAndStoreResult, error) {
	if service.cancel != nil {
		service.cancel()
	}
	return service.result, service.err
}

type terminalContextRunRepository struct {
	run ingestionrun.Run

	createCalls  int
	successCalls int
	failedCalls  int

	createContextErr  error
	successContextErr error
	failedContextErr  error
}

func (repository *terminalContextRunRepository) CreateRunning(
	ctx context.Context,
	_ string,
	_ string,
	_ time.Time,
) (ingestionrun.Run, error) {
	repository.createCalls++
	repository.createContextErr = ctx.Err()
	return repository.run, nil
}

func (repository *terminalContextRunRepository) UpdateRunningSource(
	context.Context,
	string,
	string,
) error {
	return nil
}

func (repository *terminalContextRunRepository) DeleteRunning(
	context.Context,
	string,
) error {
	return nil
}

func (repository *terminalContextRunRepository) MarkSuccess(
	ctx context.Context,
	_ string,
	_ time.Time,
	_ int,
	_ int,
	_ int,
) error {
	repository.successCalls++
	repository.successContextErr = ctx.Err()
	return nil
}

func (repository *terminalContextRunRepository) MarkPartial(
	context.Context,
	string,
	time.Time,
	int,
	int,
	int,
	string,
) error {
	return nil
}

func (repository *terminalContextRunRepository) MarkFailed(
	ctx context.Context,
	_ string,
	_ time.Time,
	_ int,
	_ int,
	_ int,
	_ string,
) error {
	repository.failedCalls++
	repository.failedContextErr = ctx.Err()
	return nil
}
