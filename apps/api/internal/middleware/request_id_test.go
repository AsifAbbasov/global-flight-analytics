package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestIDPreservesSafeClientValue(
	t *testing.T,
) {
	app := fiber.New()
	app.Use(
		RequestID(),
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

	request := httptest.NewRequest(
		fiber.MethodGet,
		"/",
		nil,
	)
	request.Header.Set(
		RequestIDHeader,
		"trace-123_ABC",
	)

	response, err := app.Test(
		request,
	)
	if err != nil {
		t.Fatalf(
			"execute request: %v",
			err,
		)
	}

	if actual := response.Header.Get(
		RequestIDHeader,
	); actual != "trace-123_ABC" {
		t.Fatalf(
			"expected preserved request id, got %q",
			actual,
		)
	}
}

func TestRequestIDReplacesUnsafeClientValue(
	t *testing.T,
) {
	app := fiber.New()
	app.Use(
		RequestID(),
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

	unsafeValue := strings.Repeat(
		"a",
		maximumRequestIDLength+1,
	)

	request := httptest.NewRequest(
		fiber.MethodGet,
		"/",
		nil,
	)
	request.Header.Set(
		RequestIDHeader,
		unsafeValue,
	)

	response, err := app.Test(
		request,
	)
	if err != nil {
		t.Fatalf(
			"execute request: %v",
			err,
		)
	}

	actual := response.Header.Get(
		RequestIDHeader,
	)
	if actual == "" || actual == unsafeValue {
		t.Fatalf(
			"expected generated request id, got %q",
			actual,
		)
	}
}

func TestNormalizeRequestIDRejectsUnsafeCharacters(
	t *testing.T,
) {
	for _, value := range []string{
		"request id with spaces",
		"request/id",
		"request?id",
		"request,id",
	} {
		if actual := normalizeRequestID(
			value,
		); actual != "" {
			t.Fatalf(
				"expected %q to be rejected, got %q",
				value,
				actual,
			)
		}
	}
}
