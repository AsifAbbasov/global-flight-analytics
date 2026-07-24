package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestLogger(
	log *slog.Logger,
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

		log.Info(
			"http request completed",
			"request_id",
			requestID,
			"method",
			c.Method(),
			"path",
			c.Path(),
			"status",
			c.Response().StatusCode(),
			"duration_ms",
			duration.Milliseconds(),
			"ip",
			c.IP(),
		)

		return err
	}
}
