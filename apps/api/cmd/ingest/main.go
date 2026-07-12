package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/ingestdaemon"
	providerhealthservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/providerhealth"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/gapdetector"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/trackbuilder"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf(
			"traffic ingest daemon failed: %v",
			err,
		)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg, err := config.LoadIngestConfig()
	if err != nil {
		return fmt.Errorf(
			"load ingest configuration: %w",
			err,
		)
	}

	daemonConfig, err := config.LoadIngestDaemonConfig()
	if err != nil {
		return fmt.Errorf(
			"load ingest daemon configuration: %w",
			err,
		)
	}

	operationContext, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	dbPool, err := database.NewPostgresPool(
		cfg.Database.URL,
		cfg.Database.ConnectTimeout,
	)
	if err != nil {
		return fmt.Errorf(
			"connect postgres: %w",
			err,
		)
	}
	defer dbPool.Close()

	budgetManager, err := providerbudget.New(
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"create provider budget manager: %w",
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
			"create provider response controller: %w",
			err,
		)
	}

	providerHealthCollector := providerhealthservice.New(nil)

	responseObserver, err := providerresponse.NewIntegrationObserverWithRecorder(
		responseController,
		providerHealthCollector,
	)
	if err != nil {
		return fmt.Errorf(
			"create provider response observer: %w",
			err,
		)
	}

	orchestrator, err := ingestionorchestrator.NewDefault[regionalprovider.ExecutionValue](
		responseController,
	)
	if err != nil {
		return fmt.Errorf(
			"create ingestion orchestrator: %w",
			err,
		)
	}

	airplanesLiveClient, err := airplaneslive.NewClientWithResponseObserver(
		integrationcommon.HTTPClientConfig{
			BaseURL:   airplaneslive.BaseURL,
			Timeout:   cfg.AirplanesLiveTimeout,
			UserAgent: "global-flight-analytics-ingest",
		},
		responseObserver,
	)
	if err != nil {
		return fmt.Errorf(
			"create airplanes.live client: %w",
			err,
		)
	}

	rawTrafficProvider := airplaneslive.NewProvider(
		airplanesLiveClient,
	)

	trafficProvider, err := regionalprovider.New(
		regionalprovider.Config{
			Provider:   rawTrafficProvider,
			ProviderID: providerpolicy.ProviderAirplanesLive,
			Executor:   orchestrator,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"create orchestrated regional traffic provider: %w",
			err,
		)
	}

	flightStateRepository := postgres.NewFlightStateRepository(
		dbPool,
	)
	dataQualityRepository := postgres.NewDataQualityRepository(
		dbPool,
	)
	trajectoryRepository := postgres.NewTrajectoryRepository(
		dbPool,
	)
	reconciliationRepository := postgres.NewReconciliationRepository(
		dbPool,
	)
	ingestionRunRepository := postgres.NewIngestionRunRepository(
		dbPool,
	)

	trafficProcessor, err := processor.New(
		processor.Config{
			TrackBuilderConfig: trackbuilder.Config{
				GapDetectorConfig: gapdetector.Config{
					MaxTimeGap:        cfg.TrajectoryMaxTimeGap,
					MaxGroundSpeedMPS: cfg.TrajectoryMaxGroundSpeedMPS,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf(
			"create traffic processor: %w",
			err,
		)
	}

	processingService, err := trafficapplication.New(
		trafficapplication.Config{
			Processor:                trafficProcessor,
			FlightStateRepository:    flightStateRepository,
			DataQualityRepository:    dataQualityRepository,
			TrajectoryRepository:     trajectoryRepository,
			ReconciliationRepository: reconciliationRepository,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"create traffic application service: %w",
			err,
		)
	}

	cycle, err := newIngestionCycle(
		ingestionCycleConfig{
			TrafficProvider:        trafficProvider,
			ProcessingService:      processingService,
			IngestionRunRepository: ingestionRunRepository,
			ObservationRecorder:    providerHealthCollector,
			Latitude:               cfg.TrafficIngestionLatitude,
			Longitude:              cfg.TrafficIngestionLongitude,
			Radius:                 cfg.TrafficIngestionRadius,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"create traffic ingestion cycle: %w",
			err,
		)
	}

	daemon, err := ingestdaemon.New(
		ingestdaemon.Config{
			RunCycle: cycle.Run,
			Interval: daemonConfig.Interval,
			Observe: func(
				result ingestdaemon.CycleResult,
			) {
				status := "success"
				lastError := ""

				if result.Err != nil {
					status = "failed"
					lastError = result.Err.Error()
				}

				fmt.Printf(
					"ingest_cycle=%d status=%s started_at=%s finished_at=%s duration=%s next_interval=%s error=%q\n",
					result.Number,
					status,
					result.StartedAt.Format(
						time.RFC3339Nano,
					),
					result.FinishedAt.Format(
						time.RFC3339Nano,
					),
					result.FinishedAt.Sub(
						result.StartedAt,
					),
					daemonConfig.Interval,
					lastError,
				)

				snapshot, snapshotErr := providerHealthCollector.Snapshot(
					providerpolicy.ProviderAirplanesLive,
				)
				if snapshotErr != nil {
					fmt.Printf(
						"provider_health provider=%s error=%q\n",
						providerpolicy.ProviderAirplanesLive,
						snapshotErr.Error(),
					)
					return
				}

				fmt.Printf(
					"provider_health provider=%s status=%s requests_total=%d requests_successful=%d success_ratio=%.4f consecutive_failures=%d latest_outcome=%s budget_state=%s reasons=%v limitations=%v\n",
					snapshot.ProviderName,
					snapshot.Status,
					snapshot.RequestsTotal,
					snapshot.RequestsSuccessful,
					snapshot.SuccessRatio,
					snapshot.ConsecutiveFailures,
					snapshot.LatestOutcome,
					snapshot.Budget.State,
					snapshot.Reasons,
					snapshot.Limitations,
				)
			},
		},
	)
	if err != nil {
		return fmt.Errorf(
			"create traffic ingest daemon: %w",
			err,
		)
	}

	fmt.Printf(
		"traffic_ingest_daemon_started interval=%s latitude=%f longitude=%f radius=%d\n",
		daemonConfig.Interval,
		cfg.TrafficIngestionLatitude,
		cfg.TrafficIngestionLongitude,
		cfg.TrafficIngestionRadius,
	)

	if err := daemon.Run(
		operationContext,
	); err != nil {
		return fmt.Errorf(
			"run traffic ingest daemon: %w",
			err,
		)
	}

	fmt.Println(
		"traffic_ingest_daemon_stopped",
	)

	return nil
}
