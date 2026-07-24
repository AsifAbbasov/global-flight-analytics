package handlers

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/buildinfo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

func Version(
	c *fiber.Ctx,
) error {
	return sendVersion(
		c,
		buildinfo.Current(),
	)
}

func sendVersion(
	c *fiber.Ctx,
	info buildinfo.Info,
) error {
	return response.OK(
		c,
		dto.VersionResponse{
			Version:  info.Version,
			Revision: info.Revision,
			BuiltAt:  info.BuiltAt,
		},
	)
}
