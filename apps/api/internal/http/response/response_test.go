package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/gofiber/fiber/v2"
)

func TestOKSerializesTypedStructPayload(
	t *testing.T,
) {
	app := fiber.New()

	app.Get(
		"/health",
		func(
			c *fiber.Ctx,
		) error {
			return OK(
				c,
				dto.HealthResponse{
					Status: "ok",
				},
			)
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	result, err := app.Test(
		request,
		-1,
	)
	if err != nil {
		t.Fatalf(
			"execute typed success response request: %v",
			err,
		)
	}
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			result.StatusCode,
		)
	}

	var envelope struct {
		Success bool               `json:"success"`
		Data    dto.HealthResponse `json:"data"`
	}

	if err := json.NewDecoder(
		result.Body,
	).Decode(
		&envelope,
	); err != nil {
		t.Fatalf(
			"decode typed success response: %v",
			err,
		)
	}

	if !envelope.Success {
		t.Fatal(
			"expected success response",
		)
	}

	if envelope.Data.Status != "ok" {
		t.Fatalf(
			"expected health status ok, got %q",
			envelope.Data.Status,
		)
	}
}

func TestOKSerializesTypedSlicePayload(
	t *testing.T,
) {
	app := fiber.New()

	app.Get(
		"/regions",
		func(
			c *fiber.Ctx,
		) error {
			return OK(
				c,
				[]dto.RegionItem{
					{
						Code: "AZ",
						Name: "Azerbaijan",
					},
				},
			)
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/regions",
		nil,
	)

	result, err := app.Test(
		request,
		-1,
	)
	if err != nil {
		t.Fatalf(
			"execute typed slice response request: %v",
			err,
		)
	}
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			result.StatusCode,
		)
	}

	var envelope struct {
		Success bool             `json:"success"`
		Data    []dto.RegionItem `json:"data"`
	}

	if err := json.NewDecoder(
		result.Body,
	).Decode(
		&envelope,
	); err != nil {
		t.Fatalf(
			"decode typed slice response: %v",
			err,
		)
	}

	if !envelope.Success {
		t.Fatal(
			"expected success response",
		)
	}

	if len(envelope.Data) != 1 {
		t.Fatalf(
			"expected one region item, got %d",
			len(envelope.Data),
		)
	}

	if envelope.Data[0].Code != "AZ" {
		t.Fatalf(
			"expected region code AZ, got %q",
			envelope.Data[0].Code,
		)
	}
}
