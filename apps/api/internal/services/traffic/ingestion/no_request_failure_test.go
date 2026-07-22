package ingestion

import (
	"context"
	"errors"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/ingestionrun"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

func TestLocalProviderDenialDoesNotCreateFailedIngestionRun(
	t *testing.T,
) {
	runRepository := &testIngestionRunRepository{
		run: ingestionrun.Run{
			ID: "must-not-be-created",
		},
	}
	service := New(
		Config{
			Provider: &testRegionalProvider{
				sourceName: "airplanes.live",
				err: &ingestionorchestrator.AccessDeniedError{
					Provider: providerpolicy.ProviderAirplanesLive,
					Reason: providerbudget.
						DecisionReasonFixedWindowExhausted,
				},
			},
			ProcessingService:      &testProcessingService{},
			IngestionRunRepository: runRepository,
			Now:                    fixedIngestionTime,
		},
	)

	result, err := service.LoadAndProcessByPoint(
		context.Background(),
		40.4093,
		49.8671,
		100,
	)
	if err == nil {
		t.Fatal(
			"expected local provider denial",
		)
	}
	var denied *ingestionorchestrator.AccessDeniedError
	if !errors.As(err, &denied) {
		t.Fatalf(
			"expected access denied error, got %v",
			err,
		)
	}
	if result.IngestionRunID != "" {
		t.Fatalf(
			"ingestion run id = %q, want empty",
			result.IngestionRunID,
		)
	}
	if runRepository.createCount != 0 ||
		runRepository.failedCount != 0 {
		t.Fatalf(
			"repository writes create=%d failed=%d, want 0 and 0",
			runRepository.createCount,
			runRepository.failedCount,
		)
	}
}
