package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestReadinessReportsReadyWhenPostgreSQLResponds(
	t *testing.T,
) {
	app := fiber.New()
	app.Get(
		"/ready",
		Readiness(
			func(
				context.Context,
			) error {
				return nil
			},
		),
	)

	response, err := app.Test(
		httptest.NewRequest(
			"GET",
			"/ready",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute readiness request: %v",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			response.StatusCode,
		)
	}

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(
		response.Body,
	).Decode(
		&payload,
	); err != nil {
		t.Fatalf(
			"decode readiness response: %v",
			err,
		)
	}
	if !payload.Success ||
		payload.Data.Status != "ready" {
		t.Fatalf(
			"unexpected readiness payload: %+v",
			payload,
		)
	}
}

func TestReadinessFailsClosedWithoutPostgreSQL(
	t *testing.T,
) {
	tests := []struct {
		name  string
		probe ReadinessProbe
	}{
		{
			name:  "database is not configured",
			probe: nil,
		},
		{
			name: "database ping fails",
			probe: func(
				context.Context,
			) error {
				return errors.New(
					"postgres unavailable",
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(
				t *testing.T,
			) {
				app := fiber.New()
				app.Get(
					"/ready",
					Readiness(
						test.probe,
					),
				)

				response, err := app.Test(
					httptest.NewRequest(
						"GET",
						"/ready",
						nil,
					),
				)
				if err != nil {
					t.Fatalf(
						"execute readiness request: %v",
						err,
					)
				}
				defer response.Body.Close()

				if response.StatusCode !=
					fiber.StatusServiceUnavailable {
					t.Fatalf(
						"expected status 503, got %d",
						response.StatusCode,
					)
				}

				var payload struct {
					Success bool `json:"success"`
					Error   struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.NewDecoder(
					response.Body,
				).Decode(
					&payload,
				); err != nil {
					t.Fatalf(
						"decode readiness error: %v",
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
			},
		)
	}
}
