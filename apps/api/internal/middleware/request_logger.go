package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ClientIPResolver func(
	c *fiber.Ctx,
) string

func RequestLogger(
	log *slog.Logger,
	clientIPResolvers ...ClientIPResolver,
) fiber.Handler {
	if log == nil {
		log = slog.Default()
	}

	resolveClientIP := ClientIPResolver(
		func(
			c *fiber.Ctx,
		) string {
			return c.IP()
		},
	)
	if len(clientIPResolvers) > 0 &&
		clientIPResolvers[0] != nil {
		resolveClientIP = clientIPResolvers[0]
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

		clientIP := strings.TrimSpace(
			resolveClientIP(
				c,
			),
		)
		if clientIP == "" {
			clientIP = c.IP()
		}

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
			clientIP,
		)

		return err
	}
}
