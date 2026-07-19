package server

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type databaseRouteGroup struct {
	name     string
	register func() error
}

func registerDatabaseRoutes(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
	openMeteoTimeout time.Duration,
	mutationAuthorization fiber.Handler,
) error {
	return registerDatabaseRouteGroups(
		databaseRouteGroups(
			v1,
			dbPool,
			openMeteoTimeout,
			mutationAuthorization,
		),
	)
}

func databaseRouteGroups(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
	openMeteoTimeout time.Duration,
	mutationAuthorization fiber.Handler,
) []databaseRouteGroup {
	return []databaseRouteGroup{
		{
			name: "core database routes",
			register: func() error {
				return registerCoreDatabaseContext(
					v1,
					dbPool,
				)
			},
		},
		{
			name: "transponder evidence route",
			register: func() error {
				return registerTransponderEvidenceDatabaseContext(
					v1,
					dbPool,
				)
			},
		},
		{
			name: "Route Intelligence routes",
			register: func() error {
				return registerRouteIntelligenceDatabaseContext(
					v1,
					dbPool,
					mutationAuthorization,
				)
			},
		},
		{
			name: "weather route",
			register: func() error {
				return registerWeatherRoute(
					v1,
					dbPool,
					openMeteoTimeout,
				)
			},
		},
		{
			name: "analytical metric routes",
			register: func() error {
				return registerAnalyticalMetricRoutes(
					v1,
					dbPool,
				)
			},
		},
		{
			name: "Airport Intelligence routes",
			register: func() error {
				return registerAirportIntelligenceRoutes(
					v1,
					dbPool,
				)
			},
		},
		{
			name: "Historical Intelligence routes",
			register: func() error {
				return registerHistoricalIntelligenceRoutes(
					v1,
					dbPool,
				)
			},
		},
		{
			name: "projection-dependent intelligence routes",
			register: func() error {
				return registerProjectionDatabaseContext(
					v1,
					dbPool,
				)
			},
		},
		{
			name: "Airspace Region Analytics route",
			register: func() error {
				return registerAirspaceDatabaseContext(
					v1,
					dbPool,
				)
			},
		},
	}
}

func registerDatabaseRouteGroups(
	groups []databaseRouteGroup,
) error {
	for _, group := range groups {
		if group.register == nil {
			return fmt.Errorf(
				"register %s: route group function is required",
				group.name,
			)
		}
		if err := group.register(); err != nil {
			return fmt.Errorf(
				"register %s: %w",
				group.name,
				err,
			)
		}
	}

	return nil
}

// STAGE-14-8-SERVER-COMPOSITION-ROOT-DECOMPOSITION

// STAGE-14-10-TRANSPONDER-EVIDENCE-PRODUCTION
