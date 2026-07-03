package weather

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"
)

func TestGetAndStoreCurrentWeatherSuccess(t *testing.T) {
	client := &fakeCurrentWeatherClient{
		snapshot: makeCurrentWeatherSnapshot(),
	}

	repository := &fakeCurrentSnapshotRepository{
		snapshotID: "weather-snapshot-1",
	}

	service := New(Config{
		Client:     client,
		Repository: repository,
	})

	result, err := service.GetAndStoreCurrentWeather(context.Background(), CurrentWeatherRequest{
		Latitude:  40.4675,
		Longitude: 50.0467,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !client.called {
		t.Fatal("expected weather client to be called")
	}

	if client.lastRequest.Latitude != 40.4675 {
		t.Fatalf("expected latitude 40.4675, got %f", client.lastRequest.Latitude)
	}

	if client.lastRequest.Longitude != 50.0467 {
		t.Fatalf("expected longitude 50.0467, got %f", client.lastRequest.Longitude)
	}

	if !repository.called {
		t.Fatal("expected weather repository to be called")
	}

	if repository.lastSnapshot.Provider != domainweather.ProviderOpenMeteo {
		t.Fatalf("expected provider %s, got %s", domainweather.ProviderOpenMeteo, repository.lastSnapshot.Provider)
	}

	if result.SnapshotID != "weather-snapshot-1" {
		t.Fatalf("expected snapshot id weather-snapshot-1, got %s", result.SnapshotID)
	}

	if result.Snapshot.Provider != domainweather.ProviderOpenMeteo {
		t.Fatalf("expected result provider %s, got %s", domainweather.ProviderOpenMeteo, result.Snapshot.Provider)
	}

	if result.StoredAt.IsZero() {
		t.Fatal("expected stored at timestamp")
	}
}

func TestGetAndStoreCurrentWeatherRejectsInvalidCoordinates(t *testing.T) {
	client := &fakeCurrentWeatherClient{}
	repository := &fakeCurrentSnapshotRepository{}

	service := New(Config{
		Client:     client,
		Repository: repository,
	})

	_, err := service.GetAndStoreCurrentWeather(context.Background(), CurrentWeatherRequest{
		Latitude:  91,
		Longitude: 50,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrInvalidWeatherCoordinates) {
		t.Fatalf("expected ErrInvalidWeatherCoordinates, got %v", err)
	}

	if client.called {
		t.Fatal("expected weather client not to be called")
	}

	if repository.called {
		t.Fatal("expected weather repository not to be called")
	}
}

func TestGetAndStoreCurrentWeatherRequiresClient(t *testing.T) {
	service := New(Config{
		Repository: &fakeCurrentSnapshotRepository{},
	})

	_, err := service.GetAndStoreCurrentWeather(context.Background(), CurrentWeatherRequest{
		Latitude:  40.4675,
		Longitude: 50.0467,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrWeatherClientRequired) {
		t.Fatalf("expected ErrWeatherClientRequired, got %v", err)
	}
}

func TestGetAndStoreCurrentWeatherRequiresRepository(t *testing.T) {
	service := New(Config{
		Client: &fakeCurrentWeatherClient{},
	})

	_, err := service.GetAndStoreCurrentWeather(context.Background(), CurrentWeatherRequest{
		Latitude:  40.4675,
		Longitude: 50.0467,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrWeatherRepositoryRequired) {
		t.Fatalf("expected ErrWeatherRepositoryRequired, got %v", err)
	}
}

func TestGetAndStoreCurrentWeatherReturnsClientError(t *testing.T) {
	expectedError := errors.New("open-meteo failed")

	client := &fakeCurrentWeatherClient{
		err: expectedError,
	}

	repository := &fakeCurrentSnapshotRepository{}

	service := New(Config{
		Client:     client,
		Repository: repository,
	})

	_, err := service.GetAndStoreCurrentWeather(context.Background(), CurrentWeatherRequest{
		Latitude:  40.4675,
		Longitude: 50.0467,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected wrapped client error, got %v", err)
	}

	if !strings.Contains(err.Error(), "get current weather") {
		t.Fatalf("expected contextual client error, got %v", err)
	}

	if repository.called {
		t.Fatal("expected repository not to be called after client error")
	}
}

func TestGetAndStoreCurrentWeatherReturnsRepositoryError(t *testing.T) {
	expectedError := errors.New("repository failed")

	client := &fakeCurrentWeatherClient{
		snapshot: makeCurrentWeatherSnapshot(),
	}

	repository := &fakeCurrentSnapshotRepository{
		err: expectedError,
	}

	service := New(Config{
		Client:     client,
		Repository: repository,
	})

	_, err := service.GetAndStoreCurrentWeather(context.Background(), CurrentWeatherRequest{
		Latitude:  40.4675,
		Longitude: 50.0467,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, expectedError) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}

	if !strings.Contains(err.Error(), "save current weather snapshot") {
		t.Fatalf("expected contextual repository error, got %v", err)
	}
}

func makeCurrentWeatherSnapshot() domainweather.CurrentSnapshot {
	now := time.Date(2026, 7, 3, 8, 15, 0, 0, time.UTC)

	return domainweather.CurrentSnapshot{
		Provider:                 domainweather.ProviderOpenMeteo,
		Latitude:                 40.4375,
		Longitude:                50.0625,
		ObservedAt:               now,
		TemperatureCelsius:       29.5,
		RelativeHumidityPercent:  55,
		PrecipitationMillimeters: 0,
		RainMillimeters:          0,
		WeatherCode:              0,
		CloudCoverPercent:        0,
		SurfacePressureHPA:       1010.3,
		WindSpeedMetersPerSecond: 5.36,
		WindDirectionDegrees:     194,
		WindGustsMetersPerSecond: 9.7,
		RetrievedAt:              now.Add(time.Second),
	}
}

type fakeCurrentWeatherClient struct {
	called      bool
	lastRequest openmeteo.CurrentWeatherRequest
	snapshot    domainweather.CurrentSnapshot
	err         error
}

func (client *fakeCurrentWeatherClient) GetCurrentWeather(ctx context.Context, request openmeteo.CurrentWeatherRequest) (domainweather.CurrentSnapshot, error) {
	client.called = true
	client.lastRequest = request

	if client.err != nil {
		return domainweather.CurrentSnapshot{}, client.err
	}

	return client.snapshot, nil
}

type fakeCurrentSnapshotRepository struct {
	called       bool
	snapshotID   string
	lastSnapshot domainweather.CurrentSnapshot
	err          error
}

func (repository *fakeCurrentSnapshotRepository) SaveCurrentSnapshot(ctx context.Context, snapshot domainweather.CurrentSnapshot) (string, error) {
	repository.called = true
	repository.lastSnapshot = snapshot

	if repository.err != nil {
		return "", repository.err
	}

	if repository.snapshotID == "" {
		return "weather-snapshot-id", nil
	}

	return repository.snapshotID, nil
}
