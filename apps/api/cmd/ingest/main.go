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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/ingestionorchestrator"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerbudget"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerdecision"
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

	trafficProviderConfig, err := config.LoadTrafficProviderConfig()
	if err != nil {
		return fmt.Errorf(
			"load traffic provider configuration: %w",
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

	budgetStore, err := postgres.NewProviderBudgetStore(
		dbPool,
		daemonConfig.TerminalTimeout,
	)
	if err != nil {
		return fmt.Errorf(
			"create PostgreSQL provider budget store: %w",
			err,
		)
	}

	budgetManager, err := providerbudget.NewDurable(
		budgetStore,
		nil,
	)
	if err != nil {
		return fmt.Errorf(
			"create durable provider budget manager: %w",
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
	providerDecisionCollector := providerdecision.New(nil)

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

	orchestrator, err := ingestionorchestrator.NewDefaultWithDecisionRecorder[regionalprovider.ExecutionValue](
		responseController,
		providerDecisionCollector,
	)
	if err != nil {
		return fmt.Errorf(
			"create ingestion orchestrator: %w",
			err,
		)
	}

	trafficSelection, err := buildTrafficProvider(
		cfg.AirplanesLiveTimeout,
		trafficProviderConfig,
		orchestrator,
		responseObserver,
		providerDecisionCollector,
		providerHealthCollector,
	)
	if err != nil {
		return fmt.Errorf(
			"build regional traffic provider: %w",
			err,
		)
	}

	trafficProvider := trafficSelection.Provider

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

	recoveredRunCount, err := recoverStaleIngestionRuns(
		operationContext,
		ingestionRunRepository,
		time.Now().UTC(),
		daemonConfig.StaleRunAfter,
		daemonConfig.TerminalTimeout,
	)
	if err != nil {
		return fmt.Errorf(
			"recover stale ingestion runs: %w",
			err,
		)
	}
	fmt.Printf(
		"ingestion_run_recovery recovered=%d stale_after=%s terminal_timeout=%s\n",
		recoveredRunCount,
		daemonConfig.StaleRunAfter,
		daemonConfig.TerminalTimeout,
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
			Processor:                        trafficProcessor,
			FlightStateRepository:            flightStateRepository,
			DataQualityRepository:            dataQualityRepository,
			TrajectoryRepository:             trajectoryRepository,
			TrajectoryContinuationRepository: trajectoryRepository,
			IdentityContinuationMaxGap:       cfg.TrajectoryMaxTimeGap,
			ReconciliationRepository:         reconciliationRepository,
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
			TerminalTimeout:        daemonConfig.TerminalTimeout,
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
			RunCycle:          cycle.Run,
			Interval:          daemonConfig.Interval,
			MaxFailureBackoff: daemonConfig.MaxBackoff,
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
					"ingest_cycle=%d status=%s started_at=%s finished_at=%s duration=%s consecutive_failures=%d retry_at=%s next_delay=%s error=%q\n",
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
					result.ConsecutiveFailures,
					result.RetryAt.Format(
						time.RFC3339Nano,
					),
					result.NextDelay,
					lastError,
				)

				printTrafficProviderEvidence(
					trafficSelection.ProviderIDs,
					providerHealthCollector,
					providerDecisionCollector,
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
		"traffic_ingest_daemon_started mode=%s primary_provider=%s providers=%v interval=%s max_backoff=%s terminal_timeout=%s stale_run_after=%s latitude=%f longitude=%f radius_nm=%d\n",
		trafficSelection.Mode,
		trafficSelection.ProviderID,
		trafficSelection.ProviderIDs,
		daemonConfig.Interval,
		daemonConfig.MaxBackoff,
		daemonConfig.TerminalTimeout,
		daemonConfig.StaleRunAfter,
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

func printTrafficProviderEvidence(
	providerIDs []providerpolicy.Provider,
	healthCollector *providerhealthservice.Collector,
	decisionCollector *providerdecision.Collector,
) {
	for _, providerID := range providerIDs {
		snapshot, snapshotErr := healthCollector.Snapshot(
			providerID,
		)
		if snapshotErr != nil {
			fmt.Printf(
				"provider_health provider=%s error=%q\n",
				providerID,
				snapshotErr.Error(),
			)
		} else {
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
		}

		decisionSnapshot, decisionSnapshotErr :=
			decisionCollector.Snapshot(
				providerID,
			)
		if decisionSnapshotErr != nil {
			fmt.Printf(
				"provider_decision provider=%s error=%q\n",
				providerID,
				decisionSnapshotErr.Error(),
			)
			continue
		}

		fmt.Printf(
			"provider_decision provider=%s decisions_total=%d allowed_total=%d denied_total=%d latest_allowed=%t latest_reason=%s latest_request_key=%q retry_at=%s fallback_observed=%t fallback_total=%d primary_selected=%d fallback_selected=%d no_provider_available=%d terminal_failure=%d latest_fallback_outcome=%s latest_selected_provider=%s health_aware=%t health_reordered=%t primary_health=%s selected_health=%s health_reason=%q latest_attempts=%v reason_counts=%v limitations=%v\n",
			decisionSnapshot.Provider,
			decisionSnapshot.DecisionsTotal,
			decisionSnapshot.AllowedTotal,
			decisionSnapshot.DeniedTotal,
			decisionSnapshot.Latest.Allowed,
			decisionSnapshot.Latest.Reason,
			decisionSnapshot.Latest.RequestKey,
			decisionSnapshot.Latest.RetryAt.Format(
				time.RFC3339Nano,
			),
			decisionSnapshot.FallbackObserved,
			decisionSnapshot.FallbackDecisionsTotal,
			decisionSnapshot.PrimarySelectedTotal,
			decisionSnapshot.FallbackSelectedTotal,
			decisionSnapshot.NoProviderAvailableTotal,
			decisionSnapshot.TerminalFailureTotal,
			decisionSnapshot.LatestFallback.Outcome,
			decisionSnapshot.LatestFallback.SelectedProvider,
			decisionSnapshot.LatestFallback.HealthAware,
			decisionSnapshot.LatestFallback.HealthReordered,
			trafficProviderHealthStatusLabel(
				decisionSnapshot.LatestFallback.PrimaryHealthStatus,
			),
			trafficProviderHealthStatusLabel(
				decisionSnapshot.LatestFallback.SelectedHealthStatus,
			),
			decisionSnapshot.LatestFallback.HealthOrderingReason,
			decisionSnapshot.LatestFallback.Attempts,
			decisionSnapshot.ReasonCounts,
			decisionSnapshot.Limitations,
		)
	}
}
