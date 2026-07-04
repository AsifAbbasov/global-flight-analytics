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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
	trafficingestion "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/ingestion"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	latitude := mustFloat64Env("TRAFFIC_INGESTION_LATITUDE")
	longitude := mustFloat64Env("TRAFFIC_INGESTION_LONGITUDE")
	radius := mustIntEnv("TRAFFIC_INGESTION_RADIUS")
	timeout := mustDurationEnv("AIRPLANES_LIVE_TIMEOUT")

	dbPool, err := database.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer dbPool.Close()

	client := airplaneslive.NewClient(
		integrationcommon.HTTPClientConfig{
			BaseURL:   airplaneslive.BaseURL,
			Timeout:   timeout,
			UserAgent: "global-flight-analytics-ingest",
		},
	)

	provider := airplaneslive.NewProvider(client)

	flightStateRepository := postgres.NewFlightStateRepository(dbPool)
	trajectoryRepository := postgres.NewTrajectoryRepository(dbPool)
	ingestionRunRepository := postgres.NewIngestionRunRepository(dbPool)

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
		log.Fatalf("regional traffic ingestion failed: %v", err)
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

func mustFloat64Env(name string) float64 {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatalf("parse %s: %v", name, err)
	}

	return parsed
}

func mustIntEnv(name string) int {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("parse %s: %v", name, err)
	}

	return parsed
}

func mustDurationEnv(name string) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s is required", name)
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("parse %s: %v", name, err)
	}

	if parsed <= 0 {
		log.Fatalf("%s must be greater than zero", name)
	}

	return parsed
}
