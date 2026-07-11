package reconciliationworker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
)

func TestRunOnceCompletesFlightStateQualityTask(
	t *testing.T,
) {
	task := makeTask(
		reconciliation.DerivationTypeFlightStateQuality,
		1,
	)

	repository := &repositoryStub{
		task: task,
	}
	flightStateRepository := &flightStateRepositoryStub{
		states: []flightstate.FlightState{
			makeFlightState(
				task.ObservedFrom,
			),
		},
	}
	dataQualityRepository := &dataQualityRepositoryStub{}
	trajectoryRepository := &trajectoryRepositoryStub{}

	worker := newTestWorker(
		t,
		repository,
		flightStateRepository,
		dataQualityRepository,
		trajectoryRepository,
	)

	result, err := worker.RunOnce(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"run worker: %v",
			err,
		)
	}

	if !result.TaskFound {
		t.Fatal(
			"expected a task",
		)
	}

	if result.FinalStatus != reconciliation.TaskStatusCompleted {
		t.Fatalf(
			"expected completed status, got %s",
			result.FinalStatus,
		)
	}

	if result.PersistedItemCount != 1 {
		t.Fatalf(
			"expected one persisted item, got %d",
			result.PersistedItemCount,
		)
	}

	if dataQualityRepository.saveCount != 1 {
		t.Fatalf(
			"expected one quality save, got %d",
			dataQualityRepository.saveCount,
		)
	}

	if repository.completedCount != 1 {
		t.Fatalf(
			"expected one completed transition, got %d",
			repository.completedCount,
		)
	}

	if dataQualityRepository.savedTaskID != task.ID {
		t.Fatalf(
			"expected quality save task id %s, got %s",
			task.ID,
			dataQualityRepository.savedTaskID,
		)
	}

	if dataQualityRepository.savedAttemptCount != task.AttemptCount {
		t.Fatalf(
			"expected quality save attempt %d, got %d",
			task.AttemptCount,
			dataQualityRepository.savedAttemptCount,
		)
	}
}

func TestRunOnceCompletesTrajectoryTask(
	t *testing.T,
) {
	task := makeTask(
		reconciliation.DerivationTypeTrajectory,
		1,
	)

	repository := &repositoryStub{
		task: task,
	}
	flightStateRepository := &flightStateRepositoryStub{
		states: []flightstate.FlightState{
			makeFlightState(
				task.ObservedFrom,
			),
			makeFlightState(
				task.ObservedTo,
			),
		},
	}
	dataQualityRepository := &dataQualityRepositoryStub{}
	trajectoryRepository := &trajectoryRepositoryStub{}

	worker := newTestWorker(
		t,
		repository,
		flightStateRepository,
		dataQualityRepository,
		trajectoryRepository,
	)

	result, err := worker.RunOnce(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"run worker: %v",
			err,
		)
	}

	if result.FinalStatus != reconciliation.TaskStatusCompleted {
		t.Fatalf(
			"expected completed status, got %s",
			result.FinalStatus,
		)
	}

	if trajectoryRepository.saveCount != 1 {
		t.Fatalf(
			"expected one trajectory save, got %d",
			trajectoryRepository.saveCount,
		)
	}

	if trajectoryRepository.savedTrajectory.ICAO24 != "ABC123" {
		t.Fatalf(
			"expected normalized trajectory icao24 ABC123, got %q",
			trajectoryRepository.savedTrajectory.ICAO24,
		)
	}

	if trajectoryRepository.savedTaskID != task.ID {
		t.Fatalf(
			"expected trajectory save task id %s, got %s",
			task.ID,
			trajectoryRepository.savedTaskID,
		)
	}

	if trajectoryRepository.savedAttemptCount != task.AttemptCount {
		t.Fatalf(
			"expected trajectory save attempt %d, got %d",
			task.AttemptCount,
			trajectoryRepository.savedAttemptCount,
		)
	}
}

