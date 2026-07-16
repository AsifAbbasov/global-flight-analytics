package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
)

const WeatherContextPath = "/trajectories/:id/weather-context"

// RegisterWeatherContextReadRoute composes the read-only Weather Context
// endpoint with an already constructed query service. Production composition
// and runtime verification may provide different readers while preserving the
// same HTTP contract.
func RegisterWeatherContextReadRoute(
	v1 fiber.Router,
	reader handlers.WeatherContextReader,
) error {
	if reader == nil {
		return fmt.Errorf(
			"Weather Context reader is required",
		)
	}

	handler := handlers.NewWeatherContextHandler(
		reader,
	)
	v1.Get(
		WeatherContextPath,
		handler.GetByTrajectoryID,
	)

	return nil
}
