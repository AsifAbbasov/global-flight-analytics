package handlers

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

func Version(
	c *fiber.Ctx,
) error {
	return response.OK(
		c,
		dto.VersionResponse{
			Version: "1.0.0",
		},
	)
}
