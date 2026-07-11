package main

import (
	"context"
	"errors"
	"fmt"
	"time"

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
)

type ingestionCycleConfig struct {
	TrafficProvider        trafficingestion.RegionalProvider
	ProcessingService      trafficingestion.ProcessingService
	IngestionRunRepository trafficingestion.IngestionRunRepository

	Latitude  float64
	Longitude float64
	Radius    int
}

type ingestionCycle struct {
	service *trafficingestion.Service

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

	return &ingestionCycle{
		service: trafficingestion.New(
			trafficingestion.Config{
				Provider:               config.TrafficProvider,
				ProcessingService:      config.ProcessingService,
				IngestionRunRepository: config.IngestionRunRepository,
			},
		),
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
	result, err := cycle.service.LoadAndProcessByPoint(
		ctx,
		cycle.latitude,
		cycle.longitude,
		cycle.radius,
	)
	if err != nil {
		return fmt.Errorf(
			"regional traffic ingestion failed: %w",
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
		result.ProcessingResult.StoredAt.Format(
			time.RFC3339,
		),
	)

	return nil
}
