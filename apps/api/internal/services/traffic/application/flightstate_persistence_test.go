package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestProcessAndStoreSavesUsableFlightStates(t *testing.T) {
	repository := &recordingFlightStateRepository{}

	service := New(Config{
		Processor:             newFixedProcessor(),
		FlightStateRepository: repository,
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

	if repository.saveCount != 1 {
		t.Fatalf(
			"expected 1 batch save, got %d",
			repository.saveCount,
		)
	}

	if len(repository.lastStates) != 2 {
		t.Fatalf(
			"expected 2 saved states, got %d",
			len(repository.lastStates),
		)
	}
}

func TestProcessAndStoreReturnsFlightStateRepositoryError(t *testing.T) {
	service := New(Config{
		Processor: newFixedProcessor(),
		FlightStateRepository: &recordingFlightStateRepository{
			err: errors.New("storage failed"),
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
		t.Fatalf("unexpected error: %v", err)
	}
}

type recordingFlightStateRepository struct {
	saveCount  int
	lastStates []flightstate.FlightState
	err        error
}

func (repository *recordingFlightStateRepository) SaveFlightStates(
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
