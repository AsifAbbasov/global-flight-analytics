package handlers

import (
	"context"
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	trafficquery "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/query"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/traffic/routecontext"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type routeContextService interface {
	GetByICAO24(
		ctx context.Context,
		icao24 string,
	) (routecontext.Context, error)
}

type RouteContextHandler struct {
	service routeContextService
}

func NewRouteContextHandler(
	service routeContextService,
) *RouteContextHandler {
	return &RouteContextHandler{
		service: service,
	}
}

func (handler *RouteContextHandler) GetByICAO24(
	ctx *fiber.Ctx,
) error {
	if handler.service == nil {
		return response.Error(
			ctx,
			fiber.StatusServiceUnavailable,
			"ROUTE_CONTEXT_SERVICE_UNAVAILABLE",
			"Route context service is unavailable",
		)
	}

	item, err := handler.service.GetByICAO24(
		ctx.Context(),
		ctx.Params("icao24"),
	)
	if err != nil {
		switch {
		case errors.Is(err, routecontext.ErrInvalidICAO24):
			return response.Error(
				ctx,
				fiber.StatusBadRequest,
				"INVALID_ICAO24",
				"Invalid ICAO24",
			)

		case errors.Is(
			err,
			routecontext.ErrTrajectoryReaderRequired,
		), errors.Is(
			err,
			routecontext.ErrAirportListerRequired,
		), errors.Is(
			err,
			trafficquery.ErrTrajectoryRepositoryRequired,
		):
			return response.Error(
				ctx,
				fiber.StatusServiceUnavailable,
				"ROUTE_CONTEXT_SERVICE_UNAVAILABLE",
				"Route context service is unavailable",
			)

		case errors.Is(err, pgx.ErrNoRows):
			return response.Error(
				ctx,
				fiber.StatusNotFound,
				"ROUTE_CONTEXT_NOT_FOUND",
				"Route context is unavailable because no trajectory was found",
			)

		default:
			return response.Error(
				ctx,
				fiber.StatusInternalServerError,
				"ROUTE_CONTEXT_LOAD_FAILED",
				"Failed to load route context",
			)
		}
	}

	return response.OK(
		ctx,
		dto.ToAircraftRouteContext(item),
	)
}
