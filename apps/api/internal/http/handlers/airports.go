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
