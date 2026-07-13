package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/flightcontinuation"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
)

type FlightStateRepository interface {
	SaveFlightStates(
		ctx context.Context,
		items []flightstate.FlightState,
	) error
}

type TrajectoryRepository interface {
	SaveTrajectory(
		ctx context.Context,
		item trajectory.FlightTrajectory,
	) error
}

type TrajectoryContinuationRepository interface {
	GetLatestTrajectoryByICAO24(
		ctx context.Context,
		icao24 string,
	) (trajectory.FlightTrajectory, error)
}

type DataQualityRepository interface {
	SaveFlightStateQuality(
		ctx context.Context,
		state flightstate.FlightState,
		quality dataquality.DataQuality,
	) error
}

type ReconciliationRepository interface {
	MarkPendingDerivation(
		ctx context.Context,
		task reconciliation.PendingDerivation,
	) error
}

type Config struct {
	Processor                        *processor.Processor
	FlightStateRepository            FlightStateRepository
	TrajectoryRepository             TrajectoryRepository
	TrajectoryContinuationRepository TrajectoryContinuationRepository
	IdentityContinuationMaxGap       time.Duration
	DataQualityRepository            DataQualityRepository
	ReconciliationRepository         ReconciliationRepository
}

type Service struct {
	processor                        *processor.Processor
	flightStateRepository            FlightStateRepository
	trajectoryRepository             TrajectoryRepository
	trajectoryContinuationRepository TrajectoryContinuationRepository
	identityContinuationConfig       flightcontinuation.Config
	dataQualityRepository            DataQualityRepository
	reconciliationRepository         ReconciliationRepository
}

type ProcessAndStoreResult struct {
	ProcessingResult         processor.ProcessingResult
	ContinuedTrajectoryCount int
	StoredFlightStateCount   int
	StoredQualityReportCount int
	StoredTrajectoryCount    int
	StoredAt                 time.Time
}

func New(
	config Config,
) (*Service, error) {
	trafficProcessor := config.Processor

	if trafficProcessor == nil {
		var err error

		trafficProcessor, err = processor.New(
			processor.Config{},
		)
		if err != nil {
			return nil, fmt.Errorf(
				"create default traffic processor: %w",
				err,
			)
		}
	}

	identityContinuationConfig := flightcontinuation.Config{
		MaxGap: config.IdentityContinuationMaxGap,
	}
	if err := identityContinuationConfig.Validate(); err != nil {
		return nil, fmt.Errorf(
			"validate flight identity continuation config: %w",
			err,
		)
	}

	return &Service{
		processor:                        trafficProcessor,
		flightStateRepository:            config.FlightStateRepository,
		trajectoryRepository:             config.TrajectoryRepository,
		trajectoryContinuationRepository: config.TrajectoryContinuationRepository,
		identityContinuationConfig:       identityContinuationConfig,
		dataQualityRepository:            config.DataQualityRepository,
		reconciliationRepository:         config.ReconciliationRepository,
	}, nil
}

// ProcessAndStore intentionally persists independent durability units in order.
//
// Flight states are source observations. Once their repository commits them,
// failures in derived quality reports or trajectories do not roll them back.
// Recoverable derived write failures are recorded as reconciliation tasks.
// Each derived persistence stage processes all items in that stage, but a
// failed stage stops later stages from starting. StoredAt is assigned only
// after every configured persistence stage completes successfully.
// Each repository owns atomicity for its own batch or aggregate.
func (service *Service) ProcessAndStore(
	ctx context.Context,
	states []flightstate.FlightState,
) (ProcessAndStoreResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	processingResult := service.processor.Process(states)

	continuedTrajectoryCount, continuationErr :=
		service.applyFlightIdentityContinuations(
			ctx,
			&processingResult,
		)

	result := ProcessAndStoreResult{
		ProcessingResult:         processingResult,
		ContinuedTrajectoryCount: continuedTrajectoryCount,
	}
	if continuationErr != nil {
		return result, continuationErr
	}

	if service.flightStateRepository != nil {
		storedFlightStateCount, err := service.saveUsableFlightStates(
			ctx,
			processingResult,
		)
		if err != nil {
			return result, err
		}

		result.StoredFlightStateCount = storedFlightStateCount
	}

	if service.dataQualityRepository != nil {
		storedQualityReportCount, err := service.saveDataQualityReports(
			ctx,
			processingResult,
		)

		result.StoredQualityReportCount = storedQualityReportCount

		if err != nil {
			return result, err
		}
	}

	if service.trajectoryRepository != nil {
		storedTrajectoryCount, err := service.saveTrajectories(
			ctx,
			processingResult,
		)

		result.StoredTrajectoryCount = storedTrajectoryCount

		if err != nil {
			return result, err
		}
	}

	result.StoredAt = time.Now().UTC()

	return result, nil
}

func (service *Service) saveUsableFlightStates(
	ctx context.Context,
	result processor.ProcessingResult,
) (int, error) {
	states := make(
		[]flightstate.FlightState,
		0,
		len(result.UsableStates),
	)

	for _, item := range result.UsableStates {
		states = append(states, item.State)
	}

	if err := service.flightStateRepository.SaveFlightStates(
		ctx,
		states,
	); err != nil {
		return 0, fmt.Errorf(
			"save usable flight states: %w",
			err,
		)
	}

	return len(states), nil
}

