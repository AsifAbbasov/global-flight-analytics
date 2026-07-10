package handlers

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

func Health(
	c *fiber.Ctx,
) error {
	return response.OK(
		c,
		dto.HealthResponse{
			Status: "ok",
		},
	)
}
