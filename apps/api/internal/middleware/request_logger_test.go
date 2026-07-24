package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestLoggerRecordsFinalErrorHandlerStatus(
	t *testing.T,
) {
	var output bytes.Buffer
	log := slog.New(
		slog.NewJSONHandler(
			&output,
			nil,
		),
	)

	app := fiber.New(
		fiber.Config{
			ErrorHandler: func(
				c *fiber.Ctx,
				_ error,
			) error {
				return c.Status(
					fiber.StatusTeapot,
				).SendString(
					"handled",
				)
			},
		},
	)
	app.Use(
		RequestLogger(
			log,
		),
	)
	app.Get(
		"/failure",
		func(
			*fiber.Ctx,
		) error {
			return errors.New(
				"route failed",
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
			"execute request: %v",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusTeapot {
		t.Fatalf(
			"expected final status 418, got %d",
			response.StatusCode,
		)
	}

	record := decodeRequestLog(
		t,
		output.Bytes(),
	)
	if status, ok := record["status"].(float64); !ok ||
		int(status) != fiber.StatusTeapot {
		t.Fatalf(
			"expected logged status 418, got %#v",
			record["status"],
		)
	}
}

func TestRequestLoggerPreservesSuccessfulResponseStatus(
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
		),
	)
	app.Get(
		"/success",
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
			"/success",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute request: %v",
			err,
		)
	}
	defer response.Body.Close()

	record := decodeRequestLog(
		t,
		output.Bytes(),
	)
	if status, ok := record["status"].(float64); !ok ||
		int(status) != fiber.StatusNoContent {
		t.Fatalf(
			"expected logged status 204, got %#v",
			record["status"],
		)
	}
}

func decodeRequestLog(
	t *testing.T,
	data []byte,
) map[string]any {
	t.Helper()

	var record map[string]any
	if err := json.Unmarshal(
		bytes.TrimSpace(
			data,
		),
		&record,
	); err != nil {
		t.Fatalf(
			"decode request log: %v; log=%s",
			err,
			data,
		)
	}
	return record
}
