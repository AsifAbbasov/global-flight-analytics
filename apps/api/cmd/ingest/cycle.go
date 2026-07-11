package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerfanout"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/sharedsnapshot"
	trafficingestion "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/ingestion"
)

var (
	errIngestionCycleExecutorRequired = errors.New(
		"ingestion cycle executor is required",
	)
	errIngestionCycleTrafficSourceRequired = errors.New(
		"ingestion cycle traffic source is required",
	)
	errIngestionCycleTrafficSourceNameRequired = errors.New(
		"ingestion cycle traffic source name is required",
	)
	errIngestionCycleProcessingServiceRequired = errors.New(
		"ingestion cycle processing service is required",
	)
	errIngestionCycleRunRepositoryRequired = errors.New(
		"ingestion cycle run repository is required",
	)
)

type ingestionCycleConfig struct {
	Executor providerfanout.Executor[sharedsnapshot.Payload]

	TrafficSource     sharedsnapshot.RegionalTrafficSource
	TrafficSourceName string

	ProcessingService      trafficingestion.ProcessingService
	IngestionRunRepository trafficingestion.IngestionRunRepository

	Latitude  float64
	Longitude float64
	Radius    int
}

type ingestionCycle struct {
	executor providerfanout.Executor[sharedsnapshot.Payload]

	trafficSource     sharedsnapshot.RegionalTrafficSource
	trafficSourceName string

	processingService      trafficingestion.ProcessingService
	ingestionRunRepository trafficingestion.IngestionRunRepository

	latitude  float64
	longitude float64
	radius    int
}

func newIngestionCycle(
	config ingestionCycleConfig,
) (*ingestionCycle, error) {
	if config.Executor == nil {
		return nil, errIngestionCycleExecutorRequired
	}

	if config.TrafficSource == nil {
		return nil, errIngestionCycleTrafficSourceRequired
	}

	trafficSourceName := strings.TrimSpace(
		config.TrafficSourceName,
	)
	if trafficSourceName == "" {
		return nil, errIngestionCycleTrafficSourceNameRequired
	}

	if config.ProcessingService == nil {
		return nil, errIngestionCycleProcessingServiceRequired
	}

	if config.IngestionRunRepository == nil {
		return nil, errIngestionCycleRunRepositoryRequired
	}

	return &ingestionCycle{
		executor: config.Executor,

		trafficSource:     config.TrafficSource,
		trafficSourceName: trafficSourceName,

		processingService:      config.ProcessingService,
		ingestionRunRepository: config.IngestionRunRepository,

		latitude:  config.Latitude,
		longitude: config.Longitude,
		radius:    config.Radius,
	}, nil
}

func (
	cycle *ingestionCycle,
) Run(
	ctx context.Context,
) error {
	snapshot, err := runSharedSnapshot(
		ctx,
		sharedSnapshotRunConfig{
			Executor:      cycle.executor,
			TrafficSource: cycle.trafficSource,
			Latitude:      cycle.latitude,
			Longitude:     cycle.longitude,
			Radius:        cycle.radius,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"shared snapshot collection failed: %w",
			err,
		)
	}

	snapshotTrafficProvider, err := newSnapshotTrafficProvider(
		snapshot,
		cycle.trafficSourceName,
	)
	if err != nil {
		return fmt.Errorf(
			"create snapshot-backed traffic provider: %w",
			err,
		)
	}

	ingestionService := trafficingestion.New(
		trafficingestion.Config{
			Provider:               snapshotTrafficProvider,
			ProcessingService:      cycle.processingService,
			IngestionRunRepository: cycle.ingestionRunRepository,
		},
	)

	result, err := ingestionService.LoadAndProcessByPoint(
		ctx,
		cycle.latitude,
		cycle.longitude,
		cycle.radius,
	)
	if err != nil {
		return fmt.Errorf(
			"snapshot-backed regional traffic ingestion failed: %w",
			err,
		)
	}

	fmt.Printf(
		"snapshot_status=%s snapshot_total=%d snapshot_successes=%d snapshot_failures=%d snapshot_cycle_started_at=%s snapshot_assembled_at=%s ingestion_run_id=%s loaded=%d received=%d usable=%d invalid=%d stored=%d trajectories=%d stored_at=%s\n",
		snapshot.Status,
		snapshot.TotalCount,
		snapshot.SuccessCount,
		snapshot.FailureCount,
		snapshot.CycleStartedAt.Format(
			time.RFC3339Nano,
		),
		snapshot.AssembledAt.Format(
			time.RFC3339Nano,
		),
		result.IngestionRunID,
		result.LoadedStateCount,
		result.ProcessingResult.ProcessingResult.Stats.ReceivedCount,
		result.ProcessingResult.ProcessingResult.Stats.UsableCount,
		result.ProcessingResult.ProcessingResult.Stats.InvalidCount,
		result.ProcessingResult.StoredFlightStateCount,
		result.ProcessingResult.ProcessingResult.Stats.TrajectoryCount,
		result.ProcessingResult.StoredAt.Format(
			time.RFC3339,
		),
	)

	return nil
}
