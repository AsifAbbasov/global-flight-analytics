package handlers

import (
	"errors"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/repository/postgres"
	"github.com/gofiber/fiber/v2"
)

type FlightStateHandler struct {
	service *flightstate.Service
}

func NewFlightStateHandler(service *flightstate.Service) *FlightStateHandler {
	return &FlightStateHandler{
		service: service,
	}
}

func (h *FlightStateHandler) ListByFlightID(c *fiber.Ctx) error {
	flightID := c.Params("flightID")

	items, err := h.service.ListByFlightID(c.Context(), flightID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FLIGHT_STATES_LIST_FAILED", "Failed to load flight states")
	}

	return response.OK(c, toFlightStateItems(items))
}

func (h *FlightStateHandler) GetLatestByICAO24(c *fiber.Ctx) error {
	icao24 := c.Params("icao24")

	item, err := h.service.GetLatestByICAO24(c.Context(), icao24)
	if err != nil {
		if errors.Is(err, postgres.ErrFlightStateNotFound) {
			return response.Error(c, fiber.StatusNotFound, "FLIGHT_STATE_NOT_FOUND", "Flight state not found")
		}

		return response.Error(c, fiber.StatusInternalServerError, "FLIGHT_STATE_LOAD_FAILED", "Failed to load flight state")
	}

	return response.OK(c, toFlightStateItem(item))
}

func toFlightStateItems(items []flightstate.FlightState) []dto.FlightStateItem {
	result := make([]dto.FlightStateItem, 0, len(items))

	for _, item := range items {
		result = append(result, toFlightStateItem(item))
	}

	return result
}

func toFlightStateItem(item flightstate.FlightState) dto.FlightStateItem {
	return dto.FlightStateItem{
		ID:                  item.ID,
		FlightID:            item.FlightID,
		AircraftID:          item.AircraftID,
		ICAO24:              item.ICAO24,
		Callsign:            item.Callsign,
		Latitude:            item.Latitude,
		Longitude:           item.Longitude,
		BarometricAltitudeM: item.BarometricAltitudeM,
		GeometricAltitudeM:  item.GeometricAltitudeM,
		VelocityMPS:         item.VelocityMPS,
		HeadingDegrees:      item.HeadingDegrees,
		VerticalRateMPS:     item.VerticalRateMPS,
		OnGround:            item.OnGround,
		OriginCountry:       item.OriginCountry,
		ObservedAt:          item.ObservedAt,
		SourceName:          item.SourceName,
	}
}
