package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airportintelligence/airportproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

const (
	airportIntelligenceAsOfTimeQuery       = "as_of_time"
	airportIntelligenceDaysQuery           = "days"
	airportIntelligenceLimitQuery          = "limit"
	defaultAirportIntelligenceRankingLimit = 50
	maximumAirportIntelligenceRankingLimit = 200
)

var (
	errAirportIntelligenceAsOfTimeInvalid = errors.New("Airport Intelligence as-of time is invalid")
	errAirportIntelligenceDaysInvalid     = errors.New("Airport Intelligence days value is invalid")
	errAirportIntelligenceLimitInvalid    = errors.New("Airport Intelligence limit is invalid")
)

type AirportIntelligenceHandler struct{ service airportproduction.ReadService }

func NewAirportIntelligenceHandler(service airportproduction.ReadService) *AirportIntelligenceHandler {
	return &AirportIntelligenceHandler{service: service}
}

func (handler *AirportIntelligenceHandler) GetOverview(ctx *fiber.Ctx) error {
	if handler == nil || handler.service == nil {
		return airportIntelligenceUnavailable(ctx)
	}
	request, err := parseAirportIntelligenceWindowRequest(ctx)
	if err != nil {
		return airportIntelligenceRequestError(ctx, err)
	}
	result, err := handler.service.GetOverview(ctx.Context(), ctx.Params("icao"), request)
	if err != nil {
		return writeAirportIntelligenceError(ctx, err)
	}
	return response.OK(ctx, dto.ToAirportIntelligenceOverviewResponse(result))
}
func (handler *AirportIntelligenceHandler) GetHistory(ctx *fiber.Ctx) error {
	if handler == nil || handler.service == nil {
		return airportIntelligenceUnavailable(ctx)
	}
	request, err := parseAirportIntelligenceWindowRequest(ctx)
	if err != nil {
		return airportIntelligenceRequestError(ctx, err)
	}
	result, err := handler.service.GetHistory(ctx.Context(), ctx.Params("icao"), request)
	if err != nil {
		return writeAirportIntelligenceError(ctx, err)
	}
	return response.OK(ctx, dto.ToAirportIntelligenceHistoryResponse(result))
}
func (handler *AirportIntelligenceHandler) GetTrends(ctx *fiber.Ctx) error {
	if handler == nil || handler.service == nil {
		return airportIntelligenceUnavailable(ctx)
	}
	request, err := parseAirportIntelligenceWindowRequest(ctx)
	if err != nil {
		return airportIntelligenceRequestError(ctx, err)
	}
	result, err := handler.service.GetTrends(ctx.Context(), ctx.Params("icao"), request)
	if err != nil {
		return writeAirportIntelligenceError(ctx, err)
	}
	return response.OK(ctx, dto.ToAirportIntelligenceTrendsResponse(result))
}
func (handler *AirportIntelligenceHandler) GetRanking(ctx *fiber.Ctx) error {
	if handler == nil || handler.service == nil {
		return airportIntelligenceUnavailable(ctx)
	}
	request, err := parseAirportIntelligenceWindowRequest(ctx)
	if err != nil {
		return airportIntelligenceRequestError(ctx, err)
	}
	limit, err := parseAirportIntelligenceRankingLimit(ctx.Query(airportIntelligenceLimitQuery))
	if err != nil {
		return airportIntelligenceRequestError(ctx, err)
	}
	result, err := handler.service.GetRanking(ctx.Context(), request)
	if err != nil {
		return writeAirportIntelligenceError(ctx, err)
	}
	return response.OK(ctx, dto.ToAirportIntelligenceRankingResponse(result, limit))
}
func parseAirportIntelligenceWindowRequest(ctx *fiber.Ctx) (airportproduction.WindowRequest, error) {
	days, err := parseAirportIntelligenceDays(ctx.Query(airportIntelligenceDaysQuery))
	if err != nil {
		return airportproduction.WindowRequest{}, err
	}
	asOfTime, err := parseAirportIntelligenceAsOfTime(ctx.Query(airportIntelligenceAsOfTimeQuery))
	if err != nil {
		return airportproduction.WindowRequest{}, err
	}
	return airportproduction.WindowRequest{AsOfTime: asOfTime, Days: days}, nil
}
func parseAirportIntelligenceDays(value string) (int, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return airportproduction.DefaultWindowDays, nil
	}
	days, err := strconv.Atoi(normalized)
	if err != nil || days < airportproduction.MinimumWindowDays || days > airportproduction.MaximumWindowDays {
		return 0, errAirportIntelligenceDaysInvalid
	}
	return days, nil
}
func parseAirportIntelligenceAsOfTime(value string) (time.Time, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, normalized)
	if err != nil {
		return time.Time{}, errAirportIntelligenceAsOfTimeInvalid
	}
	return parsed.UTC(), nil
}
func parseAirportIntelligenceRankingLimit(value string) (int, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return defaultAirportIntelligenceRankingLimit, nil
	}
	limit, err := strconv.Atoi(normalized)
	if err != nil || limit < 1 || limit > maximumAirportIntelligenceRankingLimit {
		return 0, errAirportIntelligenceLimitInvalid
	}
	return limit, nil
}
func airportIntelligenceRequestError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errAirportIntelligenceAsOfTimeInvalid):
		return response.Error(ctx, fiber.StatusBadRequest, "INVALID_AIRPORT_INTELLIGENCE_AS_OF_TIME", "Airport Intelligence as-of time must be a valid RFC 3339 timestamp")
	case errors.Is(err, errAirportIntelligenceDaysInvalid):
		return response.Error(ctx, fiber.StatusBadRequest, "INVALID_AIRPORT_INTELLIGENCE_DAYS", "Airport Intelligence days must be between 1 and 365")
	case errors.Is(err, errAirportIntelligenceLimitInvalid):
		return response.Error(ctx, fiber.StatusBadRequest, "INVALID_AIRPORT_INTELLIGENCE_LIMIT", "Airport Intelligence ranking limit must be between 1 and 200")
	default:
		return response.Error(ctx, fiber.StatusBadRequest, "INVALID_AIRPORT_INTELLIGENCE_REQUEST", "Airport Intelligence request is invalid")
	}
}
func writeAirportIntelligenceError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return response.Error(ctx, fiber.StatusGatewayTimeout, "AIRPORT_INTELLIGENCE_TIMEOUT", "Airport Intelligence request timed out")
	case errors.Is(err, context.Canceled):
		return response.Error(ctx, fiber.StatusRequestTimeout, "AIRPORT_INTELLIGENCE_REQUEST_CANCELED", "Airport Intelligence request was canceled")
	case errors.Is(err, airportproduction.ErrInvalidRequest):
		return response.Error(ctx, fiber.StatusBadRequest, "INVALID_AIRPORT_INTELLIGENCE_REQUEST", "Airport Intelligence request is outside the configured policy")
	case errors.Is(err, airport.ErrNotFound):
		return response.Error(ctx, fiber.StatusNotFound, "AIRPORT_NOT_FOUND", "Airport not found")
	case errors.Is(err, airportproduction.ErrObservationsNotFound):
		return response.Error(ctx, fiber.StatusNotFound, "AIRPORT_INTELLIGENCE_NOT_FOUND", "No Airport Intelligence observations were available for the requested completed-day window")
	case errors.Is(err, airportproduction.ErrInsufficientHistory):
		return response.Error(ctx, fiber.StatusUnprocessableEntity, "AIRPORT_INTELLIGENCE_HISTORY_INSUFFICIENT", "At least two observed daily windows are required for Airport Trends")
	default:
		return response.Error(ctx, fiber.StatusInternalServerError, "AIRPORT_INTELLIGENCE_LOAD_FAILED", "Failed to load Airport Intelligence")
	}
}
func airportIntelligenceUnavailable(ctx *fiber.Ctx) error {
	return response.Error(ctx, fiber.StatusServiceUnavailable, "AIRPORT_INTELLIGENCE_SERVICE_UNAVAILABLE", "Airport Intelligence service is unavailable")
}
