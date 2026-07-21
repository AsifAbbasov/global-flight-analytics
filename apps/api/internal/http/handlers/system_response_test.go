package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/gofiber/fiber/v2"
)

func TestHealthPublishesTypedSuccessEnvelope(
	t *testing.T,
) {
	app := fiber.New()

	app.Get(
		"/health",
		Health,
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
			"execute health request: %v",
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
			"decode health response: %v",
			err,
		)
	}

	if !envelope.Success {
		t.Fatal(
			"expected health success response",
		)
	}

	if envelope.Data.Status != "ok" {
		t.Fatalf(
			"expected health status ok, got %q",
			envelope.Data.Status,
		)
	}
}

func TestVersionPublishesTypedSuccessEnvelope(
	t *testing.T,
) {
	app := fiber.New()

	app.Get(
		"/version",
		Version,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/version",
		nil,
	)

	result, err := app.Test(
		request,
		-1,
	)
	if err != nil {
		t.Fatalf(
			"execute version request: %v",
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
		Success bool                `json:"success"`
		Data    dto.VersionResponse `json:"data"`
	}

	if err := json.NewDecoder(
		result.Body,
	).Decode(
		&envelope,
	); err != nil {
		t.Fatalf(
			"decode version response: %v",
			err,
		)
	}

	if !envelope.Success {
		t.Fatal(
			"expected version success response",
		)
	}

	if envelope.Data.Version == "" {
		t.Fatal("expected non-empty version")
	}
}
