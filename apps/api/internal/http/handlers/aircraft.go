package handlers

import (
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/aircraft"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

type AircraftHandler struct {
	service *aircraft.Service
}

func NewAircraftHandler(service *aircraft.Service) *AircraftHandler {
	return &AircraftHandler{
		service: service,
	}
}

func (h *AircraftHandler) List(c *fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "AIRCRAFT_LIST_FAILED", "Failed to load aircraft")
	}

	return response.OK(c, toAircraftListItems(items))
}

func (h *AircraftHandler) GetByICAO24(c *fiber.Ctx) error {
	icao24 := c.Params("icao24")

	item, err := h.service.GetByICAO24(c.Context(), icao24)
	if err != nil {
		if errors.Is(err, aircraft.ErrNotFound) {
			return response.Error(c, fiber.StatusNotFound, "AIRCRAFT_NOT_FOUND", "Aircraft not found")
		}

		return response.Error(c, fiber.StatusInternalServerError, "AIRCRAFT_LOAD_FAILED", "Failed to load aircraft")
	}

	return response.OK(c, toAircraftProfile(item))
}

func toAircraftListItems(items []aircraft.Aircraft) []dto.AircraftListItem {
	result := make([]dto.AircraftListItem, 0, len(items))

	for _, item := range items {
		result = append(result, dto.AircraftListItem{
			ICAO24:       item.ICAO24,
			Registration: item.Registration,
			Model:        item.Model,
			Manufacturer: item.Manufacturer,
			Airline:      item.Airline,
			Country:      item.Country,
		})
	}

	return result
}

func toAircraftProfile(item aircraft.Aircraft) dto.AircraftProfile {
	return dto.AircraftProfile{
		ICAO24:       item.ICAO24,
		Registration: item.Registration,
		Model:        item.Model,
		Manufacturer: item.Manufacturer,
		AircraftType: item.AircraftType,
		Airline:      item.Airline,
		Country:      item.Country,
	}
}
