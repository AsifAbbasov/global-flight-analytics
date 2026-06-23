package handlers

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

func ListAirports(c *fiber.Ctx) error {
	return response.OK(c, []fiber.Map{
		{
			"icao_code": "UBBB",
			"iata_code": "GYD",
			"name":      "Heydar Aliyev International Airport",
			"city":      "Baku",
			"country":   "Azerbaijan",
			"latitude":  40.4675,
			"longitude": 50.0467,
		},
	})
}
