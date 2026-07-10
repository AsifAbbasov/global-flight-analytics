package response

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/gofiber/fiber/v2"
)

type SuccessPayload interface {
	[]dto.FlightListItem |
		dto.FlightProfile |
		dto.HealthResponse |
		dto.VersionResponse |
		dto.CurrentWeatherResponse |
		[]dto.RegionItem |
		dto.RegionItem |
		[]dto.AircraftListItem |
		dto.AircraftProfile |
		[]dto.FlightStateItem |
		dto.FlightStateItem |
		[]dto.CurrentTrafficItem |
		dto.Trajectory |
		[]dto.AirportListItem |
		dto.AirportProfile
}

type SuccessResponse[T SuccessPayload] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK[T SuccessPayload](
	c *fiber.Ctx,
	data T,
) error {
	return c.JSON(
		SuccessResponse[T]{
			Success: true,
			Data:    data,
		},
	)
}

func Error(
	c *fiber.Ctx,
	status int,
	code string,
	message string,
) error {
	return c.Status(
		status,
	).JSON(
		ErrorResponse{
			Success: false,
			Error: ErrorBody{
				Code:    code,
				Message: message,
			},
		},
	)
}
