package ingestion

import (
	"context"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	trafficapplication "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/application"
)

type RegionalProvider interface {
	LoadByPoint(
		ctx context.Context,
		latitude float64,
		longitude float64,
		radius int,
	) ([]flightstate.FlightState, error)
}

type ProcessingService interface {
	ProcessAndStore(
		ctx context.Context,
		states []flightstate.FlightState,
	) (trafficapplication.ProcessAndStoreResult, error)
}

type Config struct {
	Provider          RegionalProvider
	ProcessingService ProcessingService
}

type Service struct {
	provider          RegionalProvider
	processingService ProcessingService
}

type LoadAndProcessResult struct {
	LoadedStateCount int
	ProcessingResult trafficapplication.ProcessAndStoreResult
}

func New(config Config) *Service {
	return &Service{
		provider:          config.Provider,
		processingService: config.ProcessingService,
	}
}

func (service *Service) LoadAndProcessByPoint(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radius int,
) (LoadAndProcessResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if service == nil || service.provider == nil {
		return LoadAndProcessResult{}, fmt.Errorf(
			"regional traffic provider is required",
		)
	}

	if service.processingService == nil {
		return LoadAndProcessResult{}, fmt.Errorf(
			"traffic processing service is required",
		)
	}

	states, err := service.provider.LoadByPoint(
		ctx,
		latitude,
		longitude,
		radius,
	)
	if err != nil {
		return LoadAndProcessResult{}, fmt.Errorf(
			"load regional flight states: %w",
			err,
		)
	}

	processingResult, err := service.processingService.ProcessAndStore(
		ctx,
		states,
	)
	if err != nil {
		return LoadAndProcessResult{}, fmt.Errorf(
			"process and store regional flight states: %w",
			err,
		)
	}

	return LoadAndProcessResult{
		LoadedStateCount: len(states),
		ProcessingResult: processingResult,
	}, nil
}
