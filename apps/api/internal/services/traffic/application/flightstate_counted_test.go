package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
)

func TestProcessAndStoreUsesActualReplaySafeInsertCount(
	t *testing.T,
) {
	repository := &countedFlightStateRepositoryStub{
		insertedCount: 0,
	}
	service, err := New(Config{
		Processor:             newFixedProcessor(t),
		FlightStateRepository: repository,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	result, err := service.ProcessAndStore(
		context.Background(),
		[]flightstate.FlightState{
			makeApplicationFlightState(
				"state-replayed",
				"ABC123",
				"AHY101",
				fixedNow().Add(-time.Minute),
			),
		},
	)
	if err != nil {
		t.Fatalf("process replayed observation: %v", err)
	}
	if result.StoredFlightStateCount != 0 {
		t.Fatalf(
			"stored count = %d, want 0 for replay conflict",
			result.StoredFlightStateCount,
		)
	}
	if repository.countedCalls != 1 || repository.legacyCalls != 0 {
		t.Fatalf(
			"counted_calls=%d legacy_calls=%d, want 1 and 0",
			repository.countedCalls,
			repository.legacyCalls,
		)
	}
}

func TestProcessAndStorePropagatesCountedRepositoryError(
	t *testing.T,
) {
	expectedErr := errors.New("counted persistence failed")
	repository := &countedFlightStateRepositoryStub{
		err: expectedErr,
	}
	service, err := New(Config{
		Processor:             newFixedProcessor(t),
		FlightStateRepository: repository,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	_, err = service.ProcessAndStore(
		context.Background(),
		[]flightstate.FlightState{
			makeApplicationFlightState(
				"state-1",
				"ABC123",
				"AHY101",
				fixedNow().Add(-time.Minute),
			),
		},
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected counted repository error, got %v", err)
	}
}

type countedFlightStateRepositoryStub struct {
	insertedCount int
	err           error
	countedCalls  int
	legacyCalls   int
}

func (repository *countedFlightStateRepositoryStub) SaveFlightStates(
	context.Context,
	[]flightstate.FlightState,
) error {
	repository.legacyCalls++
	return nil
}

func (repository *countedFlightStateRepositoryStub) SaveFlightStatesCounted(
	context.Context,
	[]flightstate.FlightState,
) (int, error) {
	repository.countedCalls++
	return repository.insertedCount, repository.err
}
