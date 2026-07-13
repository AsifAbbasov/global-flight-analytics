package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/executor"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricexecution"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/metricquery"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerAnalyticalMetricRoutes(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
) error {
	trajectoryRepository := postgres.NewTrajectoryRepository(
		dbPool,
	)

	queryService, err := metricquery.New(
		trajectoryRepository,
	)
	if err != nil {
		return fmt.Errorf(
			"create analytical trajectory query service: %w",
			err,
		)
	}

	metricService, err := metricexecution.New(
		executor.New(nil),
	)
	if err != nil {
		return fmt.Errorf(
			"create protected analytical metric service: %w",
			err,
		)
	}

	handler, err := handlers.NewAnalyticalMetricsHandler(
		metricService,
		queryService,
	)
	if err != nil {
		return fmt.Errorf(
			"create analytical metrics handler: %w",
			err,
		)
	}

	routes := v1.Group(
		"/analytics",
	).Group(
		"/metrics",
	)

	routes.Get(
		"/active-aircraft",
		handler.GetActiveAircraft,
	)
	routes.Get(
		"/traffic-density",
		handler.GetTrafficDensity,
	)
	routes.Get(
		"/airport-activity",
		handler.GetAirportActivity,
	)
	routes.Get(
		"/coverage-score",
		handler.GetCoverageScore,
	)
	routes.Get(
		"/data-freshness",
		handler.GetDataFreshness,
	)

	return nil
}
