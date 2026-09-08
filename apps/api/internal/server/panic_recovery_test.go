package server

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestPanicRecoveryDoesNotLeakPanicDetails(
	t *testing.T,
) {
	app, err := New(
		Config{},
	)
	if err != nil {
		t.Fatalf(
			"create server: %v",
			err,
		)
	}

	const sensitivePanic = "sensitive panic implementation detail"
	app.Get(
		"/panic-hardening-test",
		func(
			*fiber.Ctx,
		) error {
			panic(sensitivePanic)
		},
	)

	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/panic-hardening-test",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute panic request: %v",
			err,
		)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(
		response.Body,
	)
	if err != nil {
		t.Fatalf(
			"read panic response: %v",
			err,
		)
	}

	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d",
			response.StatusCode,
		)
	}
	if strings.Contains(
		string(body),
		sensitivePanic,
	) {
		t.Fatalf(
			"panic response leaked panic details: %s",
			body,
		)
	}
	if !strings.Contains(
		string(body),
		"INTERNAL_SERVER_ERROR",
	) {
		t.Fatalf(
			"expected stable internal error code, got %s",
			body,
		)
	}
}
