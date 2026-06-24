package handlers

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/traffic"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

type TrafficHandler struct {
	service *traffic.Service
}

func NewTrafficHandler(service *traffic.Service) *TrafficHandler {
	return &TrafficHandler{
		service: service,
	}
}

func (h *TrafficHandler) GetCurrent(c *fiber.Ctx) error {
	items, err := h.service.GetCurrent(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CURRENT_TRAFFIC_LOAD_FAILED", "Failed to load current traffic")
	}

	return response.OK(c, toCurrentTrafficItems(items))
}

func toCurrentTrafficItems(items []traffic.CurrentTrafficItem) []dto.CurrentTrafficItem {
	result := make([]dto.CurrentTrafficItem, 0, len(items))

	for _, item := range items {
		result = append(result, dto.CurrentTrafficItem{
			ICAO24:         item.ICAO24,
			Callsign:       item.Callsign,
			Latitude:       item.Latitude,
			Longitude:      item.Longitude,
			AltitudeM:      item.AltitudeM,
			VelocityMPS:    item.VelocityMPS,
			HeadingDegrees: item.HeadingDegrees,
			OnGround:       item.OnGround,
			ObservedAt:     item.ObservedAt,
			AircraftModel:  item.AircraftModel,
			Airline:        item.Airline,
			OriginCountry:  item.OriginCountry,
		})
	}

	return result
}
