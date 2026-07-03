package weather

import (
	"context"
	"errors"
	"fmt"
	"time"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"
)

var (
	ErrWeatherClientRequired     = errors.New("weather client is required")
	ErrWeatherRepositoryRequired = errors.New("weather repository is required")
	ErrInvalidWeatherCoordinates = errors.New("invalid weather coordinates")
)

type CurrentWeatherClient interface {
	GetCurrentWeather(ctx context.Context, request openmeteo.CurrentWeatherRequest) (domainweather.CurrentSnapshot, error)
}

type CurrentSnapshotRepository interface {
	SaveCurrentSnapshot(ctx context.Context, snapshot domainweather.CurrentSnapshot) (string, error)
}

type Config struct {
	Client     CurrentWeatherClient
	Repository CurrentSnapshotRepository
}

type Service struct {
	client     CurrentWeatherClient
	repository CurrentSnapshotRepository
}

type CurrentWeatherRequest struct {
	Latitude  float64
	Longitude float64
}

type CurrentWeatherResult struct {
	SnapshotID string
	Snapshot   domainweather.CurrentSnapshot
	StoredAt   time.Time
}

func New(config Config) *Service {
	return &Service{
		client:     config.Client,
		repository: config.Repository,
	}
}

func (service *Service) GetAndStoreCurrentWeather(ctx context.Context, request CurrentWeatherRequest) (CurrentWeatherResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if service == nil || service.client == nil {
		return CurrentWeatherResult{}, ErrWeatherClientRequired
	}

	if service.repository == nil {
		return CurrentWeatherResult{}, ErrWeatherRepositoryRequired
	}

	if !aviationconstraints.IsLatitude(request.Latitude) ||
		!aviationconstraints.IsLongitude(request.Longitude) {
		return CurrentWeatherResult{}, ErrInvalidWeatherCoordinates
	}

	snapshot, err := service.client.GetCurrentWeather(ctx, openmeteo.CurrentWeatherRequest{
		Latitude:  request.Latitude,
		Longitude: request.Longitude,
	})
	if err != nil {
		return CurrentWeatherResult{}, fmt.Errorf("get current weather: %w", err)
	}

	snapshotID, err := service.repository.SaveCurrentSnapshot(ctx, snapshot)
	if err != nil {
		return CurrentWeatherResult{}, fmt.Errorf("save current weather snapshot: %w", err)
	}

	return CurrentWeatherResult{
		SnapshotID: snapshotID,
		Snapshot:   snapshot,
		StoredAt:   time.Now().UTC(),
	}, nil
}
