package server

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestErrorHandlerDoesNotLogRawInternalError(
	t *testing.T,
) {
	const secret = "password=secret-value"

	var output bytes.Buffer
	log := slog.New(
		slog.NewJSONHandler(
			&output,
			nil,
		),
	)

	app := fiber.New(
		fiber.Config{
			ErrorHandler: newErrorHandler(
				log,
			),
		},
	)
	app.Get(
		"/failure",
		func(
			*fiber.Ctx,
		) error {
			return errors.New(
				"database " + secret,
			)
		},
	)

	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/failure",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute failure request: %v",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode !=
		fiber.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d",
			response.StatusCode,
		)
	}

	logged := output.String()
	if strings.Contains(
		logged,
		secret,
	) {
		t.Fatalf(
			"sensitive error text leaked into log: %s",
			logged,
		)
	}
	if !strings.Contains(
		logged,
		`"error_type"`,
	) {
		t.Fatalf(
			"expected safe error type metadata, got %s",
			logged,
		)
	}
	if !strings.Contains(
		logged,
		`"status":500`,
	) {
		t.Fatalf(
			"expected final status metadata, got %s",
			logged,
		)
	}
}
