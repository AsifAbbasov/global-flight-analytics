package server

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/gofiber/fiber/v2"
)

const AirspaceRegionAnalyticsPath = "/airspace/regions/:code/analytics"

func RegisterAirspaceRegionAnalyticsReadRoute(
	v1 fiber.Router,
	reader handlers.AirspaceRegionAnalyticsReader,
) error {
	if reader == nil {
		return fmt.Errorf(
			"Airspace Region Analytics reader is required",
		)
	}
	handler := handlers.NewAirspaceRegionAnalyticsHandler(reader)
	v1.Get(
		AirspaceRegionAnalyticsPath,
		handler.GetByRegionCode,
	)
	return nil
}
