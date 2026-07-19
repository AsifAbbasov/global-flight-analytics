package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routepipeline"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type routeIntelligenceDatabaseRuntime struct {
	handler *handlers.RouteIntelligenceHandler
}

func registerRouteIntelligenceDatabaseContext(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
	mutationAuthorization fiber.Handler,
) error {
	runtime, err :=
		composeRouteIntelligenceDatabaseRuntime(
			dbPool,
		)
	if err != nil {
		return err
	}

	registerRouteIntelligenceDatabaseRoutes(
		v1,
		runtime,
		mutationAuthorization,
	)
	return nil
}

func composeRouteIntelligenceDatabaseRuntime(
	dbPool *pgxpool.Pool,
) (
	routeIntelligenceDatabaseRuntime,
	error,
) {
	composition, err :=
		routepipeline.NewPostgres(
			routepipeline.PostgresConfig{
				Pool: dbPool,
			},
		)
	if err != nil {
		return routeIntelligenceDatabaseRuntime{},
			fmt.Errorf(
				"compose production Route Intelligence pipeline: %w",
				err,
			)
	}

	return routeIntelligenceDatabaseRuntime{
		handler: handlers.NewRouteIntelligenceHandler(
			composition.Pipeline,
			composition.Store,
		),
	}, nil
}
