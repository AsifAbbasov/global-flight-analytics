package server

import (
	"fmt"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/weatherprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerWeatherRoute(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
	openMeteoTimeout time.Duration,
) error {
	if openMeteoTimeout <= 0 {
		return fmt.Errorf(
			"open-meteo timeout must be greater than zero",
		)
	}

	budgetManager, err := providerbudget.New(nil)
	if err != nil {
		return fmt.Errorf(
			"initialize provider budget manager: %w",
			err,
		)
	}

	responseController, err := providerresponse.New(
		providerresponse.Config{
			BudgetManager: budgetManager,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"initialize provider response controller: %w",
			err,
		)
	}

	responseObserver, err := providerresponse.NewIntegrationObserver(
		responseController,
	)
	if err != nil {
		return fmt.Errorf(
			"initialize provider response observer: %w",
			err,
		)
	}

	orchestrator, err := ingestionorchestrator.NewDefault[weatherprovider.ExecutionValue](
		responseController,
	)
	if err != nil {
		return fmt.Errorf(
			"initialize ingestion orchestrator: %w",
			err,
		)
	}

	openMeteoClient, err := openmeteo.New(
		openmeteo.Config{
			Timeout:          openMeteoTimeout,
			ResponseObserver: responseObserver,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"initialize open-meteo client: %w",
			err,
		)
	}

	orchestratedWeatherClient, err := weatherprovider.New(
		weatherprovider.Config{
			Client:   openMeteoClient,
			Executor: orchestrator,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"initialize orchestrated weather client: %w",
			err,
		)
	}

	weatherRepository := postgres.NewWeatherRepository(dbPool)
	weatherService := weatherservice.New(
		weatherservice.Config{
			Client:     orchestratedWeatherClient,
			Repository: weatherRepository,
		},
	)
	weatherHandler := handlers.NewWeatherHandler(weatherService)

	v1.Get(
		"/weather/current",
		weatherHandler.GetCurrent,
	)

	return nil
}
