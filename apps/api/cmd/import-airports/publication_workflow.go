package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/ourairports"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
)

const ourAirportsPublicationRequestKey = "airports:regional-import"

type airportImportExecutionValue struct {
	ReconciledCount int64
}

func (airportImportExecutionValue) RequestCoalescingValue() {}

type airportPublicationExecutor interface {
	ExecutePublication(
		ctx context.Context,
		provider providerpolicy.Provider,
		requestKey string,
		publicationID string,
		function ingestionorchestrator.Function[airportImportExecutionValue],
	) (ingestionorchestrator.ExecuteResult[airportImportExecutionValue], error)
}

type airportImportRepository interface {
	UpsertImported(
		ctx context.Context,
		records []airport.ImportRecord,
	) (int64, error)
}

type airportPublicationOutcome struct {
	ReconciledCount  int64
	AlreadyProcessed bool
	Shared           bool
}

func executeAirportPublication(
	ctx context.Context,
	executor airportPublicationExecutor,
	repository airportImportRepository,
	result ourairports.LoadResult,
) (airportPublicationOutcome, error) {
	if ctx == nil {
		return airportPublicationOutcome{},
			fmt.Errorf("airport publication context is required")
	}
	if executor == nil {
		return airportPublicationOutcome{},
			fmt.Errorf("airport publication executor is required")
	}
	if repository == nil {
		return airportPublicationOutcome{},
			fmt.Errorf("airport import repository is required")
	}

	publicationID := strings.TrimSpace(result.PublicationID)
	if publicationID == "" {
		return airportPublicationOutcome{},
			fmt.Errorf("OurAirports publication identifier is required")
	}

	executionResult, err := executor.ExecutePublication(
		ctx,
		providerpolicy.ProviderOurAirports,
		ourAirportsPublicationRequestKey,
		publicationID,
		func(operationContext context.Context) (airportImportExecutionValue, error) {
			reconciledCount, importErr := repository.UpsertImported(
				operationContext,
				result.Airports,
			)
			if importErr != nil {
				return airportImportExecutionValue{}, fmt.Errorf(
					"reconcile OurAirports airports: %w",
					importErr,
				)
			}
			return airportImportExecutionValue{
				ReconciledCount: reconciledCount,
			}, nil
		},
	)
	if err != nil {
		var accessDenied *ingestionorchestrator.AccessDeniedError
		if errors.As(err, &accessDenied) &&
			accessDenied.Reason == providerbudget.DecisionReasonPublicationAlreadyProcessed {
			return airportPublicationOutcome{
				AlreadyProcessed: true,
			}, nil
		}
		return airportPublicationOutcome{}, err
	}

	return airportPublicationOutcome{
		ReconciledCount: executionResult.Value.ReconciledCount,
		Shared:          executionResult.Shared,
	}, nil
}
