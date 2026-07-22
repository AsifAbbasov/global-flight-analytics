package ingestion

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

type LoadResult struct {
	SourceName string
	States     []flightstate.FlightState
}

type SourceAwareRegionalProvider interface {
	LoadByPointWithSource(
		ctx context.Context,
		latitude float64,
		longitude float64,
		radius int,
	) (LoadResult, error)
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
	TerminalTimeout        time.Duration
	Now                    func() time.Time
}

type Service struct {
	provider               RegionalProvider
	processingService      ProcessingService
	ingestionRunRepository IngestionRunRepository
	regionID               string
	terminalTimeout        time.Duration
	now                    func() time.Time
}

type LoadAndProcessResult struct {
	IngestionRunID   string
	SourceName       string
	LoadedStateCount int
	ProcessingResult trafficapplication.ProcessAndStoreResult
}

func New(config Config) *Service {
	now := config.Now
	if now == nil {
		now = time.Now
	}

	terminalTimeout := config.TerminalTimeout
	if terminalTimeout <= 0 {
		terminalTimeout = 15 * time.Second
	}

	return &Service{
		provider:               config.Provider,
		processingService:      config.ProcessingService,
		ingestionRunRepository: config.IngestionRunRepository,
		regionID:               config.RegionID,
		terminalTimeout:        terminalTimeout,
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

	loadResult, loadErr := service.loadByPoint(
		ctx,
		latitude,
		longitude,
		radius,
	)
	if loadErr != nil {
		return service.recordProviderFailure(
			ctx,
			startedAt,
			loadResult,
			loadErr,
		)
	}

	sourceName, err := normalizedSourceName(
		loadResult.SourceName,
		service.provider.SourceName(),
	)
	if err != nil {
		return LoadAndProcessResult{}, err
	}

	run, err := service.ingestionRunRepository.CreateRunning(
		ctx,
		sourceName,
		service.regionID,
		startedAt,
	)
	if err != nil {
		return LoadAndProcessResult{
				SourceName:       sourceName,
				LoadedStateCount: len(loadResult.States),
			}, fmt.Errorf(
				"create traffic ingestion run: %w",
				err,
			)
	}

	states := append(
		[]flightstate.FlightState(nil),
		loadResult.States...,
	)
	for index := range states {
		states[index].IngestionRunID = run.ID
		if strings.TrimSpace(states[index].SourceName) == "" {
			states[index].SourceName = sourceName
		}
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

		markErr := service.markRunFailed(
			ctx,
			run.ID,
			len(states),
			processingResult.StoredFlightStateCount,
			0,
			operationErr.Error(),
		)

		return LoadAndProcessResult{
			IngestionRunID:   run.ID,
			SourceName:       sourceName,
			LoadedStateCount: len(states),
			ProcessingResult: processingResult,
		}, errors.Join(operationErr, markErr)
	}

	err = service.markRunSuccess(
		ctx,
		run.ID,
		len(states),
		processingResult.StoredFlightStateCount,
		0,
	)
	if err != nil {
		return LoadAndProcessResult{
				IngestionRunID:   run.ID,
				SourceName:       sourceName,
				LoadedStateCount: len(states),
				ProcessingResult: processingResult,
			}, fmt.Errorf(
				"mark traffic ingestion run success: %w",
				err,
			)
	}

	return LoadAndProcessResult{
		IngestionRunID:   run.ID,
		SourceName:       sourceName,
		LoadedStateCount: len(states),
		ProcessingResult: processingResult,
	}, nil
}

func (service *Service) loadByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) (LoadResult, error) {
	sourceAwareProvider, supported :=
		service.provider.(SourceAwareRegionalProvider)
	if supported {
		return sourceAwareProvider.LoadByPointWithSource(
			ctx,
			latitude,
			longitude,
			radius,
		)
	}

	states, err := service.provider.LoadByPoint(
		ctx,
		latitude,
		longitude,
		radius,
	)
	return LoadResult{
		SourceName: service.provider.SourceName(),
		States:     states,
	}, err
}

func (service *Service) recordProviderFailure(
	ctx context.Context,
	startedAt time.Time,
	loadResult LoadResult,
	loadErr error,
) (LoadAndProcessResult, error) {
	operationErr := fmt.Errorf(
		"load regional flight states: %w",
		loadErr,
	)

	sourceName, sourceErr := normalizedSourceName(
		loadResult.SourceName,
		service.provider.SourceName(),
	)
	if sourceErr != nil {
		return LoadAndProcessResult{
			LoadedStateCount: len(loadResult.States),
		}, errors.Join(operationErr, sourceErr)
	}

	if !externalRequestAttempted(loadErr) {
		return LoadAndProcessResult{
			SourceName:       sourceName,
			LoadedStateCount: len(loadResult.States),
		}, operationErr
	}

	createContext, cancel := service.newTerminalContext(ctx)
	run, createErr := service.ingestionRunRepository.CreateRunning(
		createContext,
		sourceName,
		service.regionID,
		startedAt,
	)
	cancel()
	if createErr != nil {
		return LoadAndProcessResult{
				SourceName:       sourceName,
				LoadedStateCount: len(loadResult.States),
			}, errors.Join(
				operationErr,
				fmt.Errorf(
					"create failed traffic ingestion run: %w",
					createErr,
				),
			)
	}

	markErr := service.markRunFailed(
		ctx,
		run.ID,
		len(loadResult.States),
		0,
		0,
		operationErr.Error(),
	)

	return LoadAndProcessResult{
		IngestionRunID:   run.ID,
		SourceName:       sourceName,
		LoadedStateCount: len(loadResult.States),
	}, errors.Join(operationErr, markErr)
}

func (service *Service) markRunSuccess(
	ctx context.Context,
	runID string,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
) error {
	terminalContext, cancel := service.newTerminalContext(ctx)
	defer cancel()

	return service.ingestionRunRepository.MarkSuccess(
		terminalContext,
		runID,
		service.now().UTC(),
		recordsReceived,
		recordsInserted,
		recordsUpdated,
	)
}

func (service *Service) markRunFailed(
	ctx context.Context,
	runID string,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
	errorMessage string,
) error {
	terminalContext, cancel := service.newTerminalContext(ctx)
	defer cancel()

	return service.ingestionRunRepository.MarkFailed(
		terminalContext,
		runID,
		service.now().UTC(),
		recordsReceived,
		recordsInserted,
		recordsUpdated,
		errorMessage,
	)
}

func (service *Service) newTerminalContext(
	ctx context.Context,
) (context.Context, context.CancelFunc) {
	baseContext := context.Background()
	if ctx != nil {
		baseContext = context.WithoutCancel(ctx)
	}

	return context.WithTimeout(
		baseContext,
		service.terminalTimeout,
	)
}

func externalRequestAttempted(
	err error,
) bool {
	var evidence interface {
		ExternalRequestAttempted() bool
	}
	if errors.As(
		err,
		&evidence,
	) {
		return evidence.ExternalRequestAttempted()
	}

	return true
}

func normalizedSourceName(
	candidate string,
	fallback string,
) (string, error) {
	sourceName := strings.TrimSpace(candidate)
	if sourceName == "" {
		sourceName = strings.TrimSpace(fallback)
	}
	if sourceName == "" {
		return "", fmt.Errorf(
			"regional traffic provider source name is required",
		)
	}
	return sourceName, nil
}
