package middleware

import (
	"bytes"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestLoggerUsesConfiguredClientIdentity(
	t *testing.T,
) {
	var output bytes.Buffer
	log := slog.New(
		slog.NewJSONHandler(
			&output,
			nil,
		),
	)

	app := fiber.New()
	app.Use(
		RequestLogger(
			log,
			func(
				*fiber.Ctx,
			) string {
				return "203.0.113.50"
			},
		),
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

	httpResponse, err := app.Test(
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
	defer httpResponse.Body.Close()

	if !strings.Contains(
		output.String(),
		`"ip":"203.0.113.50"`,
	) {
		t.Fatalf(
			"expected resolved client identity in log, got %s",
			output.String(),
		)
	}
}
