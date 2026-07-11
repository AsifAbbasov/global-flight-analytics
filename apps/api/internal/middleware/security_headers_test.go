package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSecurityHeadersSetsAPIResponseProtection(
	t *testing.T,
) {
	app := fiber.New()
	app.Use(
		SecurityHeaders(),
	)
	app.Get(
		"/",
		func(
			c *fiber.Ctx,
		) error {
			return c.SendStatus(
				fiber.StatusNoContent,
			)
		},
	)

	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute request: %v",
			err,
		)
	}

	expectedHeaders := map[string]string{
		fiber.HeaderXContentTypeOptions:     "nosniff",
		fiber.HeaderXFrameOptions:           "DENY",
		"Referrer-Policy":                   "no-referrer",
		"Permissions-Policy":                "camera=(), geolocation=(), microphone=()",
		"Content-Security-Policy":           "default-src 'none'; base-uri 'none'; frame-ancestors 'none'",
		"X-Permitted-Cross-Domain-Policies": "none",
	}

	for name, expected := range expectedHeaders {
		if actual := response.Header.Get(
			name,
		); actual != expected {
			t.Fatalf(
				"expected %s=%q, got %q",
				name,
				expected,
				actual,
			)
		}
	}
}
