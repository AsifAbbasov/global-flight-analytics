package handlers

import (
	"errors"

	aviationconstraints "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/constraints"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/flightstate"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

type FlightStateHandler struct {
	service *flightstate.Service
}

func NewFlightStateHandler(
	service *flightstate.Service,
) *FlightStateHandler {
	return &FlightStateHandler{
		service: service,
	}
}

func (h *FlightStateHandler) ListByFlightID(
	c *fiber.Ctx,
) error {
	flightID := c.Params(
		"flightID",
	)

	items, err := h.service.ListByFlightID(
		c.Context(),
		flightID,
	)
	if err != nil {
		return response.Error(
			c,
			fiber.StatusInternalServerError,
			"FLIGHT_STATES_LIST_FAILED",
			"Failed to load flight states",
		)
	}

	return response.OK(
		c,
		toFlightStateItems(
			items,
		),
	)
}

func (h *FlightStateHandler) GetLatestByICAO24(
	c *fiber.Ctx,
) error {
	icao24 := c.Params(
		"icao24",
	)

	item, err := h.service.GetLatestByICAO24(
		c.Context(),
		icao24,
	)
	if err != nil {
		if errors.Is(
			err,
			flightstate.ErrNotFound,
		) {
			return response.Error(
				c,
				fiber.StatusNotFound,
				"FLIGHT_STATE_NOT_FOUND",
				"Flight state not found",
			)
		}

		return response.Error(
			c,
			fiber.StatusInternalServerError,
			"FLIGHT_STATE_LOAD_FAILED",
			"Failed to load flight state",
		)
	}

	return response.OK(
		c,
		toFlightStateItem(
			item,
		),
	)
}

func toFlightStateItems(
	items []flightstate.FlightState,
) []dto.FlightStateItem {
	result := make(
		[]dto.FlightStateItem,
		0,
		len(items),
	)

	for _, item := range items {
		result = append(
			result,
			toFlightStateItem(
				item,
			),
		)
	}

	return result
}

func toFlightStateItem(
	item flightstate.FlightState,
) dto.FlightStateItem {
	barometricValue, barometricStatus :=
		toPublicAltitude(
			item.BarometricAltitudeM,
			item.BarometricAltitudeStatus,
			item.OnGround,
		)

	geometricValue, geometricStatus :=
		toPublicAltitude(
			item.GeometricAltitudeM,
			item.GeometricAltitudeStatus,
			item.OnGround,
		)

	return dto.FlightStateItem{
		ID:                       item.ID,
		FlightID:                 item.FlightID,
		AircraftID:               item.AircraftID,
		ICAO24:                   item.ICAO24,
		Callsign:                 item.Callsign,
		Latitude:                 item.Latitude,
		Longitude:                item.Longitude,
		BarometricAltitudeM:      barometricValue,
		BarometricAltitudeStatus: barometricStatus,
		GeometricAltitudeM:       geometricValue,
		GeometricAltitudeStatus:  geometricStatus,
		VelocityMPS:              item.VelocityMPS,
		HeadingDegrees:           item.HeadingDegrees,
		VerticalRateMPS:          item.VerticalRateMPS,
		OnGround:                 item.OnGround,
		OriginCountry:            item.OriginCountry,
		ObservedAt:               item.ObservedAt,
		SourceName:               item.SourceName,
	}
}

func toPublicAltitude(
	value float64,
	status flightstate.AltitudeStatus,
	onGround bool,
) (
	*float64,
	flightstate.AltitudeStatus,
) {
	effectiveStatus := flightstate.ResolveAltitudeStatus(
		value,
		status,
	)

	if !flightstate.IsKnownAltitudeStatus(
		effectiveStatus,
	) {
		return nil,
			flightstate.AltitudeStatusInvalid
	}

	switch effectiveStatus {
	case flightstate.AltitudeStatusObserved:
		if !aviationconstraints.IsNonNegativeFloat64(
			value,
		) {
			return nil,
				flightstate.AltitudeStatusInvalid
		}

		result := value

		return &result,
			effectiveStatus

	case flightstate.AltitudeStatusGround:
		if value != 0 ||
			!onGround {
			return nil,
				flightstate.AltitudeStatusInvalid
		}

		result := 0.0

		return &result,
			effectiveStatus

	case flightstate.AltitudeStatusUnknown,
		flightstate.AltitudeStatusUnavailable,
		flightstate.AltitudeStatusInvalid:
		return nil,
			effectiveStatus

	default:
		return nil,
			flightstate.AltitudeStatusInvalid
	}
}
