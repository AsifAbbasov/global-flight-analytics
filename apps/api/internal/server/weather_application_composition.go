package server

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
	"github.com/jackc/pgx/v5/pgxpool"
)

type weatherApplication struct {
	repository *postgres.WeatherRepository
	service    *weatherservice.Service
	handler    *handlers.WeatherHandler
}

func composeWeatherApplication(
	databasePool *pgxpool.Pool,
	client weatherservice.CurrentWeatherClient,
) weatherApplication {
	repository := postgres.NewWeatherRepository(
		databasePool,
	)
	service := weatherservice.New(
		weatherservice.Config{
			Client:     client,
			Repository: repository,
		},
	)
	handler := handlers.NewWeatherHandler(
		service,
	)

	return weatherApplication{
		repository: repository,
		service:    service,
		handler:    handler,
	}
}
