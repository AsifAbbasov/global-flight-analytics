package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/stabilityintelligence/stabilityproduction"
	"github.com/gofiber/fiber/v2"
)

const StabilityIntelligencePath = "/trajectories/:id/stability-intelligence"

func RegisterStabilityIntelligenceReadRoute(
	v1 fiber.Router,
	reader handlers.StabilityIntelligenceReader,
) error {
	if reader == nil {
		return fmt.Errorf(
			"Stability Intelligence reader is required",
		)
	}

	handler :=
		handlers.NewStabilityIntelligenceHandler(
			reader,
		)
	v1.Get(
		StabilityIntelligencePath,
		handler.GetByTrajectoryID,
	)

	return nil
}

type stabilityProjectionReaderAdapter struct {
	reader handlers.ProjectionIntelligenceReader
}

func (
	adapter stabilityProjectionReaderAdapter,
) ReadProjection(
	ctx context.Context,
	request stabilityproduction.ProjectionRequest,
) (
	projectionproduction.Result,
	error,
) {
	if adapter.reader == nil {
		return projectionproduction.Result{},
			stabilityproduction.
				ErrServiceUnavailable
	}

	result, err :=
		adapter.reader.GetProjectionIntelligence(
			ctx,
			handlers.ProjectionIntelligenceReadRequest{
				TrajectoryID:      request.TrajectoryID,
				AsOfTime:          request.AsOfTime,
				RequestedDuration: request.RequestedDuration,
			},
		)
	if err != nil {
		switch {
		case errors.Is(
			err,
			handlers.
				ErrProjectionIntelligenceNotFound,
		):
			return projectionproduction.Result{},
				stabilityproduction.
					ErrTrajectoryNotFound

		case errors.Is(
			err,
			handlers.
				ErrProjectionIntelligenceServiceUnavailable,
		):
			return projectionproduction.Result{},
				stabilityproduction.
					ErrServiceUnavailable

		case errors.Is(
			err,
			handlers.
				ErrProjectionIntelligenceInvalidRequest,
		):
			return projectionproduction.Result{},
				stabilityproduction.
					ErrInvalidRequest

		default:
			return projectionproduction.Result{},
				err
		}
	}

	return result.Clone(), nil
}
