package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/transponderalert"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transponderEvidenceDatabaseRuntime struct {
	handler *handlers.TransponderEvidenceHandler
}

func registerTransponderEvidenceDatabaseContext(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
) error {
	runtime, err := composeTransponderEvidenceDatabaseRuntime(
		dbPool,
	)
	if err != nil {
		return err
	}

	registerTransponderEvidenceDatabaseRoutes(
		v1,
		runtime,
	)
	return nil
}

func composeTransponderEvidenceDatabaseRuntime(
	dbPool *pgxpool.Pool,
) (
	transponderEvidenceDatabaseRuntime,
	error,
) {
	flightStateService := flightstate.MustNewService(
		postgres.NewFlightStateRepository(
			dbPool,
		),
	)
	service, err := transponderalert.NewService(
		transponderalert.ServiceConfig{
			LatestStateReader: flightStateService,
		},
	)
	if err != nil {
		return transponderEvidenceDatabaseRuntime{},
			fmt.Errorf(
				"compose production transponder evidence service: %w",
				err,
			)
	}

	return transponderEvidenceDatabaseRuntime{
		handler: handlers.NewTransponderEvidenceHandler(
			service,
		),
	}, nil
}
