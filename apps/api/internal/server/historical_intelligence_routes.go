package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	historicalIntelligenceLatestPath  = "/historical-intelligence/aggregates/latest"
	historicalIntelligenceHistoryPath = "/historical-intelligence/aggregates/history"
)

func registerHistoricalIntelligenceRoutes(
	v1 fiber.Router,
	databasePool *pgxpool.Pool,
) error {
	store, err := historicalaggregate.NewPostgres(
		historicalaggregate.PostgresConfig{
			Pool: databasePool,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"compose Historical Intelligence aggregate store: %w",
			err,
		)
	}

	handler :=
		handlers.NewHistoricalIntelligenceHandler(
			store,
		)

	v1.Get(
		historicalIntelligenceLatestPath,
		handler.GetLatest,
	)
	v1.Get(
		historicalIntelligenceHistoryPath,
		handler.ListHistory,
	)

	return nil
}
