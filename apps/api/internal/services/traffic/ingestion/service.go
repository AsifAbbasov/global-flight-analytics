package ingestion

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/providerbatch"
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
	SourceName    string
	States        []flightstate.FlightState
	BatchEvidence providerbatch.Evidence
}

type SourceAwareRegionalProvider interface {
	LoadByPointWithSource(
		ctx context.Context,
		latitude float64,
		longitude float64,
		radius int,
	) (LoadResult, error)
}

type BatchEvidenceRegionalProvider interface {
	LoadByPointWithBatchEvidence(
		ctx context.Context,
		latitude float64,
		longitude float64,
		radius int,
	) (
		[]flightstate.FlightState,
		providerbatch.Evidence,
		error,
	)
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

	UpdateRunningSource(
		ctx context.Context,
		id string,
		sourceName string,
	) error

	DeleteRunning(
		ctx context.Context,
		id string,
	) error

	MarkSuccess(
		ctx context.Context,
		id string,
		finishedAt time.Time,
		recordsReceived int,
		recordsInserted int,
		recordsUpdated int,
	) error

	MarkPartial(
		ctx context.Context,
		id string,
		finishedAt time.Time,
		recordsReceived int,
		recordsInserted int,
		recordsUpdated int,
		errorMessage string,
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
	IngestionRunID        string
	SourceName            string
	LoadedStateCount      int
	ProviderBatchEvidence providerbatch.Evidence
	ProcessingResult      trafficapplication.ProcessAndStoreResult
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
	initialSourceName, err := normalizedSourceName(
		service.provider.SourceName(),
		"",
	)
	if err != nil {
		return LoadAndProcessResult{}, err
	}

	// The running row is committed before any provider call. A process crash,
	// container termination, or panic during transport therefore leaves durable
	// evidence that startup stale-run recovery can finalize.
	run, err := service.ingestionRunRepository.CreateRunning(
		ctx,
		initialSourceName,
		service.regionID,
		startedAt,
	)
	if err != nil {
		return LoadAndProcessResult{
				SourceName: initialSourceName,
			}, fmt.Errorf(
				"create traffic ingestion run before provider request: %w",
				err,
			)
	}

	loadResult, loadErr := service.loadByPoint(
		ctx,
		latitude,
		longitude,
		radius,
	)
	if loadErr != nil {
		return service.recordProviderFailure(
			ctx,
			run,
			loadResult,
			loadErr,
		)
	}

	sourceName, err := normalizedSourceName(
		loadResult.SourceName,
		initialSourceName,
	)
	if err != nil {
		operationErr := fmt.Errorf(
			"resolve selected regional traffic provider: %w",
			err,
		)
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
			LoadedStateCount: len(loadResult.States),
		}, errors.Join(operationErr, markErr)
	}

	if err := service.updateRunningSource(
		ctx,
		run.ID,
		run.SourceName,
		sourceName,
	); err != nil {
		operationErr := fmt.Errorf(
			"update traffic ingestion run selected source: %w",
			err,
		)
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

	batchEvidence, err := providerbatch.Resolve(
		loadResult.BatchEvidence,
		len(loadResult.States),
	)
	if err != nil {
		operationErr := fmt.Errorf(
			"validate provider batch evidence: %w",
			err,
		)
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

		var markErr error
		if processingResult.StoredFlightStateCount > 0 {
			markErr = service.markRunPartial(
				ctx,
				run.ID,
				batchEvidence.Received,
				processingResult.StoredFlightStateCount,
				0,
				operationErr.Error(),
			)
		} else {
			markErr = service.markRunFailed(
				ctx,
				run.ID,
				batchEvidence.Received,
				0,
				0,
				operationErr.Error(),
			)
		}

		return LoadAndProcessResult{
			IngestionRunID:        run.ID,
			SourceName:            sourceName,
			LoadedStateCount:      len(states),
			ProviderBatchEvidence: batchEvidence,
			ProcessingResult:      processingResult,
		}, errors.Join(operationErr, markErr)
	}

	partialBatch, partialMessage :=
		providerBatchPartialFailure(batchEvidence)
	if partialBatch {
		err = service.markRunPartial(
			ctx,
			run.ID,
			batchEvidence.Received,
			processingResult.StoredFlightStateCount,
			0,
			partialMessage,
		)
	} else {
		err = service.markRunSuccess(
			ctx,
			run.ID,
			batchEvidence.Received,
			processingResult.StoredFlightStateCount,
			0,
		)
	}
	if err != nil {
		return LoadAndProcessResult{
				IngestionRunID:        run.ID,
				SourceName:            sourceName,
				LoadedStateCount:      len(states),
				ProviderBatchEvidence: batchEvidence,
				ProcessingResult:      processingResult,
			}, fmt.Errorf(
				"mark traffic ingestion run success: %w",
				err,
			)
	}

	return LoadAndProcessResult{
		IngestionRunID:        run.ID,
		SourceName:            sourceName,
		LoadedStateCount:      len(states),
		ProviderBatchEvidence: batchEvidence,
		ProcessingResult:      processingResult,
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

	evidenceProvider, supported :=
		service.provider.(BatchEvidenceRegionalProvider)
	if supported {
		states, evidence, err :=
			evidenceProvider.LoadByPointWithBatchEvidence(
				ctx,
				latitude,
				longitude,
				radius,
			)
		return LoadResult{
			SourceName:    service.provider.SourceName(),
			States:        states,
			BatchEvidence: evidence,
		}, err
	}

	states, err := service.provider.LoadByPoint(
		ctx,
		latitude,
		longitude,
		radius,
	)
	return LoadResult{
		SourceName:    service.provider.SourceName(),
		States:        states,
		BatchEvidence: providerbatch.AcceptedOnly(len(states)),
	}, err
}

func (service *Service) recordProviderFailure(
	ctx context.Context,
	run ingestionrun.Run,
	loadResult LoadResult,
	loadErr error,
) (LoadAndProcessResult, error) {
	operationErr := fmt.Errorf(
		"load regional flight states: %w",
		loadErr,
	)

	sourceName, sourceErr := normalizedSourceName(
		loadResult.SourceName,
		run.SourceName,
	)
	if sourceErr != nil {
		sourceName = strings.TrimSpace(run.SourceName)
	}

	// A local budget or polling denial did not execute provider transport. The
	// provisional row is removed instead of being retained as a false failed run.
	if !externalRequestAttempted(loadErr) {
		deleteErr := service.deleteRunningRun(
			ctx,
			run.ID,
		)
		result := LoadAndProcessResult{
			SourceName:       sourceName,
			LoadedStateCount: len(loadResult.States),
		}
		if deleteErr != nil {
			result.IngestionRunID = run.ID
			deleteErr = fmt.Errorf(
				"delete ingestion run without external request: %w",
				deleteErr,
			)
		}
		return result, errors.Join(operationErr, sourceErr, deleteErr)
	}

	updateErr := service.updateRunningSource(
		ctx,
		run.ID,
		run.SourceName,
		sourceName,
	)
	if updateErr != nil {
		updateErr = fmt.Errorf(
			"update failed provider source evidence: %w",
			updateErr,
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
	}, errors.Join(operationErr, sourceErr, updateErr, markErr)
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

func (service *Service) markRunPartial(
	ctx context.Context,
	runID string,
	recordsReceived int,
	recordsInserted int,
	recordsUpdated int,
	errorMessage string,
) error {
	terminalContext, cancel := service.newTerminalContext(ctx)
	defer cancel()

	return service.ingestionRunRepository.MarkPartial(
		terminalContext,
		runID,
		service.now().UTC(),
		recordsReceived,
		recordsInserted,
		recordsUpdated,
		errorMessage,
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

func (service *Service) updateRunningSource(
	ctx context.Context,
	runID string,
	currentSourceName string,
	selectedSourceName string,
) error {
	if strings.TrimSpace(currentSourceName) ==
		strings.TrimSpace(selectedSourceName) {
		return nil
	}

	terminalContext, cancel := service.newTerminalContext(ctx)
	defer cancel()

	return service.ingestionRunRepository.UpdateRunningSource(
		terminalContext,
		runID,
		selectedSourceName,
	)
}

func (service *Service) deleteRunningRun(
	ctx context.Context,
	runID string,
) error {
	terminalContext, cancel := service.newTerminalContext(ctx)
	defer cancel()

	return service.ingestionRunRepository.DeleteRunning(
		terminalContext,
		runID,
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
