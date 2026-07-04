package ingestion

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
)

type RegionalProvider interface {
	SourceName() string

	LoadByPoint(
		ctx context.Context,
		latitude float64,
		longitude float64,
		radius int,
	) ([]flightstate.FlightState, error)
}

type ProcessingService interface {
	ProcessAndStore(
		ctx context.Context,
		states []flightstate.FlightState,
	) (trafficapplication.ProcessAndStoreResult, error)
}

type IngestionRunRepository interface {
	CreateRunning(
		ctx context.Context,
		sourceName string,
		regionID string,
		startedAt time.Time,
	) (ingestionrun.Run, error)

	MarkSuccess(
		ctx context.Context,
		id string,
		finishedAt time.Time,
		recordsReceived int,
		recordsInserted int,
		recordsUpdated int,
	) error

	MarkFailed(
		ctx context.Context,
		id string,
		finishedAt time.Time,
		recordsReceived int,
		recordsInserted int,
		recordsUpdated int,
		errorMessage string,
	) error
}

type Config struct {
	Provider               RegionalProvider
	ProcessingService      ProcessingService
	IngestionRunRepository IngestionRunRepository
	RegionID               string
	Now                    func() time.Time
}

type Service struct {
	provider               RegionalProvider
	processingService      ProcessingService
	ingestionRunRepository IngestionRunRepository
	regionID               string
	now                    func() time.Time
}

type LoadAndProcessResult struct {
	IngestionRunID   string
	LoadedStateCount int
	ProcessingResult trafficapplication.ProcessAndStoreResult
}

func New(config Config) *Service {
	now := config.Now
	if now == nil {
		now = time.Now
	}

	return &Service{
		provider:               config.Provider,
		processingService:      config.ProcessingService,
		ingestionRunRepository: config.IngestionRunRepository,
		regionID:               config.RegionID,
		now:                    now,
	}
}

func (service *Service) LoadAndProcessByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) (LoadAndProcessResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if service == nil || service.provider == nil {
		return LoadAndProcessResult{}, fmt.Errorf(
			"regional traffic provider is required",
		)
	}

	if service.processingService == nil {
		return LoadAndProcessResult{}, fmt.Errorf(
			"traffic processing service is required",
		)
	}

	if service.ingestionRunRepository == nil {
		return LoadAndProcessResult{}, fmt.Errorf(
			"ingestion run repository is required",
		)
	}

	startedAt := service.now().UTC()

	run, err := service.ingestionRunRepository.CreateRunning(
		ctx,
		service.provider.SourceName(),
		service.regionID,
		startedAt,
	)
	if err != nil {
		return LoadAndProcessResult{}, fmt.Errorf(
			"create traffic ingestion run: %w",
			err,
		)
	}

	states, err := service.provider.LoadByPoint(
		ctx,
		latitude,
		longitude,
		radius,
	)
	if err != nil {
		operationErr := fmt.Errorf(
			"load regional flight states: %w",
			err,
		)

		markErr := service.ingestionRunRepository.MarkFailed(
			ctx,
			run.ID,
			service.now().UTC(),
			len(states),
			0,
			0,
			operationErr.Error(),
		)

		return LoadAndProcessResult{
			IngestionRunID:   run.ID,
			LoadedStateCount: len(states),
		}, errors.Join(operationErr, markErr)
	}

	for index := range states {
		states[index].IngestionRunID = run.ID
	}

	processingResult, err := service.processingService.ProcessAndStore(
		ctx,
		states,
	)
	if err != nil {
		operationErr := fmt.Errorf(
			"process and store regional flight states: %w",
			err,
		)

		markErr := service.ingestionRunRepository.MarkFailed(
			ctx,
			run.ID,
			service.now().UTC(),
			len(states),
			processingResult.StoredFlightStateCount,
			0,
			operationErr.Error(),
		)

		return LoadAndProcessResult{
			IngestionRunID:   run.ID,
			LoadedStateCount: len(states),
			ProcessingResult: processingResult,
		}, errors.Join(operationErr, markErr)
	}

	err = service.ingestionRunRepository.MarkSuccess(
		ctx,
		run.ID,
		service.now().UTC(),
		len(states),
		processingResult.StoredFlightStateCount,
		0,
	)
	if err != nil {
		return LoadAndProcessResult{
				IngestionRunID:   run.ID,
				LoadedStateCount: len(states),
				ProcessingResult: processingResult,
			}, fmt.Errorf(
				"mark traffic ingestion run success: %w",
				err,
			)
	}

	return LoadAndProcessResult{
		IngestionRunID:   run.ID,
		LoadedStateCount: len(states),
		ProcessingResult: processingResult,
	}, nil
}
