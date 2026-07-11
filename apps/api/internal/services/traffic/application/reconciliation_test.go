package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/reconciliation"
)

func TestProcessAndStoreMarksPendingTrajectoryDerivationWhenTrajectoryPersistenceFails(
	t *testing.T,
) {
	trajectoryErr := errors.New(
		"trajectory persistence failed",
	)
	reconciliationRepository := &fakeReconciliationRepository{}

	service := mustNewService(
		t,
		Config{
			Processor:             newFixedProcessor(t),
			FlightStateRepository: &fakeFlightStateRepository{},
			TrajectoryRepository: &fakeTrajectoryRepository{
				err: trajectoryErr,
			},
			ReconciliationRepository: reconciliationRepository,
		},
	)

	_, err := service.ProcessAndStore(
		context.Background(),
		[]flightstate.FlightState{
			makeApplicationFlightStateWithIngestionRun(
				"state-1",
				"ABC123",
				"AHY101",
				fixedNow().Add(-60*time.Second),
				"550e8400-e29b-41d4-a716-446655440000",
			),
			makeApplicationFlightStateWithIngestionRun(
				"state-2",
				"ABC123",
				"AHY101",
				fixedNow().Add(-30*time.Second),
				"550e8400-e29b-41d4-a716-446655440000",
			),
		},
	)
	if err == nil {
		t.Fatal(
			"expected trajectory persistence error",
		)
	}

	if reconciliationRepository.saveCount != 1 {
		t.Fatalf(
			"expected 1 pending reconciliation task, got %d",
			reconciliationRepository.saveCount,
		)
	}

	task := reconciliationRepository.lastTask
	if task.DerivationType != reconciliation.DerivationTypeTrajectory {
		t.Fatalf(
			"expected trajectory derivation type, got %s",
			task.DerivationType,
		)
	}

	if task.ICAO24 != "ABC123" {
		t.Fatalf(
			"expected icao24 ABC123, got %s",
			task.ICAO24,
		)
	}

	if task.IngestionRunID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf(
			"expected ingestion run id to be copied, got %s",
			task.IngestionRunID,
		)
	}

	if !strings.Contains(
		task.LastError,
		trajectoryErr.Error(),
	) {
		t.Fatalf(
			"expected last error to contain original failure, got %s",
			task.LastError,
		)
	}

	if task.ObservedFrom.IsZero() || task.ObservedTo.IsZero() {
		t.Fatal(
			"expected observed time bounds for trajectory reconciliation task",
		)
	}
}

func TestProcessAndStoreMarksPendingQualityDerivationWhenQualityPersistenceFails(
	t *testing.T,
) {
	qualityErr := errors.New(
		"quality persistence failed",
	)
	reconciliationRepository := &fakeReconciliationRepository{}

	service := mustNewService(
		t,
		Config{
			Processor: newFixedProcessor(t),
			DataQualityRepository: &fakeDataQualityRepository{
				err: qualityErr,
			},
			ReconciliationRepository: reconciliationRepository,
		},
	)

	_, err := service.ProcessAndStore(
		context.Background(),
		[]flightstate.FlightState{
			makeApplicationFlightStateWithIngestionRun(
				"state-1",
				"ABC123",
				"AHY101",
				fixedNow().Add(-60*time.Second),
				"550e8400-e29b-41d4-a716-446655440000",
			),
		},
	)
	if err == nil {
		t.Fatal(
			"expected quality persistence error",
		)
	}

	if reconciliationRepository.saveCount != 1 {
		t.Fatalf(
			"expected 1 pending reconciliation task, got %d",
			reconciliationRepository.saveCount,
		)
	}

	task := reconciliationRepository.lastTask
	if task.DerivationType != reconciliation.DerivationTypeFlightStateQuality {
		t.Fatalf(
			"expected flight state quality derivation type, got %s",
			task.DerivationType,
		)
	}

	if task.ICAO24 != "ABC123" {
		t.Fatalf(
			"expected icao24 ABC123, got %s",
			task.ICAO24,
		)
	}

	if !strings.Contains(
		task.LastError,
		qualityErr.Error(),
	) {
		t.Fatalf(
			"expected last error to contain original failure, got %s",
			task.LastError,
		)
	}
}

func makeApplicationFlightStateWithIngestionRun(
	id string,
	icao24 string,
	callsign string,
	observedAt time.Time,
	ingestionRunID string,
) flightstate.FlightState {
	state := makeApplicationFlightState(
		id,
		icao24,
		callsign,
		observedAt,
	)
	state.IngestionRunID = ingestionRunID

	return state
}

type fakeReconciliationRepository struct {
	saveCount int
	lastTask  reconciliation.PendingDerivation
	err       error
}

func (repository *fakeReconciliationRepository) MarkPendingDerivation(
	ctx context.Context,
	task reconciliation.PendingDerivation,
) error {
	if repository.err != nil {
		return repository.err
	}

	repository.saveCount++
	repository.lastTask = task

	return nil
}
