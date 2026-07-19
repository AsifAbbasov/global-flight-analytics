package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/stabilityproduction"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type projectionDatabaseRuntime struct {
	projection handlers.ProjectionIntelligenceReader
	stability  handlers.StabilityIntelligenceReader
	weather    handlers.WeatherContextReader
}

func registerProjectionDatabaseContext(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
) error {
	runtime, err :=
		composeProjectionDatabaseRuntime(
			dbPool,
		)
	if err != nil {
		return err
	}

	return registerProjectionDatabaseRoutes(
		v1,
		runtime,
	)
}

func composeProjectionDatabaseRuntime(
	dbPool *pgxpool.Pool,
) (
	projectionDatabaseRuntime,
	error,
) {
	projectionReader, err :=
		newProjectionIntelligencePostgresReader(
			dbPool,
		)
	if err != nil {
		return projectionDatabaseRuntime{},
			fmt.Errorf(
				"compose production Projection Intelligence reader: %w",
				err,
			)
	}

	stabilityService, err :=
		stabilityproduction.New(
			stabilityproduction.Config{
				ProjectionReader: stabilityProjectionReaderAdapter{
					reader: projectionReader,
				},
			},
		)
	if err != nil {
		return projectionDatabaseRuntime{},
			fmt.Errorf(
				"compose production Stability Intelligence service: %w",
				err,
			)
	}

	weatherReader, err :=
		newWeatherContextPostgresReader(
			dbPool,
			projectionReader,
		)
	if err != nil {
		return projectionDatabaseRuntime{},
			fmt.Errorf(
				"compose production Weather Context reader: %w",
				err,
			)
	}

	return projectionDatabaseRuntime{
		projection: projectionReader,
		stability:  stabilityService,
		weather:    weatherReader,
	}, nil
}
