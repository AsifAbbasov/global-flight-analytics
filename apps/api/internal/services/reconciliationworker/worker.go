package reconciliationworker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
)

var (
	ErrRepositoryRequired              = errors.New("reconciliation worker repository is required")
	ErrFlightStateRepositoryRequired   = errors.New("reconciliation worker flight state repository is required")
	ErrDataQualityRepositoryRequired   = errors.New("reconciliation worker data quality repository is required")
	ErrTrajectoryRepositoryRequired    = errors.New("reconciliation worker trajectory repository is required")
	ErrProcessorRequired               = errors.New("reconciliation worker processor is required")
	ErrMaxAttemptsInvalid              = errors.New("reconciliation worker maximum attempts must be greater than zero")
	ErrRetryBaseDelayInvalid           = errors.New("reconciliation worker retry base delay must be greater than zero")
	ErrRetryMaximumDelayInvalid        = errors.New("reconciliation worker retry maximum delay must be greater than zero")
	ErrRetryDelayRangeInvalid          = errors.New("reconciliation worker retry maximum delay must not be less than retry base delay")
	ErrSourceFlightStateNotFound       = errors.New("reconciliation source flight state was not found")
	ErrSourceFlightStateAmbiguous      = errors.New("reconciliation source flight state is ambiguous")
	ErrQualityDerivationUnavailable    = errors.New("flight state quality derivation is unavailable")
	ErrTrajectoryDerivationUnavailable = errors.New("trajectory derivation is unavailable")
	ErrDerivationTypeUnsupported       = errors.New("reconciliation derivation type is unsupported")
)

type Clock func() time.Time

type TaskRepository interface {
	ClaimNextAvailable(
		ctx context.Context,
	) (reconciliation.Task, error)
	MarkCompleted(
		ctx context.Context,
		taskID string,
		attemptCount int,
	) (reconciliation.TaskStatus, error)
	MarkRetry(
		ctx context.Context,
		taskID string,
		attemptCount int,
		nextAttemptAt time.Time,
		lastError string,
	) error
	MarkFailed(
		ctx context.Context,
		taskID string,
		attemptCount int,
		lastError string,
	) (reconciliation.TaskStatus, error)
}

type FlightStateRepository interface {
	ListByReconciliationScope(
		ctx context.Context,
		icao24 string,
		ingestionRunID string,
		observedFrom time.Time,
		observedTo time.Time,
	) ([]flightstate.FlightState, error)
}

type DataQualityRepository interface {
	SaveReconciledFlightStateQuality(
		ctx context.Context,
		taskID string,
		attemptCount int,
		state flightstate.FlightState,
		quality dataquality.DataQuality,
	) error
}

type TrajectoryRepository interface {
	SaveReconciledTrajectory(
		ctx context.Context,
		taskID string,
		attemptCount int,
		item trajectory.FlightTrajectory,
	) error
}

type Config struct {
	Repository            TaskRepository
	FlightStateRepository FlightStateRepository
	DataQualityRepository DataQualityRepository
	TrajectoryRepository  TrajectoryRepository
	Processor             *processor.Processor
	Now                   Clock
	MaxAttempts           int
	RetryBaseDelay        time.Duration
	RetryMaximumDelay     time.Duration
}

type Worker struct {
	repository            TaskRepository
	flightStateRepository FlightStateRepository
	dataQualityRepository DataQualityRepository
	trajectoryRepository  TrajectoryRepository
	processor             *processor.Processor
	now                   Clock
	maxAttempts           int
	retryBaseDelay        time.Duration
	retryMaximumDelay     time.Duration
}

type RunResult struct {
	TaskFound          bool
	TaskID             string
	ICAO24             string
	DerivationType     reconciliation.DerivationType
	AttemptCount       int
	FinalStatus        reconciliation.TaskStatus
	PersistedItemCount int
	NextAttemptAt      time.Time
	LastError          string
}

