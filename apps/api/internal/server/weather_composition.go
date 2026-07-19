package server

import (
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	weatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
	"github.com/jackc/pgx/v5/pgxpool"
)

type weatherCompositionConfig struct {
	databasePool     *pgxpool.Pool
	openMeteoTimeout time.Duration
}

type weatherRouteDependencies struct {
	client     weatherservice.CurrentWeatherClient
	repository *postgres.WeatherRepository
	service    *weatherservice.Service
	handler    *handlers.WeatherHandler
}

func composeWeatherRouteDependencies(
	config weatherCompositionConfig,
) (weatherRouteDependencies, error) {
	client, err := composeWeatherProvider(
		config.openMeteoTimeout,
	)
	if err != nil {
		return weatherRouteDependencies{},
			err
	}

	application := composeWeatherApplication(
		config.databasePool,
		client,
	)

	return weatherRouteDependencies{
		client:     client,
		repository: application.repository,
		service:    application.service,
		handler:    application.handler,
	}, nil
}
