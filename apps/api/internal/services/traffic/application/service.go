package application

import (
	"context"
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
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

type DataQualityRepository interface {
	SaveFlightStateQuality(
		ctx context.Context,
		state flightstate.FlightState,
		quality dataquality.DataQuality,
	) error
}

type Config struct {
	Processor             *processor.Processor
	FlightStateRepository FlightStateRepository
	TrajectoryRepository  TrajectoryRepository
	DataQualityRepository DataQualityRepository
}

type Service struct {
	processor             *processor.Processor
	flightStateRepository FlightStateRepository
	trajectoryRepository  TrajectoryRepository
	dataQualityRepository DataQualityRepository
}

type ProcessAndStoreResult struct {
	ProcessingResult processor.ProcessingResult
	StoredAt         time.Time
}

func New(config Config) *Service {
	trafficProcessor := config.Processor

	if trafficProcessor == nil {
		trafficProcessor = processor.New(processor.Config{})
	}

	return &Service{
		processor:             trafficProcessor,
		flightStateRepository: config.FlightStateRepository,
		trajectoryRepository:  config.TrajectoryRepository,
		dataQualityRepository: config.DataQualityRepository,
	}
}

func (service *Service) ProcessAndStore(
	ctx context.Context,
	states []flightstate.FlightState,
) (ProcessAndStoreResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	result := service.processor.Process(states)

	if service.flightStateRepository != nil {
		if err := service.saveUsableFlightStates(ctx, result); err != nil {
			return ProcessAndStoreResult{}, err
		}
	}

	if service.dataQualityRepository != nil {
		if err := service.saveDataQualityReports(ctx, result); err != nil {
			return ProcessAndStoreResult{}, err
		}
	}

	if service.trajectoryRepository != nil {
		if err := service.saveTrajectories(
			ctx,
			result.Trajectories,
		); err != nil {
			return ProcessAndStoreResult{}, err
		}
	}

	return ProcessAndStoreResult{
		ProcessingResult: result,
		StoredAt:         time.Now().UTC(),
	}, nil
}

func (service *Service) saveUsableFlightStates(
	ctx context.Context,
	result processor.ProcessingResult,
) error {
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
		return fmt.Errorf(
			"save usable flight states: %w",
			err,
		)
	}

	return nil
}

func (service *Service) saveDataQualityReports(
	ctx context.Context,
	result processor.ProcessingResult,
) error {
	for _, item := range result.UsableStates {
		if err := service.dataQualityRepository.SaveFlightStateQuality(
			ctx,
			item.State,
			item.Quality,
		); err != nil {
			return fmt.Errorf(
				"save usable flight state quality report for icao24 %s: %w",
				item.State.ICAO24,
				err,
			)
		}
	}

	for _, item := range result.InvalidStates {
		if err := service.dataQualityRepository.SaveFlightStateQuality(
			ctx,
			item.State,
			item.Quality,
		); err != nil {
			return fmt.Errorf(
				"save invalid flight state quality report for icao24 %s: %w",
				item.State.ICAO24,
				err,
			)
		}
	}

	return nil
}

func (service *Service) saveTrajectories(
	ctx context.Context,
	trajectories map[string]trajectory.FlightTrajectory,
) error {
	for icao24, item := range trajectories {
		if err := service.trajectoryRepository.SaveTrajectory(
			ctx,
			item,
		); err != nil {
			return fmt.Errorf(
				"save trajectory for icao24 %s: %w",
				icao24,
				err,
			)
		}
	}

	return nil
}
