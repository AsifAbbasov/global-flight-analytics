package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/historicalintelligence/historicalaggregate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	HistoricalIntelligenceLatestPath  = "/historical-intelligence/aggregates/latest"
	HistoricalIntelligenceHistoryPath = "/historical-intelligence/aggregates/history"
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

	return RegisterHistoricalIntelligenceReadRoutes(
		v1,
		store,
	)
}

// RegisterHistoricalIntelligenceReadRoutes composes the read-only Historical
// Intelligence endpoints with an already constructed aggregate store. The
// production server supplies a PostgreSQL pool-backed store, while runtime
// verification may safely supply a rollback-only transaction-backed store.
func RegisterHistoricalIntelligenceReadRoutes(
	v1 fiber.Router,
	store historicalaggregate.Store,
) error {
	if store == nil {
		return fmt.Errorf(
			"Historical Intelligence aggregate store is required",
		)
	}

	handler :=
		handlers.NewHistoricalIntelligenceHandler(
			store,
		)

	v1.Get(
		HistoricalIntelligenceLatestPath,
		handler.GetLatest,
	)
	v1.Get(
		HistoricalIntelligenceHistoryPath,
		handler.ListHistory,
	)

	return nil
}
