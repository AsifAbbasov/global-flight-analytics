package handlers

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

func ListAirports(c *fiber.Ctx) error {
	return response.OK(c, []dto.AirportListItem{
		{
			ICAOCode:  "UBBB",
			IATACode:  "GYD",
			Name:      "Heydar Aliyev International Airport",
			City:      "Baku",
			Country:   "Azerbaijan",
			Latitude:  40.4675,
			Longitude: 50.0467,
		},
	})
}

func GetAirport(c *fiber.Ctx) error {
	icao := c.Params("icao")

	if icao != "UBBB" {
		return response.Error(
			c,
			fiber.StatusNotFound,
			"AIRPORT_NOT_FOUND",
			"Airport not found",
		)
	}

	return response.OK(c, dto.AirportProfile{
		ICAOCode:    "UBBB",
		IATACode:    "GYD",
		Name:        "Heydar Aliyev International Airport",
		City:        "Baku",
		Country:     "Azerbaijan",
		Latitude:    40.4675,
		Longitude:   50.0467,
		ElevationFt: 10,
		Timezone:    "Asia/Baku",
		Description: "Main international airport serving Baku and Azerbaijan.",
	})
}