func TestRunOnceSchedulesRetryForRecoverableFailure(
	t *testing.T,
) {
	task := makeTask(
		reconciliation.DerivationTypeTrajectory,
		2,
	)

	repository := &repositoryStub{
		task: task,
	}
	flightStateRepository := &flightStateRepositoryStub{
		states: []flightstate.FlightState{
			makeFlightState(
				task.ObservedFrom,
			),
		},
	}
	dataQualityRepository := &dataQualityRepositoryStub{}
	trajectoryRepository := &trajectoryRepositoryStub{
		saveErr: errors.New(
			"temporary database failure",
		),
	}

	now := time.Date(
		2026,
		time.July,
		11,
		18,
		0,
		0,
		0,
		time.UTC,
	)

	worker := newTestWorkerWithClock(
		t,
		repository,
		flightStateRepository,
		dataQualityRepository,
		trajectoryRepository,
		func() time.Time {
			return now
		},
	)

	result, err := worker.RunOnce(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"run worker: %v",
			err,
		)
	}

	expectedNextAttemptAt := now.Add(
		2 * time.Minute,
	)

	if result.FinalStatus != reconciliation.TaskStatusPending {
		t.Fatalf(
			"expected pending status, got %s",
			result.FinalStatus,
		)
	}

	if !result.NextAttemptAt.Equal(
		expectedNextAttemptAt,
	) {
		t.Fatalf(
			"expected next attempt at %s, got %s",
			expectedNextAttemptAt,
			result.NextAttemptAt,
		)
	}

	if repository.retryCount != 1 {
		t.Fatalf(
			"expected one retry transition, got %d",
			repository.retryCount,
		)
	}
}

func TestRunOnceMarksPermanentSourceFailureFailed(
	t *testing.T,
) {
	task := makeTask(
		reconciliation.DerivationTypeTrajectory,
		1,
	)

	repository := &repositoryStub{
		task: task,
	}
	flightStateRepository := &flightStateRepositoryStub{}
	dataQualityRepository := &dataQualityRepositoryStub{}
	trajectoryRepository := &trajectoryRepositoryStub{}

	worker := newTestWorker(
		t,
		repository,
		flightStateRepository,
		dataQualityRepository,
		trajectoryRepository,
	)

	result, err := worker.RunOnce(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"run worker: %v",
			err,
		)
	}

	if result.FinalStatus != reconciliation.TaskStatusFailed {
		t.Fatalf(
			"expected failed status, got %s",
			result.FinalStatus,
		)
	}

	if repository.failedCount != 1 {
		t.Fatalf(
			"expected one failed transition, got %d",
			repository.failedCount,
		)
	}
}

func TestRunOnceMarksRecoverableFailureFailedAtMaximumAttempts(
	t *testing.T,
) {
	task := makeTask(
		reconciliation.DerivationTypeTrajectory,
		5,
	)

	repository := &repositoryStub{
		task: task,
	}
	flightStateRepository := &flightStateRepositoryStub{
		states: []flightstate.FlightState{
			makeFlightState(
				task.ObservedFrom,
			),
		},
	}
	dataQualityRepository := &dataQualityRepositoryStub{}
	trajectoryRepository := &trajectoryRepositoryStub{
		saveErr: errors.New(
			"database remains unavailable",
		),
	}

	worker := newTestWorker(
		t,
		repository,
		flightStateRepository,
		dataQualityRepository,
		trajectoryRepository,
	)

	result, err := worker.RunOnce(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"run worker: %v",
			err,
		)
	}

	if result.FinalStatus != reconciliation.TaskStatusFailed {
		t.Fatalf(
			"expected failed status, got %s",
			result.FinalStatus,
		)
	}

	if repository.failedCount != 1 {
		t.Fatalf(
			"expected one failed transition, got %d",
			repository.failedCount,
		)
	}
}

func TestRunOnceReturnsNoTaskWithoutError(
	t *testing.T,
) {
	repository := &repositoryStub{
		claimErr: reconciliation.ErrNoTaskAvailable,
	}

	worker := newTestWorker(
		t,
		repository,
		&flightStateRepositoryStub{},
		&dataQualityRepositoryStub{},
		&trajectoryRepositoryStub{},
	)

	result, err := worker.RunOnce(
		context.Background(),
	)
	if err != nil {
		t.Fatalf(
			"run worker: %v",
			err,
		)
	}

	if result.TaskFound {
		t.Fatal(
			"did not expect a task",
		)
	}
}

func newTestWorker(
	t *testing.T,
	repository *repositoryStub,
	flightStateRepository *flightStateRepositoryStub,
	dataQualityRepository *dataQualityRepositoryStub,
	trajectoryRepository *trajectoryRepositoryStub,
) *Worker {
	t.Helper()

	return newTestWorkerWithClock(
		t,
		repository,
		flightStateRepository,
		dataQualityRepository,
		trajectoryRepository,
		func() time.Time {
			return time.Date(
				2026,
				time.July,
				11,
				18,
				0,
				0,
				0,
				time.UTC,
			)
		},
	)
}

