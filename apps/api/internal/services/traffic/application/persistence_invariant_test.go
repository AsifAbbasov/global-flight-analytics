package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestProcessAndStorePreservesDurableFlightStatesWhenTrajectoryPersistenceFails(
	t *testing.T,
) {
	trajectoryErr := errors.New(
		"trajectory persistence failed",
	)

	flightStateRepository := &fakeFlightStateRepository{}
	trajectoryRepository := &fakeTrajectoryRepository{
		err: trajectoryErr,
	}

	service := mustNewService(
		t,
		Config{
			Processor:             newFixedProcessor(t),
			FlightStateRepository: flightStateRepository,
			TrajectoryRepository:  trajectoryRepository,
		},
	)

	result, err := service.ProcessAndStore(
		context.Background(),
		[]flightstate.FlightState{
			makeApplicationFlightState(
				"state-1",
				"ABC123",
				"AHY101",
				fixedNow().Add(-60*time.Second),
			),
		},
	)

	if !errors.Is(
		err,
		trajectoryErr,
	) {
		t.Fatalf(
			"expected trajectory persistence error, got %v",
			err,
		)
	}

	if result.StoredFlightStateCount != 1 {
		t.Fatalf(
			"expected 1 durable flight state, got %d",
			result.StoredFlightStateCount,
		)
	}

	if result.StoredTrajectoryCount != 0 {
		t.Fatalf(
			"expected 0 stored trajectories, got %d",
			result.StoredTrajectoryCount,
		)
	}

	if flightStateRepository.saveCount != 1 {
		t.Fatalf(
			"expected flight state batch to remain committed, got %d saves",
			flightStateRepository.saveCount,
		)
	}

	if !result.StoredAt.IsZero() {
		t.Fatalf(
			"expected incomplete operation to keep StoredAt zero, got %s",
			result.StoredAt,
		)
	}
}

func TestProcessAndStoreReportsPartialQualityPersistenceProgress(
	t *testing.T,
) {
	qualityErr := errors.New(
		"quality persistence failed",
	)

	flightStateRepository := &fakeFlightStateRepository{}
	qualityRepository := &failAfterDataQualityRepository{
		failAtCall: 2,
		err:        qualityErr,
	}
	trajectoryRepository := &fakeTrajectoryRepository{}

	service := mustNewService(
		t,
		Config{
			Processor:             newFixedProcessor(t),
			FlightStateRepository: flightStateRepository,
			DataQualityRepository: qualityRepository,
			TrajectoryRepository:  trajectoryRepository,
		},
	)

	result, err := service.ProcessAndStore(
		context.Background(),
		[]flightstate.FlightState{
			makeApplicationFlightState(
				"state-1",
				"ABC123",
				"AHY101",
				fixedNow().Add(-60*time.Second),
			),
			makeApplicationFlightState(
				"state-2",
				"ABC123",
				"AHY101",
				fixedNow().Add(-30*time.Second),
			),
		},
	)

	if !errors.Is(
		err,
		qualityErr,
	) {
		t.Fatalf(
			"expected quality persistence error, got %v",
			err,
		)
	}

	if result.StoredFlightStateCount != 2 {
		t.Fatalf(
			"expected 2 durable flight states, got %d",
			result.StoredFlightStateCount,
		)
	}

	if result.StoredQualityReportCount != 1 {
		t.Fatalf(
			"expected 1 stored quality report before failure, got %d",
			result.StoredQualityReportCount,
		)
	}

	if result.StoredTrajectoryCount != 0 {
		t.Fatalf(
			"expected trajectory persistence not to start, got %d stored trajectories",
			result.StoredTrajectoryCount,
		)
	}

	if trajectoryRepository.saveCount != 0 {
		t.Fatalf(
			"expected no trajectory repository calls, got %d",
			trajectoryRepository.saveCount,
		)
	}
}

func TestProcessAndStoreReportsSuccessfulDerivedPersistenceCounts(
	t *testing.T,
) {
	qualityRepository := &fakeDataQualityRepository{}
	trajectoryRepository := &fakeTrajectoryRepository{}

	service := mustNewService(
		t,
		Config{
			Processor:             newFixedProcessor(t),
			FlightStateRepository: &fakeFlightStateRepository{},
			DataQualityRepository: qualityRepository,
			TrajectoryRepository:  trajectoryRepository,
		},
	)

	result, err := service.ProcessAndStore(
		context.Background(),
		[]flightstate.FlightState{
			makeApplicationFlightState(
				"state-1",
				"ABC123",
				"AHY101",
				fixedNow().Add(-60*time.Second),
			),
			makeApplicationFlightState(
				"state-2",
				"ABC123",
				"AHY101",
				fixedNow().Add(-30*time.Second),
			),
		},
	)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.StoredFlightStateCount != 2 {
		t.Fatalf(
			"expected 2 stored flight states, got %d",
			result.StoredFlightStateCount,
		)
	}

	if result.StoredQualityReportCount != 2 {
		t.Fatalf(
			"expected 2 stored quality reports, got %d",
			result.StoredQualityReportCount,
		)
	}

	if result.StoredTrajectoryCount != 1 {
		t.Fatalf(
			"expected 1 stored trajectory, got %d",
			result.StoredTrajectoryCount,
		)
	}

	if result.StoredAt.IsZero() {
		t.Fatal(
			"expected completed persistence timestamp",
		)
	}
}

type failAfterDataQualityRepository struct {
	callCount  int
	failAtCall int
	err        error
}

func (repository *failAfterDataQualityRepository) SaveFlightStateQuality(
	context.Context,
	flightstate.FlightState,
	dataquality.DataQuality,
) error {
	repository.callCount++

	if repository.callCount == repository.failAtCall {
		return repository.err
	}

	return nil
}

var _ DataQualityRepository = (*failAfterDataQualityRepository)(nil)
