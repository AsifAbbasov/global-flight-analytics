package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func registerProjectionDatabaseRoutes(
	v1 fiber.Router,
	runtime projectionDatabaseRuntime,
) error {
	if err := RegisterProjectionIntelligenceReadRoute(
		v1,
		runtime.projection,
	); err != nil {
		return fmt.Errorf(
			"register Projection Intelligence route: %w",
			err,
		)
	}

	if err := RegisterStabilityIntelligenceReadRoute(
		v1,
		runtime.stability,
	); err != nil {
		return fmt.Errorf(
			"register Stability Intelligence route: %w",
			err,
		)
	}

	if err := RegisterWeatherContextReadRoute(
		v1,
		runtime.weather,
	); err != nil {
		return fmt.Errorf(
			"register Weather Context route: %w",
			err,
		)
	}

	return nil
}