func newTestWorkerWithClock(
	t *testing.T,
	repository *repositoryStub,
	flightStateRepository *flightStateRepositoryStub,
	dataQualityRepository *dataQualityRepositoryStub,
	trajectoryRepository *trajectoryRepositoryStub,
	now Clock,
) *Worker {
	t.Helper()

	trafficProcessor, err := processor.New(
		processor.Config{
			Now: processor.Clock(now),
		},
	)
	if err != nil {
		t.Fatalf(
			"create processor: %v",
			err,
		)
	}

	worker, err := New(
		Config{
			Repository:            repository,
			FlightStateRepository: flightStateRepository,
			DataQualityRepository: dataQualityRepository,
			TrajectoryRepository:  trajectoryRepository,
			Processor:             trafficProcessor,
			Now:                   now,
			MaxAttempts:           5,
			RetryBaseDelay:        time.Minute,
			RetryMaximumDelay:     8 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf(
			"create worker: %v",
			err,
		)
	}

	return worker
}

func makeTask(
	derivationType reconciliation.DerivationType,
	attemptCount int,
) reconciliation.Task {
	observedFrom := time.Date(
		2026,
		time.July,
		11,
		17,
		0,
		0,
		0,
		time.UTC,
	)

	return reconciliation.Task{
		ID:                   "task-id",
		ICAO24:               "abc123",
		DerivationType:       derivationType,
		Status:               reconciliation.TaskStatusProcessing,
		ObservedFrom:         observedFrom,
		ObservedTo:           observedFrom.Add(time.Minute),
		AttemptCount:         attemptCount,
		SignalVersion:        1,
		ClaimedSignalVersion: 1,
	}
}

func makeFlightState(
	observedAt time.Time,
) flightstate.FlightState {
	return flightstate.FlightState{
		ID:                       "state-id",
		ICAO24:                   "abc123",
		Callsign:                 "TEST123",
		Latitude:                 40.4675,
		Longitude:                50.0467,
		BarometricAltitudeM:      1000,
		BarometricAltitudeStatus: flightstate.AltitudeStatusObserved,
		GeometricAltitudeM:       1100,
		GeometricAltitudeStatus:  flightstate.AltitudeStatusObserved,
		VelocityMPS:              200,
		HeadingDegrees:           90,
		VerticalRateMPS:          0,
		OriginCountry:            "Azerbaijan",
		ObservedAt:               observedAt,
		SourceName:               "test",
	}
}

type repositoryStub struct {
	task           reconciliation.Task
	claimErr       error
	completedCount int
	retryCount     int
	failedCount    int
}

func (repository *repositoryStub) ClaimNextAvailable(
	context.Context,
) (reconciliation.Task, error) {
	if repository.claimErr != nil {
		return reconciliation.Task{}, repository.claimErr
	}

	return repository.task, nil
}

func (repository *repositoryStub) MarkCompleted(
	context.Context,
	string,
	int,
) (reconciliation.TaskStatus, error) {
	repository.completedCount++

	return reconciliation.TaskStatusCompleted, nil
}

func (repository *repositoryStub) MarkRetry(
	context.Context,
	string,
	int,
	time.Time,
	string,
) error {
	repository.retryCount++

	return nil
}

func (repository *repositoryStub) MarkFailed(
	context.Context,
	string,
	int,
	string,
) (reconciliation.TaskStatus, error) {
	repository.failedCount++

	return reconciliation.TaskStatusFailed, nil
}

func (repository *repositoryStub) RequeueStaleProcessing(
	context.Context,
	time.Time,
) (int64, error) {
	return 0, nil
}

type flightStateRepositoryStub struct {
	states []flightstate.FlightState
	err    error
}

func (repository *flightStateRepositoryStub) ListByReconciliationScope(
	context.Context,
	string,
	string,
	time.Time,
	time.Time,
) ([]flightstate.FlightState, error) {
	return repository.states, repository.err
}

type dataQualityRepositoryStub struct {
	saveCount         int
	saveErr           error
	savedTaskID       string
	savedAttemptCount int
}

func (repository *dataQualityRepositoryStub) SaveReconciledFlightStateQuality(
	_ context.Context,
	taskID string,
	attemptCount int,
	_ flightstate.FlightState,
	_ dataquality.DataQuality,
) error {
	repository.saveCount++
	repository.savedTaskID = taskID
	repository.savedAttemptCount = attemptCount

	return repository.saveErr
}

type trajectoryRepositoryStub struct {
	saveCount         int
	saveErr           error
	savedTaskID       string
	savedAttemptCount int
	savedTrajectory   trajectory.FlightTrajectory
}

func (repository *trajectoryRepositoryStub) SaveReconciledTrajectory(
	_ context.Context,
	taskID string,
	attemptCount int,
	item trajectory.FlightTrajectory,
) error {
	repository.saveCount++
	repository.savedTaskID = taskID
	repository.savedAttemptCount = attemptCount
	repository.savedTrajectory = item

	return repository.saveErr
}
