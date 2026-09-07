package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ClientIPResolver func(
	c *fiber.Ctx,
) string

func RequestLogger(
	log *slog.Logger,
	_ ...ClientIPResolver,
) fiber.Handler {
	if log == nil {
		log = slog.Default()
	}

	return func(
		c *fiber.Ctx,
	) error {
		start := time.Now()

		err := c.Next()
		if err != nil {
			err = c.App().ErrorHandler(
				c,
				err,
			)
		}

		duration := time.Since(
			start,
		)
		requestID, _ := c.Locals(
			RequestIDLocalKey,
		).(string)

		route := "unmatched"
		if currentRoute := c.Route(); currentRoute != nil &&
			currentRoute.Path != "" {
			route = currentRoute.Path
		}

		log.Info(
			"http request completed",
			"request_id",
			requestID,
			"method",
			c.Method(),
			"route",
			route,
			"status",
			c.Response().StatusCode(),
			"duration_ms",
			duration.Milliseconds(),
		)

		return err
	}
}
