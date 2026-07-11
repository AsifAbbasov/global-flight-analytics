package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/metrics"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/region"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

type MetricsHandler struct {
	service *metrics.Service
}

func NewMetricsHandler(
	service *metrics.Service,
) *MetricsHandler {
	return &MetricsHandler{
		service: service,
	}
}

func (
	h *MetricsHandler,
) GetActiveAircraft(
	c *fiber.Ctx,
) error {
	windowMinutes, err := parseActiveAircraftWindowMinutes(
		c.Query("window_minutes"),
	)
	if err != nil {
		return response.Error(
			c,
			fiber.StatusBadRequest,
			"INVALID_WINDOW_MINUTES",
			"window_minutes must be an integer between 1 and 180",
		)
	}

	metric, err := h.service.CalculateActiveAircraft(
		c.Context(),
		metrics.ActiveAircraftRequest{
			RegionCode:    c.Query("region"),
			WindowMinutes: windowMinutes,
		},
	)
	if err != nil {
		if errors.Is(
			err,
			region.ErrRegionNotFound,
		) {
			return response.Error(
				c,
				fiber.StatusNotFound,
				"REGION_NOT_FOUND",
				"Region not found",
			)
		}

		if errors.Is(
			err,
			metrics.ErrInvalidWindowMinutes,
		) {
			return response.Error(
				c,
				fiber.StatusBadRequest,
				"INVALID_WINDOW_MINUTES",
				"window_minutes must be an integer between 1 and 180",
			)
		}

		return response.Error(
			c,
			fiber.StatusInternalServerError,
			"ACTIVE_AIRCRAFT_METRIC_FAILED",
			"Failed to calculate active aircraft metric",
		)
	}

	return response.OK(
		c,
		toActiveAircraftMetricResponse(
			metric,
		),
	)
}

func parseActiveAircraftWindowMinutes(
	rawValue string,
) (int, error) {
	trimmed := strings.TrimSpace(
		rawValue,
	)
	if trimmed == "" {
		return metrics.DefaultActiveAircraftWindowMinutes,
			nil
	}

	value, err := strconv.Atoi(
		trimmed,
	)
	if err != nil {
		return 0,
			err
	}

	if value < metrics.MinimumActiveAircraftWindowMinutes ||
		value > metrics.MaximumActiveAircraftWindowMinutes {
		return 0,
			metrics.ErrInvalidWindowMinutes
	}

	return value,
		nil
}

func toActiveAircraftMetricResponse(
	metric metrics.ActiveAircraftMetric,
) dto.ActiveAircraftMetricResponse {
	return dto.ActiveAircraftMetricResponse{
		Metric:        string(metric.Metric),
		Value:         metric.Value,
		WindowMinutes: metric.WindowMinutes,
		Scope: dto.MetricScopeResponse{
			Type: string(metric.Scope.Type),
			Code: metric.Scope.Code,
		},
		ObservedFrom: metric.ObservedFrom,
		ObservedTo:   metric.ObservedTo,
		CalculatedAt: metric.CalculatedAt,
		Confidence: dto.MetricConfidenceResponse{
			Level: string(metric.Confidence.Level),
			Score: metric.Confidence.Score,
			Reasons: append(
				[]string{},
				metric.Confidence.Reasons...,
			),
		},
		Sources: toMetricSourceResponses(
			metric.Sources,
		),
		Limitations: append(
			[]string{},
			metric.Limitations...,
		),
	}
}

func toMetricSourceResponses(
	sources []metrics.MetricSource,
) []dto.MetricSourceResponse {
	result := make(
		[]dto.MetricSourceResponse,
		0,
		len(sources),
	)
	for _, source := range sources {
		result = append(
			result,
			dto.MetricSourceResponse{
				Name: source.Name,
				Role: source.Role,
			},
		)
	}

	return result
}