func New(
	config Config,
) (*Worker, error) {
	if config.Repository == nil {
		return nil, ErrRepositoryRequired
	}

	if config.FlightStateRepository == nil {
		return nil, ErrFlightStateRepositoryRequired
	}

	if config.DataQualityRepository == nil {
		return nil, ErrDataQualityRepositoryRequired
	}

	if config.TrajectoryRepository == nil {
		return nil, ErrTrajectoryRepositoryRequired
	}

	if config.Processor == nil {
		return nil, ErrProcessorRequired
	}

	if config.MaxAttempts <= 0 {
		return nil, ErrMaxAttemptsInvalid
	}

	if config.RetryBaseDelay <= 0 {
		return nil, ErrRetryBaseDelayInvalid
	}

	if config.RetryMaximumDelay <= 0 {
		return nil, ErrRetryMaximumDelayInvalid
	}

	if config.RetryMaximumDelay < config.RetryBaseDelay {
		return nil, ErrRetryDelayRangeInvalid
	}

	now := config.Now
	if now == nil {
		now = func() time.Time {
			return time.Now().UTC()
		}
	}

	return &Worker{
		repository:            config.Repository,
		flightStateRepository: config.FlightStateRepository,
		dataQualityRepository: config.DataQualityRepository,
		trajectoryRepository:  config.TrajectoryRepository,
		processor:             config.Processor,
		now:                   now,
		maxAttempts:           config.MaxAttempts,
		retryBaseDelay:        config.RetryBaseDelay,
		retryMaximumDelay:     config.RetryMaximumDelay,
	}, nil
}

func (worker *Worker) RunOnce(
	ctx context.Context,
) (RunResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	task, err := worker.repository.ClaimNextAvailable(
		ctx,
	)
	if errors.Is(
		err,
		reconciliation.ErrNoTaskAvailable,
	) {
		return RunResult{}, nil
	}
	if err != nil {
		return RunResult{}, fmt.Errorf(
			"claim reconciliation task: %w",
			err,
		)
	}

	result := RunResult{
		TaskFound:      true,
		TaskID:         task.ID,
		ICAO24:         task.ICAO24,
		DerivationType: task.DerivationType,
		AttemptCount:   task.AttemptCount,
		FinalStatus:    reconciliation.TaskStatusProcessing,
	}

	persistedItemCount, executionErr := worker.execute(
		ctx,
		task,
	)
	if executionErr != nil {
		return worker.finishFailure(
			ctx,
			task,
			result,
			executionErr,
		)
	}

	status, err := worker.repository.MarkCompleted(
		ctx,
		task.ID,
		task.AttemptCount,
	)
	if err != nil {
		return result, fmt.Errorf(
			"mark reconciliation task completed: %w",
			err,
		)
	}

	result.FinalStatus = status
	result.PersistedItemCount = persistedItemCount

	return result, nil
}

func (worker *Worker) execute(
	ctx context.Context,
	task reconciliation.Task,
) (int, error) {
	states, err := worker.flightStateRepository.ListByReconciliationScope(
		ctx,
		task.ICAO24,
		task.IngestionRunID,
		task.ObservedFrom,
		task.ObservedTo,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"load reconciliation source flight states: %w",
			err,
		)
	}

	if len(states) == 0 {
		return 0, ErrSourceFlightStateNotFound
	}

	switch task.DerivationType {
	case reconciliation.DerivationTypeFlightStateQuality:
		return worker.reconcileFlightStateQuality(
			ctx,
			task,
			states,
		)

	case reconciliation.DerivationTypeTrajectory:
		return worker.reconcileTrajectory(
			ctx,
			task,
			states,
		)

	default:
		return 0, fmt.Errorf(
			"%w: %s",
			ErrDerivationTypeUnsupported,
			task.DerivationType,
		)
	}
}

