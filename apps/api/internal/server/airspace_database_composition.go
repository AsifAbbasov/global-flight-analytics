package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type airspaceDatabaseRuntime struct {
	regionAnalytics handlers.AirspaceRegionAnalyticsReader
}

func registerAirspaceDatabaseContext(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
) error {
	runtime, err :=
		composeAirspaceDatabaseRuntime(
			dbPool,
		)
	if err != nil {
		return err
	}

	return registerAirspaceDatabaseRoutes(
		v1,
		runtime,
	)
}

func composeAirspaceDatabaseRuntime(
	dbPool *pgxpool.Pool,
) (
	airspaceDatabaseRuntime,
	error,
) {
	observationReader, err :=
		airspaceproduction.
			NewPostgresObservationReader(
				dbPool,
			)
	if err != nil {
		return airspaceDatabaseRuntime{},
			fmt.Errorf(
				"compose production Airspace Intelligence PostgreSQL reader: %w",
				err,
			)
	}

	service, err :=
		airspaceproduction.New(
			airspaceproduction.Config{
				ObservationReader: observationReader,
				RegionResolver:    region.NewService(),
			},
		)
	if err != nil {
		return airspaceDatabaseRuntime{},
			fmt.Errorf(
				"compose production Airspace Region Analytics service: %w",
				err,
			)
	}

	return airspaceDatabaseRuntime{
		regionAnalytics: service,
	}, nil
}