func (service *Service) saveDataQualityReports(
	ctx context.Context,
	result processor.ProcessingResult,
) (int, error) {
	storedCount := 0
	saveErrors := make(
		[]error,
		0,
	)

	for _, item := range result.UsableStates {
		err := service.dataQualityRepository.SaveFlightStateQuality(
			ctx,
			item.State,
			item.Quality,
		)
		if err == nil {
			storedCount++
			continue
		}

		saveErrors = append(
			saveErrors,
			fmt.Errorf(
				"save usable flight state quality report for icao24 %s: %w",
				item.State.ICAO24,
				err,
			),
		)

		if markErr := service.markPendingFlightStateQuality(
			ctx,
			item.State,
			err,
		); markErr != nil {
			saveErrors = append(
				saveErrors,
				markErr,
			)
		}
	}

	// Invalid states are not persisted in flight_states, so they cannot be
	// reconstructed by a worker that uses flight_states as its durable source.
	// Their persistence failures stay visible in the aggregate error instead of
	// creating pending tasks that can never be completed.
	for _, item := range result.InvalidStates {
		err := service.dataQualityRepository.SaveFlightStateQuality(
			ctx,
			item.State,
			item.Quality,
		)
		if err == nil {
			storedCount++
			continue
		}

		saveErrors = append(
			saveErrors,
			fmt.Errorf(
				"save invalid flight state quality report for icao24 %s: %w",
				item.State.ICAO24,
				err,
			),
		)
	}

	return storedCount, errors.Join(saveErrors...)
}

func (service *Service) saveTrajectories(
	ctx context.Context,
	result processor.ProcessingResult,
) (int, error) {
	storedCount := 0
	saveErrors := make(
		[]error,
		0,
	)
	latestStates := latestUsableStatesByICAO24(
		result.UsableStates,
	)

	for collectionKey, item := range result.Trajectories {
		icao24 := item.ICAO24
		if icao24 == "" {
			icao24 = collectionKey
		}

		err := service.trajectoryRepository.SaveTrajectory(
			ctx,
			item,
		)
		if err == nil {
			storedCount++
			continue
		}

		saveErrors = append(
			saveErrors,
			fmt.Errorf(
				"save trajectory for icao24 %s: %w",
				icao24,
				err,
			),
		)

		if markErr := service.markPendingTrajectory(
			ctx,
			icao24,
			item,
			latestStates[icao24],
			err,
		); markErr != nil {
			saveErrors = append(
				saveErrors,
				markErr,
			)
		}
	}

	return storedCount, errors.Join(saveErrors...)
}

func (service *Service) markPendingFlightStateQuality(
	ctx context.Context,
	state flightstate.FlightState,
	cause error,
) error {
	if service.reconciliationRepository == nil {
		return nil
	}

	err := service.reconciliationRepository.MarkPendingDerivation(
		ctx,
		reconciliation.PendingDerivation{
			IngestionRunID: state.IngestionRunID,
			ICAO24:         state.ICAO24,
			DerivationType: reconciliation.DerivationTypeFlightStateQuality,
			ObservedFrom:   state.ObservedAt,
			ObservedTo:     state.ObservedAt,
			LastError:      cause.Error(),
		},
	)
	if err != nil {
		return fmt.Errorf(
			"mark pending flight state quality derivation for icao24 %s after save failure: %w",
			state.ICAO24,
			err,
		)
	}

	return nil
}

func (service *Service) markPendingTrajectory(
	ctx context.Context,
	icao24 string,
	item trajectory.FlightTrajectory,
	latestState flightstate.FlightState,
	cause error,
) error {
	if service.reconciliationRepository == nil {
		return nil
	}

	observedFrom := item.StartTime
	observedTo := item.EndTime
	if observedFrom.IsZero() {
		observedFrom = latestState.ObservedAt
	}
	if observedTo.IsZero() {
		observedTo = latestState.ObservedAt
	}

	ingestionRunID := latestState.IngestionRunID

	err := service.reconciliationRepository.MarkPendingDerivation(
		ctx,
		reconciliation.PendingDerivation{
			IngestionRunID: ingestionRunID,
			ICAO24:         icao24,
			DerivationType: reconciliation.DerivationTypeTrajectory,
			ObservedFrom:   observedFrom,
			ObservedTo:     observedTo,
			LastError:      cause.Error(),
		},
	)
	if err != nil {
		return fmt.Errorf(
			"mark pending trajectory derivation for icao24 %s after save failure: %w",
			icao24,
			err,
		)
	}

	return nil
}

func latestUsableStatesByICAO24(
	states []processor.ProcessedFlightState,
) map[string]flightstate.FlightState {
	result := make(
		map[string]flightstate.FlightState,
		len(states),
	)

	for _, item := range states {
		current, exists := result[item.State.ICAO24]
		if !exists || item.State.ObservedAt.After(current.ObservedAt) {
			result[item.State.ICAO24] = item.State
		}
	}

	return result
}
