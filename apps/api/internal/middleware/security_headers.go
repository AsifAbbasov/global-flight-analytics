package middleware

import "github.com/gofiber/fiber/v2"

const strictTransportSecurityPolicy = "max-age=31536000"

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(
			fiber.HeaderXContentTypeOptions,
			"nosniff",
		)
		c.Set(
			fiber.HeaderXFrameOptions,
			"DENY",
		)
		c.Set(
			"Referrer-Policy",
			"no-referrer",
		)
		c.Set(
			"Permissions-Policy",
			"camera=(), geolocation=(), microphone=()",
		)
		c.Set(
			"Content-Security-Policy",
			"default-src 'none'; base-uri 'none'; frame-ancestors 'none'",
		)
		c.Set(
			"X-Permitted-Cross-Domain-Policies",
			"none",
		)
		c.Set(
			"Strict-Transport-Security",
			strictTransportSecurityPolicy,
		)

		return c.Next()
	}
}
