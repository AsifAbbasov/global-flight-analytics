package handlers

import (
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

type AirportHandler struct {
	service *airport.Service
}

func NewAirportHandler(service *airport.Service) *AirportHandler {
	return &AirportHandler{
		service: service,
	}
}

func (h *AirportHandler) List(c *fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "AIRPORT_LIST_FAILED", "Failed to load airports")
	}

	return response.OK(c, toAirportListItems(items))
}

func (h *AirportHandler) GetByICAO(c *fiber.Ctx) error {
	icao := c.Params("icao")

	item, err := h.service.GetByICAO(c.Context(), icao)
	if err != nil {
		if errors.Is(err, airport.ErrNotFound) {
			return response.Error(c, fiber.StatusNotFound, "AIRPORT_NOT_FOUND", "Airport not found")
		}

		return response.Error(c, fiber.StatusInternalServerError, "AIRPORT_LOAD_FAILED", "Failed to load airport")
	}

	return response.OK(c, toAirportProfile(item))
}

func toAirportListItems(items []airport.Airport) []dto.AirportListItem {
	result := make([]dto.AirportListItem, 0, len(items))

	for _, item := range items {
		result = append(result, dto.AirportListItem{
			ICAOCode:  item.ICAOCode,
			IATACode:  item.IATACode,
			Name:      item.Name,
			City:      item.City,
			Country:   item.Country,
			Latitude:  item.Latitude,
			Longitude: item.Longitude,
		})
	}

	return result
}

func toAirportProfile(item airport.Airport) dto.AirportProfile {
	elevationM, elevationStatus := dto.ToAirportElevation(
		item.ElevationM,
		item.ElevationAvailable,
	)

	return dto.AirportProfile{
		ICAOCode:        item.ICAOCode,
		IATACode:        item.IATACode,
		Name:            item.Name,
		City:            item.City,
		Country:         item.Country,
		Latitude:        item.Latitude,
		Longitude:       item.Longitude,
		ElevationM:      elevationM,
		ElevationStatus: elevationStatus,
		Timezone:        item.Timezone,
		Description:     item.Description,
	}
}
