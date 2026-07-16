package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceproduction"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/airspaceintelligence/airspaceregionanalytics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

const (
	airspaceAnalyticsAsOfTimeQuery        = "as_of_time"
	airspaceAnalyticsWindowSecondsQuery   = "window_seconds"
	defaultAirspaceAnalyticsWindowSeconds = 300
	minimumAirspaceAnalyticsWindowSeconds = 60
	maximumAirspaceAnalyticsWindowSeconds = 3600
)

var (
	errAirspaceAnalyticsRegionRequired = errors.New(
		"airspace analytics region code is required",
	)
	errAirspaceAnalyticsAsOfTimeInvalid = errors.New(
		"airspace analytics as-of time is invalid",
	)
	errAirspaceAnalyticsWindowInvalid = errors.New(
		"airspace analytics window is invalid",
	)
)

type AirspaceRegionAnalyticsReader interface {
	GetAirspaceRegionAnalytics(
		context.Context,
		airspaceproduction.Request,
	) (airspaceregionanalytics.Result, error)
}

type AirspaceRegionAnalyticsHandler struct {
	reader AirspaceRegionAnalyticsReader
}

func NewAirspaceRegionAnalyticsHandler(
	reader AirspaceRegionAnalyticsReader,
) *AirspaceRegionAnalyticsHandler {
	return &AirspaceRegionAnalyticsHandler{
		reader: reader,
	}
}

func (handler *AirspaceRegionAnalyticsHandler) GetByRegionCode(
	ctx *fiber.Ctx,
) error {
	if handler == nil || handler.reader == nil {
		return response.Error(
			ctx,
			fiber.StatusServiceUnavailable,
			"AIRSPACE_ANALYTICS_SERVICE_UNAVAILABLE",
			"Airspace Region Analytics service is unavailable",
		)
	}

	request, err := parseAirspaceRegionAnalyticsRequest(
		ctx.Params("code"),
		ctx.Query(airspaceAnalyticsAsOfTimeQuery),
		ctx.Query(airspaceAnalyticsWindowSecondsQuery),
	)
	if err != nil {
		return writeAirspaceRegionAnalyticsRequestError(ctx, err)
	}

	result, err := handler.reader.GetAirspaceRegionAnalytics(
		ctx.Context(),
		request,
	)
	if err != nil {
		return writeAirspaceRegionAnalyticsError(ctx, err)
	}

	report := airspaceregionanalytics.Validate(
		result,
		airspaceregionanalytics.DefaultPolicy(),
	)
	if report.Status != airspaceregionanalytics.ValidationStatusValid {
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"AIRSPACE_ANALYTICS_CONTRACT_INVALID",
			"Airspace Region Analytics service returned an invalid result",
		)
	}

	return response.OK(
		ctx,
		dto.ToAirspaceRegionAnalyticsResponse(result),
	)
}

func parseAirspaceRegionAnalyticsRequest(
	regionCodeValue string,
	asOfTimeValue string,
	windowSecondsValue string,
) (airspaceproduction.Request, error) {
	regionCode := strings.ToLower(
		strings.TrimSpace(regionCodeValue),
	)
	if regionCode == "" {
		return airspaceproduction.Request{},
			errAirspaceAnalyticsRegionRequired
	}

	asOfTime, err := time.Parse(
		time.RFC3339Nano,
		strings.TrimSpace(asOfTimeValue),
	)
	if err != nil || strings.TrimSpace(asOfTimeValue) == "" {
		return airspaceproduction.Request{},
			errAirspaceAnalyticsAsOfTimeInvalid
	}

	windowSeconds := int64(defaultAirspaceAnalyticsWindowSeconds)
	if normalized := strings.TrimSpace(windowSecondsValue); normalized != "" {
		windowSeconds, err = strconv.ParseInt(normalized, 10, 64)
		if err != nil {
			return airspaceproduction.Request{},
				errAirspaceAnalyticsWindowInvalid
		}
	}
	if windowSeconds < minimumAirspaceAnalyticsWindowSeconds ||
		windowSeconds > maximumAirspaceAnalyticsWindowSeconds ||
		windowSeconds%60 != 0 {
		return airspaceproduction.Request{},
			errAirspaceAnalyticsWindowInvalid
	}

	return airspaceproduction.Request{
		RegionCode: regionCode,
		AsOfTime:   asOfTime.UTC(),
		Window:     time.Duration(windowSeconds) * time.Second,
	}, nil
}

func writeAirspaceRegionAnalyticsRequestError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(err, errAirspaceAnalyticsRegionRequired):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_AIRSPACE_ANALYTICS_REGION",
			"Airspace Region Analytics region code is required",
		)
	case errors.Is(err, errAirspaceAnalyticsAsOfTimeInvalid):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_AIRSPACE_ANALYTICS_AS_OF_TIME",
			"Airspace Region Analytics as-of time is required and must be a valid RFC 3339 timestamp",
		)
	case errors.Is(err, errAirspaceAnalyticsWindowInvalid):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_AIRSPACE_ANALYTICS_WINDOW",
			"Airspace Region Analytics window must be a whole number of minutes between 60 and 3600 seconds",
		)
	default:
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_AIRSPACE_ANALYTICS_REQUEST",
			"Airspace Region Analytics request is invalid",
		)
	}
}

func writeAirspaceRegionAnalyticsError(
	ctx *fiber.Ctx,
	err error,
) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return response.Error(
			ctx,
			fiber.StatusGatewayTimeout,
			"AIRSPACE_ANALYTICS_TIMEOUT",
			"Airspace Region Analytics request timed out",
		)
	case errors.Is(err, context.Canceled):
		return response.Error(
			ctx,
			fiber.StatusRequestTimeout,
			"AIRSPACE_ANALYTICS_REQUEST_CANCELED",
			"Airspace Region Analytics request was canceled",
		)
	case errors.Is(err, region.ErrRegionNotFound):
		return response.Error(
			ctx,
			fiber.StatusNotFound,
			"AIRSPACE_ANALYTICS_REGION_NOT_FOUND",
			"The requested Airspace Region Analytics region was not found",
		)
	case errors.Is(err, airspaceproduction.ErrInvalidRequest):
		return response.Error(
			ctx,
			fiber.StatusBadRequest,
			"INVALID_AIRSPACE_ANALYTICS_REQUEST",
			"Airspace Region Analytics request is outside the configured production policy",
		)
	case errors.Is(err, airspaceproduction.ErrObservationCapacityExceeded):
		return response.Error(
			ctx,
			fiber.StatusUnprocessableEntity,
			"AIRSPACE_ANALYTICS_CAPACITY_EXCEEDED",
			"Airspace Region Analytics observation capacity was exceeded",
		)
	default:
		return response.Error(
			ctx,
			fiber.StatusInternalServerError,
			"AIRSPACE_ANALYTICS_LOAD_FAILED",
			"Failed to build Airspace Region Analytics",
		)
	}
}
