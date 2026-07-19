package server

import (
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/weatherprovider"
	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
)

func composeWeatherProvider(
	openMeteoTimeout time.Duration,
) (
	weatherservice.CurrentWeatherClient,
	error,
) {
	controller, err :=
		composeWeatherResponseController()
	if err != nil {
		return nil, err
	}

	observer, err :=
		composeWeatherResponseObserver(
			controller,
		)
	if err != nil {
		return nil, err
	}

	orchestrator, err :=
		composeWeatherOrchestrator(
			controller,
		)
	if err != nil {
		return nil, err
	}

	openMeteoClient, err :=
		composeOpenMeteoClient(
			openMeteoTimeout,
			observer,
		)
	if err != nil {
		return nil, err
	}

	return composeOrchestratedWeatherClient(
		openMeteoClient,
		orchestrator,
	)
}

func composeWeatherResponseController() (
	*providerresponse.Controller,
	error,
) {
	budgetManager, err := providerbudget.New(nil)
	if err != nil {
		return nil, fmt.Errorf(
			"initialize provider budget manager: %w",
			err,
		)
	}

	controller, err := providerresponse.New(
		providerresponse.Config{
			BudgetManager: budgetManager,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"initialize provider response controller: %w",
			err,
		)
	}

	return controller, nil
}

func composeWeatherResponseObserver(
	controller *providerresponse.Controller,
) (
	*providerresponse.IntegrationObserver,
	error,
) {
	observer, err :=
		providerresponse.
			NewIntegrationObserver(
				controller,
			)
	if err != nil {
		return nil, fmt.Errorf(
			"initialize provider response observer: %w",
			err,
		)
	}

	return observer, nil
}

func composeWeatherOrchestrator(
	controller *providerresponse.Controller,
) (
	*ingestionorchestrator.Orchestrator[weatherprovider.ExecutionValue],
	error,
) {
	orchestrator, err :=
		ingestionorchestrator.
			NewDefault[weatherprovider.ExecutionValue](
			controller,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"initialize ingestion orchestrator: %w",
			err,
		)
	}

	return orchestrator, nil
}

func composeOpenMeteoClient(
	timeout time.Duration,
	observer *providerresponse.IntegrationObserver,
) (*openmeteo.Client, error) {
	client, err := openmeteo.New(
		openmeteo.Config{
			Timeout:          timeout,
			ResponseObserver: observer,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"initialize open-meteo client: %w",
			err,
		)
	}

	return client, nil
}

func composeOrchestratedWeatherClient(
	client weatherprovider.Delegate,
	executor weatherprovider.Executor,
) (*weatherprovider.Client, error) {
	orchestratedClient, err :=
		weatherprovider.New(
			weatherprovider.Config{
				Client:   client,
				Executor: executor,
			},
		)
	if err != nil {
		return nil, fmt.Errorf(
			"initialize orchestrated weather client: %w",
			err,
		)
	}

	return orchestratedClient, nil
}
