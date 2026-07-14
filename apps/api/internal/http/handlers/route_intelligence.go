package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routepipeline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	routeIntelligenceHistoryLimitQuery  = "limit"
	routeIntelligenceHistoryBeforeQuery = "before_as_of_time"
)

type routeIntelligencePipeline interface {
	Process(context.Context, routepipeline.Request) (routepipeline.Result, error)
}

type routeIntelligenceStore interface {
	GetLatest(context.Context, string, routecontract.SchemaVersion) (routestore.Record, error)
	List(context.Context, routestore.ListQuery) (routestore.Page, error)
}

type RouteIntelligenceHandler struct {
	pipeline routeIntelligencePipeline
	store    routeIntelligenceStore
}

func NewRouteIntelligenceHandler(pipeline routeIntelligencePipeline, store routeIntelligenceStore) *RouteIntelligenceHandler {
	return &RouteIntelligenceHandler{pipeline: pipeline, store: store}
}

func (h *RouteIntelligenceHandler) ProcessByTrajectoryID(c *fiber.Ctx) error {
	if h.pipeline == nil {
		return routeIntelligenceServiceUnavailable(c)
	}
	id, err := parseRouteTrajectoryID(c.Params("id"))
	if err != nil {
		return routeIntelligenceRequestError(c, err)
	}
	result, err := h.pipeline.Process(c.Context(), routepipeline.Request{TrajectoryID: id})
	if err != nil {
		return writeRouteIntelligenceError(c, err, "ROUTE_INTELLIGENCE_PROCESS_FAILED", "Failed to process Route Intelligence")
	}
	return response.OK(c, dto.ToRouteIntelligenceRecord(result.Record))
}

func (h *RouteIntelligenceHandler) GetLatestByTrajectoryID(c *fiber.Ctx) error {
	if h.store == nil {
		return routeIntelligenceServiceUnavailable(c)
	}
	id, err := parseRouteTrajectoryID(c.Params("id"))
	if err != nil {
		return routeIntelligenceRequestError(c, err)
	}
	record, err := h.store.GetLatest(c.Context(), id, routecontract.SchemaVersionV1)
	if err != nil {
		return writeRouteIntelligenceError(c, err, "ROUTE_INTELLIGENCE_LOAD_FAILED", "Failed to load the latest Route Intelligence result")
	}
	return response.OK(c, dto.ToRouteIntelligenceRecord(record))
}

func (h *RouteIntelligenceHandler) ListHistoryByTrajectoryID(c *fiber.Ctx) error {
	if h.store == nil {
		return routeIntelligenceServiceUnavailable(c)
	}
	id, err := parseRouteTrajectoryID(c.Params("id"))
	if err != nil {
		return routeIntelligenceRequestError(c, err)
	}
	query, err := parseRouteIntelligenceHistoryQuery(id, c.Query(routeIntelligenceHistoryLimitQuery), c.Query(routeIntelligenceHistoryBeforeQuery))
	if err != nil {
		return routeIntelligenceRequestError(c, err)
	}
	page, err := h.store.List(c.Context(), query)
	if err != nil {
		return writeRouteIntelligenceError(c, err, "ROUTE_INTELLIGENCE_HISTORY_LOAD_FAILED", "Failed to load Route Intelligence history")
	}
	return response.OK(c, dto.ToRouteIntelligenceHistory(page))
}

func parseRouteTrajectoryID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", routestore.ErrTrajectoryIDRequired
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return "", routestore.ErrInvalidTrajectoryID
	}
	return strings.ToLower(parsed.String()), nil
}

func parseRouteIntelligenceHistoryQuery(id, limitValue, beforeValue string) (routestore.ListQuery, error) {
	limit := routestore.DefaultListLimit
	if strings.TrimSpace(limitValue) != "" {
		n, err := strconv.Atoi(strings.TrimSpace(limitValue))
		if err != nil || n < 1 || n > routestore.MaximumListLimit {
			return routestore.ListQuery{}, routestore.ErrInvalidListLimit
		}
		limit = n
	}
	var before time.Time
	if strings.TrimSpace(beforeValue) != "" {
		parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(beforeValue))
		if err != nil {
			return routestore.ListQuery{}, routestore.ErrAsOfTimeRequired
		}
		before = parsed.UTC()
	}
	return routestore.ListQuery{TrajectoryID: id, SchemaVersion: routecontract.SchemaVersionV1, BeforeAsOfTime: before, Limit: limit}, nil
}

func routeIntelligenceRequestError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, routestore.ErrInvalidListLimit):
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ROUTE_INTELLIGENCE_LIMIT", "Route Intelligence history limit must be between one and one hundred")
	case errors.Is(err, routestore.ErrAsOfTimeRequired):
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ROUTE_INTELLIGENCE_CURSOR", "Route Intelligence history cursor must be a valid RFC 3339 timestamp")
	default:
		return response.Error(c, fiber.StatusBadRequest, "INVALID_TRAJECTORY_ID", "Trajectory identifier must be a valid UUID")
	}
}

func routeIntelligenceServiceUnavailable(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusServiceUnavailable, "ROUTE_INTELLIGENCE_SERVICE_UNAVAILABLE", "Route Intelligence service is unavailable")
}

func writeRouteIntelligenceError(c *fiber.Ctx, err error, defaultCode, defaultMessage string) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return response.Error(c, fiber.StatusGatewayTimeout, "ROUTE_INTELLIGENCE_TIMEOUT", "Route Intelligence request timed out")
	case errors.Is(err, context.Canceled):
		return response.Error(c, fiber.StatusRequestTimeout, "ROUTE_INTELLIGENCE_REQUEST_CANCELED", "Route Intelligence request was canceled")
	case errors.Is(err, pgx.ErrNoRows), errors.Is(err, routestore.ErrResultNotFound):
		return response.Error(c, fiber.StatusNotFound, "ROUTE_INTELLIGENCE_NOT_FOUND", "Route Intelligence is unavailable because no matching trajectory result was found")
	case errors.Is(err, routestore.ErrResultConflict):
		return response.Error(c, fiber.StatusConflict, "ROUTE_INTELLIGENCE_CONFLICT", "Route Intelligence result conflicts with previously stored evidence")
	case errors.Is(err, routestore.ErrInvalidTrajectoryID), errors.Is(err, routestore.ErrTrajectoryIDRequired), errors.Is(err, routepipeline.ErrTrajectoryIDRequired):
		return response.Error(c, fiber.StatusBadRequest, "INVALID_TRAJECTORY_ID", "Trajectory identifier must be a valid UUID")
	case errors.Is(err, routepipeline.ErrTrajectoryReaderRequired), errors.Is(err, routepipeline.ErrAirportListerRequired), errors.Is(err, routepipeline.ErrStoreRequired), errors.Is(err, routepipeline.ErrPostgresPoolRequired), errors.Is(err, routestore.ErrPostgresPoolRequired), errors.Is(err, routestore.ErrPostgresExecutorRequired):
		return routeIntelligenceServiceUnavailable(c)
	default:
		return response.Error(c, fiber.StatusInternalServerError, defaultCode, defaultMessage)
	}
}
