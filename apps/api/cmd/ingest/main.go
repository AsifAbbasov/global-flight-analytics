package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/config"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/database"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/airplaneslive"
	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerresponse"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/sharedsnapshot"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/gapdetector"
	trafficingestion "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/ingestion"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/trackbuilder"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.LoadIngestConfig()
	if err != nil {
		log.Fatalf(
			"load ingest configuration: %v",
			err,
		)
	}

	latitude := cfg.TrafficIngestionLatitude
	longitude := cfg.TrafficIngestionLongitude
	radius := cfg.TrafficIngestionRadius

	airplanesLiveTimeout := cfg.AirplanesLiveTimeout

	trajectoryMaxTimeGap := cfg.TrajectoryMaxTimeGap
	trajectoryMaxGroundSpeedMPS := cfg.TrajectoryMaxGroundSpeedMPS

	operationContext := context.Background()

	dbPool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
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

	orchestrator, err := ingestionorchestrator.NewDefault[sharedsnapshot.Payload](
		responseController,
	)
	if err != nil {
		log.Fatalf(
			"create ingestion orchestrator: %v",
			err,
		)
	}

	airplanesLiveClient, err := airplaneslive.NewClientWithResponseObserver(
		integrationcommon.HTTPClientConfig{
			BaseURL:   airplaneslive.BaseURL,
			Timeout:   airplanesLiveTimeout,
			UserAgent: "global-flight-analytics-ingest",
		},
		responseObserver,
	)
	if err != nil {
		log.Fatalf(
			"create airplanes.live client: %v",
			err,
		)
	}

	rawTrafficProvider := airplaneslive.NewProvider(
		airplanesLiveClient,
	)

	snapshot, err := runSharedSnapshot(
		operationContext,
		sharedSnapshotRunConfig{
			Executor:      orchestrator,
			TrafficSource: rawTrafficProvider,
			Latitude:      latitude,
			Longitude:     longitude,
			Radius:        radius,
		},
	)
	if err != nil {
		log.Fatalf(
			"shared snapshot collection failed: %v",
			err,
		)
	}

	snapshotTrafficProvider, err := newSnapshotTrafficProvider(
		snapshot,
		rawTrafficProvider.SourceName(),
	)
	if err != nil {
		log.Fatalf(
			"create snapshot-backed traffic provider: %v",
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

	trafficProcessor, err := processor.New(
		processor.Config{
			TrackBuilderConfig: trackbuilder.Config{
				GapDetectorConfig: gapdetector.Config{
					MaxTimeGap:        trajectoryMaxTimeGap,
					MaxGroundSpeedMPS: trajectoryMaxGroundSpeedMPS,
				},
			},
		},
	)
	if err != nil {
		log.Fatalf(
			"create traffic processor: %v",
			err,
		)
	}

	processingService, err := trafficapplication.New(
		trafficapplication.Config{
			Processor:             trafficProcessor,
			FlightStateRepository: flightStateRepository,
			TrajectoryRepository:  trajectoryRepository,
		},
	)
	if err != nil {
		log.Fatalf(
			"create traffic application service: %v",
			err,
		)
	}

	ingestionService := trafficingestion.New(
		trafficingestion.Config{
			Provider:               snapshotTrafficProvider,
			ProcessingService:      processingService,
			IngestionRunRepository: ingestionRunRepository,
		},
	)

	result, err := ingestionService.LoadAndProcessByPoint(
		operationContext,
		latitude,
		longitude,
		radius,
	)
	if err != nil {
		log.Fatalf(
			"snapshot-backed regional traffic ingestion failed: %v",
			err,
		)
	}

	fmt.Printf(
		"snapshot_status=%s snapshot_total=%d snapshot_successes=%d snapshot_failures=%d snapshot_cycle_started_at=%s snapshot_assembled_at=%s ingestion_run_id=%s loaded=%d received=%d usable=%d invalid=%d stored=%d trajectories=%d stored_at=%s\n",
		snapshot.Status,
		snapshot.TotalCount,
		snapshot.SuccessCount,
		snapshot.FailureCount,
		snapshot.CycleStartedAt.Format(time.RFC3339Nano),
		snapshot.AssembledAt.Format(time.RFC3339Nano),
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
