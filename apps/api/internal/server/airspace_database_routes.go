package server

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func registerAirspaceDatabaseRoutes(
	v1 fiber.Router,
	runtime airspaceDatabaseRuntime,
) error {
	if err := RegisterAirspaceRegionAnalyticsReadRoute(
		v1,
		runtime.regionAnalytics,
	); err != nil {
		return fmt.Errorf(
			"register Airspace Region Analytics route: %w",
			err,
		)
	}

	return nil
}
