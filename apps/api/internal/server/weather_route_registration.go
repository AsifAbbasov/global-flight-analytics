package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
)

const CurrentWeatherPath = "/weather/current"

func registerCurrentWeatherRoute(
	v1 fiber.Router,
	handler *handlers.WeatherHandler,
) error {
	if v1 == nil {
		return fmt.Errorf(
			"weather route router is required",
		)
	}
	if handler == nil {
		return fmt.Errorf(
			"weather route handler is required",
		)
	}

	v1.Get(
		CurrentWeatherPath,
		handler.GetCurrent,
	)

	return nil
}
