package handlers

import (
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flight"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/gofiber/fiber/v2"
)

type FlightHandler struct {
	service *flight.Service
}

func NewFlightHandler(service *flight.Service) *FlightHandler {
	return &FlightHandler{
		service: service,
	}
}

func (h *FlightHandler) List(c *fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FLIGHT_LIST_FAILED", "Failed to load flights")
	}

	return response.OK(c, toFlightListItems(items))
}

func (h *FlightHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	item, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, postgres.ErrFlightNotFound) {
			return response.Error(c, fiber.StatusNotFound, "FLIGHT_NOT_FOUND", "Flight not found")
		}

		return response.Error(c, fiber.StatusInternalServerError, "FLIGHT_LOAD_FAILED", "Failed to load flight")
	}

	return response.OK(c, toFlightProfile(item))
}

func toFlightListItems(items []flight.Flight) []dto.FlightListItem {
	result := make([]dto.FlightListItem, 0, len(items))

	for _, item := range items {
		result = append(result, dto.FlightListItem{
			ID:            item.ID,
			AircraftID:    item.AircraftID,
			ICAO24:        item.ICAO24,
			Callsign:      item.Callsign,
			Status:        item.Status,
			FirstSeenAt:   item.FirstSeenAt,
			LastSeenAt:    item.LastSeenAt,
			AircraftModel: item.AircraftModel,
			Airline:       item.Airline,
			Country:       item.Country,
		})
	}

	return result
}

func toFlightProfile(item flight.Flight) dto.FlightProfile {
	return dto.FlightProfile{
		ID:            item.ID,
		AircraftID:    item.AircraftID,
		ICAO24:        item.ICAO24,
		Callsign:      item.Callsign,
		Status:        item.Status,
		FirstSeenAt:   item.FirstSeenAt,
		LastSeenAt:    item.LastSeenAt,
		AircraftModel: item.AircraftModel,
		Airline:       item.Airline,
		Country:       item.Country,
	}
}
