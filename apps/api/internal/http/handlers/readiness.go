package handlers

import (
	"context"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

const readinessProbeTimeout = 2 * time.Second

type ReadinessProbe func(context.Context) error

func Readiness(
	probe ReadinessProbe,
) fiber.Handler {
	return func(
		ctx *fiber.Ctx,
	) error {
		if probe == nil {
			return response.Error(
				ctx,
				fiber.StatusServiceUnavailable,
				"SERVICE_NOT_READY",
				"PostgreSQL is not configured",
			)
		}

		probeContext, cancel := context.WithTimeout(
			ctx.UserContext(),
			readinessProbeTimeout,
		)
		defer cancel()

		if err := probe(probeContext); err != nil {
			return response.Error(
				ctx,
				fiber.StatusServiceUnavailable,
				"SERVICE_NOT_READY",
				"PostgreSQL is unavailable",
			)
		}

		return response.OK(
			ctx,
			dto.HealthResponse{
				Status: "ready",
			},
		)
	}
}
