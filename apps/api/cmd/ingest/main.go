package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/airplaneslive"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/regionalprovider"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
	trafficingestion "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/ingestion"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal(
			"DATABASE_URL is required",
		)
	}

	latitude := mustFloat64Env(
		"TRAFFIC_INGESTION_LATITUDE",
	)

	longitude := mustFloat64Env(
		"TRAFFIC_INGESTION_LONGITUDE",
	)

	radius := mustIntEnv(
		"TRAFFIC_INGESTION_RADIUS",
	)

	timeout := mustDurationEnv(
		"AIRPLANES_LIVE_TIMEOUT",
	)

	dbPool, err := database.NewPostgresPool(
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatalf(
			"connect postgres: %v",
			err,
		)
	}
	defer dbPool.Close()

	budgetManager, err := providerbudget.New(
		nil,
	)
	if err != nil {
		log.Fatalf(
			"create provider budget manager: %v",
			err,
		)
	}

	responseController, err := providerresponse.New(
		providerresponse.Config{
			BudgetManager: budgetManager,
		},
	)
	if err != nil {
		log.Fatalf(
			"create provider response controller: %v",
			err,
		)
	}

	responseObserver, err := providerresponse.NewIntegrationObserver(
		responseController,
	)
	if err != nil {
		log.Fatalf(
			"create provider response observer: %v",
			err,
		)
	}

	orchestrator, err := ingestionorchestrator.NewDefault(
		responseController,
	)
	if err != nil {
		log.Fatalf(
			"create ingestion orchestrator: %v",
			err,
		)
	}

	client := airplaneslive.NewClientWithResponseObserver(
		integrationcommon.HTTPClientConfig{
			BaseURL:   airplaneslive.BaseURL,
			Timeout:   timeout,
			UserAgent: "global-flight-analytics-ingest",
		},
		responseObserver,
	)

	rawProvider := airplaneslive.NewProvider(
		client,
	)

	provider, err := regionalprovider.New(
		regionalprovider.Config{
			Provider:   rawProvider,
			ProviderID: providerpolicy.ProviderAirplanesLive,
			Executor:   orchestrator,
		},
	)
	if err != nil {
		log.Fatalf(
			"create orchestrated regional provider: %v",
			err,
		)
	}

	flightStateRepository := postgres.NewFlightStateRepository(
		dbPool,
	)

	trajectoryRepository := postgres.NewTrajectoryRepository(
		dbPool,
	)

	ingestionRunRepository := postgres.NewIngestionRunRepository(
		dbPool,
	)

	processingService := trafficapplication.New(
		trafficapplication.Config{
			FlightStateRepository: flightStateRepository,
			TrajectoryRepository:  trajectoryRepository,
		},
	)

	ingestionService := trafficingestion.New(
		trafficingestion.Config{
			Provider:               provider,
			ProcessingService:      processingService,
			IngestionRunRepository: ingestionRunRepository,
		},
	)

	result, err := ingestionService.LoadAndProcessByPoint(
		context.Background(),
		latitude,
		longitude,
		radius,
	)
	if err != nil {
		log.Fatalf(
			"regional traffic ingestion failed: %v",
			err,
		)
	}

	fmt.Printf(
		"ingestion_run_id=%s loaded=%d received=%d usable=%d invalid=%d stored=%d trajectories=%d stored_at=%s\n",
		result.IngestionRunID,
		result.LoadedStateCount,
		result.ProcessingResult.ProcessingResult.Stats.ReceivedCount,
		result.ProcessingResult.ProcessingResult.Stats.UsableCount,
		result.ProcessingResult.ProcessingResult.Stats.InvalidCount,
		result.ProcessingResult.StoredFlightStateCount,
		result.ProcessingResult.ProcessingResult.Stats.TrajectoryCount,
		result.ProcessingResult.StoredAt.Format(time.RFC3339),
	)
}

func mustFloat64Env(
	name string,
) float64 {
	value := os.Getenv(
		name,
	)
	if value == "" {
		log.Fatalf(
			"%s is required",
			name,
		)
	}

	parsed, err := strconv.ParseFloat(
		value,
		64,
	)
	if err != nil {
		log.Fatalf(
			"parse %s: %v",
			name,
			err,
		)
	}

	return parsed
}

func mustIntEnv(
	name string,
) int {
	value := os.Getenv(
		name,
	)
	if value == "" {
		log.Fatalf(
			"%s is required",
			name,
		)
	}

	parsed, err := strconv.Atoi(
		value,
	)
	if err != nil {
		log.Fatalf(
			"parse %s: %v",
			name,
			err,
		)
	}

	return parsed
}

func mustDurationEnv(
	name string,
) time.Duration {
	value := os.Getenv(
		name,
	)
	if value == "" {
		log.Fatalf(
			"%s is required",
			name,
		)
	}

	parsed, err := time.ParseDuration(
		value,
	)
	if err != nil {
		log.Fatalf(
			"parse %s: %v",
			name,
			err,
		)
	}

	if parsed <= 0 {
		log.Fatalf(
			"%s must be greater than zero",
			name,
		)
	}

	return parsed
}
