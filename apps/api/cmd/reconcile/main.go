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
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/reconciliationworker"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/gapdetector"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/processor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/trackbuilder"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf(
			"reconciliation worker failed: %v",
			err,
		)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg, err := config.LoadReconciliationWorkerConfig()
	if err != nil {
		return fmt.Errorf(
			"load reconciliation worker configuration: %w",
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

	reconciliationRepository := postgres.NewReconciliationRepository(
		dbPool,
	)
	flightStateRepository := postgres.NewFlightStateRepository(
		dbPool,
	)
	dataQualityRepository := postgres.NewDataQualityRepository(
		dbPool,
	)
	trajectoryRepository := postgres.NewTrajectoryRepository(
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

	worker, err := reconciliationworker.New(
		reconciliationworker.Config{
			Repository:            reconciliationRepository,
			FlightStateRepository: flightStateRepository,
			DataQualityRepository: dataQualityRepository,
			TrajectoryRepository:  trajectoryRepository,
			Processor:             trafficProcessor,
			MaxAttempts:           cfg.MaxAttempts,
			RetryBaseDelay:        cfg.RetryBaseDelay,
			RetryMaximumDelay:     cfg.RetryMaximumDelay,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"create reconciliation worker: %w",
			err,
		)
	}

	requeuedCount, err := reconciliationRepository.RequeueStaleProcessing(
		operationContext,
		time.Now().
			UTC().
			Add(
				-cfg.StaleAfter,
			),
	)
	if err != nil {
		return fmt.Errorf(
			"requeue stale reconciliation tasks: %w",
			err,
		)
	}

	processedCount := 0
	completedCount := 0
	retryCount := 0
	failedCount := 0
	requeuedBySignalCount := 0

	for processedCount < cfg.MaximumTasks {
		if err := operationContext.Err(); err != nil {
			return err
		}

		result, err := worker.RunOnce(
			operationContext,
		)
		if err != nil {
			return err
		}

		if !result.TaskFound {
			break
		}

		processedCount++

		switch result.FinalStatus {
		case "completed":
			completedCount++

		case "pending":
			if result.PersistedItemCount > 0 {
				requeuedBySignalCount++
			} else {
				retryCount++
			}

		case "failed":
			failedCount++
		}

		fmt.Printf(
			"task_id=%s icao24=%s derivation_type=%s attempt=%d final_status=%s persisted=%d next_attempt_at=%s last_error=%q\n",
			result.TaskID,
			result.ICAO24,
			result.DerivationType,
			result.AttemptCount,
			result.FinalStatus,
			result.PersistedItemCount,
			formatOptionalTime(
				result.NextAttemptAt,
			),
			result.LastError,
		)
	}

	fmt.Printf(
		"requeued_stale=%d processed=%d completed=%d retries=%d failed=%d requeued_by_new_signal=%d maximum_tasks=%d\n",
		requeuedCount,
		processedCount,
		completedCount,
		retryCount,
		failedCount,
		requeuedBySignalCount,
		cfg.MaximumTasks,
	)

	return nil
}

func formatOptionalTime(
	value time.Time,
) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(
		time.RFC3339Nano,
	)
}
