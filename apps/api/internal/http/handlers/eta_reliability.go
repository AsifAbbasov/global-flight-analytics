package handlers

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/etareliability"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/projectionintelligence/projectionread"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type ETAReliabilityReader interface {
	GetETAReliability(
		context.Context,
		ProjectionIntelligenceReadRequest,
	) (etareliability.Result, error)
}

type ETAReliabilityHandler struct {
	reader ETAReliabilityReader
}

func NewETAReliabilityHandler(reader ETAReliabilityReader) *ETAReliabilityHandler {
	return &ETAReliabilityHandler{reader: reader}
}

func (handler *ETAReliabilityHandler) GetByTrajectoryID(ctx *fiber.Ctx) error {
	if handler == nil || handler.reader == nil {
		return response.Error(ctx, fiber.StatusServiceUnavailable, "ETA_RELIABILITY_SERVICE_UNAVAILABLE", "ETA Reliability service is unavailable")
	}
	request, err := parseProjectionIntelligenceReadRequest(
		ctx.Params("id"),
		ctx.Query(projectionIntelligenceAsOfTimeQuery),
		ctx.Query(projectionIntelligenceDurationSecondsQuery),
	)
	if err != nil {
		return projectionIntelligenceRequestError(ctx, err)
	}
	result, err := handler.reader.GetETAReliability(ctx.UserContext(), request)
	if err != nil {
		return writeETAReliabilityError(ctx, err)
	}
	if err := result.Validate(); err != nil {
		return response.Error(ctx, fiber.StatusInternalServerError, "ETA_RELIABILITY_CONTRACT_INVALID", "ETA Reliability service returned an invalid result")
	}
	return response.OK(ctx, dto.ToETAReliabilityResponse(result))
}

func writeETAReliabilityError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return response.Error(ctx, fiber.StatusGatewayTimeout, "ETA_RELIABILITY_TIMEOUT", "ETA Reliability request timed out")
	case errors.Is(err, context.Canceled):
		return response.Error(ctx, fiber.StatusRequestTimeout, "ETA_RELIABILITY_REQUEST_CANCELED", "ETA Reliability request was canceled")
	case errors.Is(err, pgx.ErrNoRows), errors.Is(err, projectionread.ErrTrajectoryNotFound):
		return response.Error(ctx, fiber.StatusNotFound, "ETA_RELIABILITY_NOT_FOUND", "No matching trajectory was available for ETA Reliability")
	case errors.Is(err, etareliability.ErrServiceUnavailable):
		return response.Error(ctx, fiber.StatusServiceUnavailable, "ETA_RELIABILITY_SERVICE_UNAVAILABLE", "ETA Reliability service is unavailable")
	case errors.Is(err, etareliability.ErrInvalidRequest):
		return response.Error(ctx, fiber.StatusBadRequest, "INVALID_ETA_RELIABILITY_REQUEST", "ETA Reliability request is invalid")
	case errors.Is(err, etareliability.ErrCurrentETAUnavailable), errors.Is(err, etareliability.ErrRouteUnavailable), errors.Is(err, etareliability.ErrLeadOutsidePolicy):
		return response.Error(ctx, fiber.StatusUnprocessableEntity, "ETA_RELIABILITY_UNAVAILABLE", "ETA Reliability is unavailable for the current trajectory evidence")
	default:
		return response.Error(ctx, fiber.StatusInternalServerError, "ETA_RELIABILITY_LOAD_FAILED", "Failed to load ETA Reliability")
	}
}
