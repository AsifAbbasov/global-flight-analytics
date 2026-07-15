package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/jackc/pgx/v5/pgxpool"
)

type projectionIntelligenceApplicationReader interface {
	Get(
		context.Context,
		projectionread.Request,
	) (projectionproduction.Result, error)
}

type projectionIntelligenceReaderAdapter struct {
	reader projectionIntelligenceApplicationReader
}

func (
	adapter projectionIntelligenceReaderAdapter,
) GetProjectionIntelligence(
	ctx context.Context,
	request handlers.ProjectionIntelligenceReadRequest,
) (projectionproduction.Result, error) {
	if adapter.reader == nil {
		return projectionproduction.Result{},
			handlers.
				ErrProjectionIntelligenceServiceUnavailable
	}

	result, err := adapter.reader.Get(
		ctx,
		projectionread.Request{
			TrajectoryID:      request.TrajectoryID,
			AsOfTime:          request.AsOfTime,
			RequestedDuration: request.RequestedDuration,
		},
	)
	if err != nil {
		return projectionproduction.Result{},
			mapProjectionIntelligenceReadError(
				err,
			)
	}

	return result.Clone(), nil
}

func newProjectionIntelligencePostgresReader(
	pool *pgxpool.Pool,
) (
	handlers.ProjectionIntelligenceReader,
	error,
) {
	composition, err :=
		projectionread.NewPostgres(
			projectionread.PostgresConfig{
				Pool: pool,
				Policy: projectionread.
					DefaultPolicy(),
			},
		)
	if err != nil {
		return nil,
			fmt.Errorf(
				"compose PostgreSQL Projection Intelligence read service: %w",
				err,
			)
	}

	return projectionIntelligenceReaderAdapter{
			reader: composition.Service,
		},
		nil
}

func mapProjectionIntelligenceReadError(
	err error,
) error {
	switch {
	case errors.Is(
		err,
		projectionread.ErrTrajectoryNotFound,
	):
		return handlers.
			ErrProjectionIntelligenceNotFound

	case errors.Is(
		err,
		projectionread.ErrServiceUnavailable,
	):
		return handlers.
			ErrProjectionIntelligenceServiceUnavailable

	case errors.Is(
		err,
		projectionread.ErrInvalidRequest,
	):
		return handlers.
			ErrProjectionIntelligenceInvalidRequest

	default:
		return err
	}
}
