package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSystemRoutesSeparateLivenessFromReadiness(
	t *testing.T,
) {
	app := fiber.New()
	v1 := app.Group(
		"/api",
	).Group(
		"/v1",
	)
	registerSystemRoutes(
		v1,
		nil,
	)

	healthResponse, err := app.Test(
		httptest.NewRequest(
			"GET",
			"/api/v1/health",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute health request: %v",
			err,
		)
	}
	defer healthResponse.Body.Close()

	if healthResponse.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"expected health status 200, got %d",
			healthResponse.StatusCode,
		)
	}

	readyResponse, err := app.Test(
		httptest.NewRequest(
			"GET",
			"/api/v1/ready",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute readiness request: %v",
			err,
		)
	}
	defer readyResponse.Body.Close()

	if readyResponse.StatusCode !=
		fiber.StatusServiceUnavailable {
		t.Fatalf(
			"expected readiness status 503, got %d",
			readyResponse.StatusCode,
		)
	}

	var payload struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(
		readyResponse.Body,
	).Decode(
		&payload,
	); err != nil {
		t.Fatalf(
			"decode readiness response: %v",
			err,
		)
	}
	if payload.Success ||
		payload.Error.Code !=
			"SERVICE_NOT_READY" {
		t.Fatalf(
			"unexpected readiness response: %+v",
			payload,
		)
	}
}
