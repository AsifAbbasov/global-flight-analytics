package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/dataquality"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/trajectory"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
)

func TestProcessAndStoreWithoutRepositories(t *testing.T) {
	service := New(Config{
		Processor: newFixedProcessor(),
	})

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
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.ProcessingResult.Stats.ReceivedCount != 1 {
		t.Fatalf(
			"expected 1 received state, got %d",
			result.ProcessingResult.Stats.ReceivedCount,
		)
	}

	if result.ProcessingResult.Stats.TrajectoryCount != 1 {
		t.Fatalf(
			"expected 1 trajectory, got %d",
			result.ProcessingResult.Stats.TrajectoryCount,
		)
	}

	if result.StoredAt.IsZero() {
		t.Fatal("expected stored at timestamp")
	}
}

func TestProcessAndStoreSavesFlightStatesDataQualityAndTrajectories(
	t *testing.T,
) {
	flightStateRepository := &fakeFlightStateRepository{}
	trajectoryRepository := &fakeTrajectoryRepository{}
	dataQualityRepository := &fakeDataQualityRepository{}

	service := New(Config{
		Processor:             newFixedProcessor(),
		FlightStateRepository: flightStateRepository,
		TrajectoryRepository:  trajectoryRepository,
		DataQualityRepository: dataQualityRepository,
	})

	_, err := service.ProcessAndStore(
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
		t.Fatalf("expected no error, got %v", err)
	}

	if flightStateRepository.saveCount != 1 {
		t.Fatalf(
			"expected 1 flight state batch save, got %d",
			flightStateRepository.saveCount,
		)
	}

	if len(flightStateRepository.lastStates) != 2 {
		t.Fatalf(
			"expected 2 usable flight states, got %d",
			len(flightStateRepository.lastStates),
		)
	}

	if dataQualityRepository.saveCount != 2 {
		t.Fatalf(
			"expected 2 data quality reports, got %d",
			dataQualityRepository.saveCount,
		)
	}

	if trajectoryRepository.saveCount != 1 {
		t.Fatalf(
			"expected 1 trajectory, got %d",
			trajectoryRepository.saveCount,
		)
	}

	if trajectoryRepository.lastTrajectory.ICAO24 != "ABC123" {
		t.Fatalf(
			"expected trajectory for ABC123, got %s",
			trajectoryRepository.lastTrajectory.ICAO24,
		)
	}
}

func TestProcessAndStoreReturnsFlightStateRepositoryError(t *testing.T) {
	expectedError := errors.New("flight state storage failed")

	service := New(Config{
		Processor: newFixedProcessor(),
		FlightStateRepository: &fakeFlightStateRepository{
			err: expectedError,
		},
	})

	_, err := service.ProcessAndStore(
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

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "save usable flight states") {
		t.Fatalf(
			"expected contextual flight state error, got %v",
			err,
		)
	}
}

func TestProcessAndStoreReturnsDataQualityRepositoryError(t *testing.T) {
	expectedError := errors.New("data quality storage failed")

	service := New(Config{
		Processor: newFixedProcessor(),
		DataQualityRepository: &fakeDataQualityRepository{
			err: expectedError,
		},
	})

	_, err := service.ProcessAndStore(
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

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(
		err.Error(),
		"save usable flight state quality report",
	) {
		t.Fatalf(
			"expected contextual data quality error, got %v",
			err,
		)
	}
}

func TestProcessAndStoreReturnsTrajectoryRepositoryError(t *testing.T) {
	expectedError := errors.New("trajectory storage failed")

	service := New(Config{
		Processor: newFixedProcessor(),
		TrajectoryRepository: &fakeTrajectoryRepository{
			err: expectedError,
		},
	})

	_, err := service.ProcessAndStore(
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

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(
		err.Error(),
		"save trajectory for icao24 ABC123",
	) {
		t.Fatalf(
			"expected contextual trajectory error, got %v",
			err,
		)
	}
}

func newFixedProcessor() *processor.Processor {
	now := fixedNow()

	return processor.New(processor.Config{
		Now: func() time.Time {
			return now
		},
	})
}

func fixedNow() time.Time {
	return time.Date(
		2026,
		time.July,
		2,
		10,
		0,
		0,
		0,
		time.UTC,
	)
}

func makeApplicationFlightState(
	id string,
	icao24 string,
	callsign string,
	observedAt time.Time,
) flightstate.FlightState {
	return flightstate.FlightState{
		ID:                  id,
		FlightID:            "flight-" + icao24,
		AircraftID:          "aircraft-" + icao24,
		ICAO24:              icao24,
		Callsign:            callsign,
		Latitude:            40.4093,
		Longitude:           49.8671,
		BarometricAltitudeM: 10000,
		GeometricAltitudeM:  10050,
		VelocityMPS:         230,
		HeadingDegrees:      90,
		VerticalRateMPS:     0,
		OnGround:            false,
		OriginCountry:       "Azerbaijan",
		ObservedAt:          observedAt,
		SourceName:          "test",
	}
}

type fakeFlightStateRepository struct {
	saveCount  int
	lastStates []flightstate.FlightState
	err        error
}

func (repository *fakeFlightStateRepository) SaveFlightStates(
	ctx context.Context,
	items []flightstate.FlightState,
) error {
	if repository.err != nil {
		return repository.err
	}

	repository.saveCount++

	repository.lastStates = append(
		[]flightstate.FlightState(nil),
		items...,
	)

	return nil
}

type fakeTrajectoryRepository struct {
	saveCount      int
	lastTrajectory trajectory.FlightTrajectory
	err            error
}

func (repository *fakeTrajectoryRepository) SaveTrajectory(
	ctx context.Context,
	item trajectory.FlightTrajectory,
) error {
	if repository.err != nil {
		return repository.err
	}

	repository.saveCount++
	repository.lastTrajectory = item

	return nil
}

type fakeDataQualityRepository struct {
	saveCount   int
	lastState   flightstate.FlightState
	lastQuality dataquality.DataQuality
	err         error
}

func (repository *fakeDataQualityRepository) SaveFlightStateQuality(
	ctx context.Context,
	state flightstate.FlightState,
	quality dataquality.DataQuality,
) error {
	if repository.err != nil {
		return repository.err
	}

	repository.saveCount++
	repository.lastState = state
	repository.lastQuality = quality

	return nil
}
