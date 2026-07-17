package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/orchestration/providerpolicy"
	trafficingestion "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/ingestion"
)

var (
	errIngestionCycleTrafficProviderRequired = errors.New(
		"ingestion cycle traffic provider is required",
	)
	errIngestionCycleProcessingServiceRequired = errors.New(
		"ingestion cycle processing service is required",
	)
	errIngestionCycleRunRepositoryRequired = errors.New(
		"ingestion cycle run repository is required",
	)
	errIngestionCycleProviderIdentityInvalid = errors.New(
		"ingestion cycle traffic provider identity is invalid",
	)
)

type providerObservationRecorder interface {
	RecordObservationEvidence(
		provider providerpolicy.Provider,
		received int64,
		accepted int64,
		rejected int64,
	) error
}

type ingestionCycleConfig struct {
	TrafficProvider        trafficingestion.RegionalProvider
	ProcessingService      trafficingestion.ProcessingService
	IngestionRunRepository trafficingestion.IngestionRunRepository
	ObservationRecorder    providerObservationRecorder

	Latitude  float64
	Longitude float64
	Radius    int
}

type ingestionCycle struct {
	service             *trafficingestion.Service
	providerID          providerpolicy.Provider
	observationRecorder providerObservationRecorder

	latitude  float64
	longitude float64
	radius    int
}

func newIngestionCycle(
	config ingestionCycleConfig,
) (*ingestionCycle, error) {
	if config.TrafficProvider == nil {
		return nil, errIngestionCycleTrafficProviderRequired
	}

	if config.ProcessingService == nil {
		return nil, errIngestionCycleProcessingServiceRequired
	}

	if config.IngestionRunRepository == nil {
		return nil, errIngestionCycleRunRepositoryRequired
	}

	providerID := providerpolicy.Provider("")
	if config.ObservationRecorder != nil {
		providerID = providerpolicy.Provider(
			strings.TrimSpace(
				config.TrafficProvider.SourceName(),
			),
		)

		if _, err := providerpolicy.Get(providerID); err != nil {
			return nil, errors.Join(
				errIngestionCycleProviderIdentityInvalid,
				err,
			)
		}
	}

	return &ingestionCycle{
		service: trafficingestion.New(
			trafficingestion.Config{
				Provider:               config.TrafficProvider,
				ProcessingService:      config.ProcessingService,
				IngestionRunRepository: config.IngestionRunRepository,
			},
		),
		providerID:          providerID,
		observationRecorder: config.ObservationRecorder,
		latitude:            config.Latitude,
		longitude:           config.Longitude,
		radius:              config.Radius,
	}, nil
}

func (
	cycle *ingestionCycle,
) Run(
	ctx context.Context,
) error {
	result, err := cycle.service.LoadAndProcessByPoint(
		ctx,
		cycle.latitude,
		cycle.longitude,
		cycle.radius,
	)

	cycle.observeProviderEvidence(result)

	if err != nil {
		return fmt.Errorf(
			"regional traffic ingestion failed: %w",
			err,
		)
	}

	fmt.Printf(
		"ingestion_run_id=%s source=%s loaded=%d received=%d usable=%d invalid=%d stored=%d trajectories=%d stored_at=%s\n",
		result.IngestionRunID,
		result.SourceName,
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

func (cycle *ingestionCycle) observeProviderEvidence(
	result trafficingestion.LoadAndProcessResult,
) {
	if cycle.observationRecorder == nil {
		return
	}

	processingResult := result.ProcessingResult.ProcessingResult
	if processingResult.ProcessedAt.IsZero() {
		return
	}

	providerID := cycle.providerID
	if normalizedSourceName := strings.TrimSpace(
		result.SourceName,
	); normalizedSourceName != "" {
		candidate := providerpolicy.Provider(
			normalizedSourceName,
		)
		if _, err := providerpolicy.Get(candidate); err != nil {
			fmt.Printf(
				"provider_health_observation_evidence source=%q error=%q\n",
				normalizedSourceName,
				err.Error(),
			)
			return
		}
		providerID = candidate
	}

	stats := processingResult.Stats
	err := cycle.observationRecorder.RecordObservationEvidence(
		providerID,
		int64(stats.ReceivedCount),
		int64(stats.UsableCount),
		int64(stats.InvalidCount+stats.DuplicateCount),
	)
	if err != nil {
		fmt.Printf(
			"provider_health_observation_evidence provider=%s error=%q\n",
			providerID,
			err.Error(),
		)
	}
}