func (worker *Worker) reconcileFlightStateQuality(
	ctx context.Context,
	task reconciliation.Task,
	states []flightstate.FlightState,
) (int, error) {
	if len(states) != 1 {
		return 0, fmt.Errorf(
			"%w: expected 1 state, got %d",
			ErrSourceFlightStateAmbiguous,
			len(states),
		)
	}

	processingResult := worker.processor.Process(
		states,
	)
	if len(processingResult.UsableStates) != 1 {
		return 0, fmt.Errorf(
			"%w: expected 1 usable state, got %d",
			ErrQualityDerivationUnavailable,
			len(processingResult.UsableStates),
		)
	}

	item := processingResult.UsableStates[0]

	if err := worker.dataQualityRepository.SaveReconciledFlightStateQuality(
		ctx,
		task.ID,
		task.AttemptCount,
		item.State,
		item.Quality,
	); err != nil {
		return 0, fmt.Errorf(
			"save reconciled flight state quality: %w",
			err,
		)
	}

	return 1, nil
}

func (worker *Worker) reconcileTrajectory(
	ctx context.Context,
	task reconciliation.Task,
	states []flightstate.FlightState,
) (int, error) {
	processingResult := worker.processor.Process(
		states,
	)

	trajectoryKey := strings.ToUpper(
		strings.TrimSpace(
			task.ICAO24,
		),
	)

	item, exists := processingResult.Trajectories[trajectoryKey]
	if !exists {
		return 0, fmt.Errorf(
			"%w for icao24 %s",
			ErrTrajectoryDerivationUnavailable,
			task.ICAO24,
		)
	}

	if err := worker.trajectoryRepository.SaveReconciledTrajectory(
		ctx,
		task.ID,
		task.AttemptCount,
		item,
	); err != nil {
		return 0, fmt.Errorf(
			"save reconciled trajectory for icao24 %s: %w",
			task.ICAO24,
			err,
		)
	}

	return 1, nil
}

func (worker *Worker) finishFailure(
	ctx context.Context,
	task reconciliation.Task,
	result RunResult,
	cause error,
) (RunResult, error) {
	result.LastError = reconciliation.NormalizeLastError(
		cause.Error(),
	)

	if errors.Is(cause, context.Canceled) ||
		errors.Is(cause, context.DeadlineExceeded) {
		return result, cause
	}

	if isPermanentFailure(cause) ||
		task.AttemptCount >= worker.maxAttempts {
		status, err := worker.repository.MarkFailed(
			ctx,
			task.ID,
			task.AttemptCount,
			result.LastError,
		)
		if err != nil {
			return result, errors.Join(
				cause,
				fmt.Errorf(
					"mark reconciliation task failed: %w",
					err,
				),
			)
		}

		result.FinalStatus = status

		return result, nil
	}

	nextAttemptAt := worker.now().
		UTC().
		Add(
			worker.retryDelay(
				task.AttemptCount,
			),
		)

	if err := worker.repository.MarkRetry(
		ctx,
		task.ID,
		task.AttemptCount,
		nextAttemptAt,
		result.LastError,
	); err != nil {
		return result, errors.Join(
			cause,
			fmt.Errorf(
				"mark reconciliation task for retry: %w",
				err,
			),
		)
	}

	result.FinalStatus = reconciliation.TaskStatusPending
	result.NextAttemptAt = nextAttemptAt

	return result, nil
}

func (worker *Worker) retryDelay(
	attemptCount int,
) time.Duration {
	delay := worker.retryBaseDelay

	for attempt := 1; attempt < attemptCount; attempt++ {
		if delay >= worker.retryMaximumDelay {
			return worker.retryMaximumDelay
		}

		if delay > worker.retryMaximumDelay/2 {
			return worker.retryMaximumDelay
		}

		delay *= 2
	}

	if delay > worker.retryMaximumDelay {
		return worker.retryMaximumDelay
	}

	return delay
}

func isPermanentFailure(
	err error,
) bool {
	return errors.Is(err, ErrSourceFlightStateNotFound) ||
		errors.Is(err, ErrSourceFlightStateAmbiguous) ||
		errors.Is(err, ErrQualityDerivationUnavailable) ||
		errors.Is(err, ErrTrajectoryDerivationUnavailable) ||
		errors.Is(err, ErrDerivationTypeUnsupported)
}
