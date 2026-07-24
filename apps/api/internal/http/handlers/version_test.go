package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/buildinfo"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/dto"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/gofiber/fiber/v2"
)

func TestVersionReturnsBuildMetadata(
	t *testing.T,
) {
	app := fiber.New()
	app.Get(
		"/version",
		func(
			c *fiber.Ctx,
		) error {
			return sendVersion(
				c,
				buildinfo.Info{
					Version:  "1.2.3",
					Revision: "abcdef",
					BuiltAt:  "2026-07-24T00:00:00Z",
				},
			)
		},
	)

	httpResponse, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/version",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute version request: %v",
			err,
		)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			httpResponse.StatusCode,
		)
	}

	var payload response.SuccessResponse[dto.VersionResponse]
	if err := json.NewDecoder(
		httpResponse.Body,
	).Decode(
		&payload,
	); err != nil {
		t.Fatalf(
			"decode version response: %v",
			err,
		)
	}

	if !payload.Success {
		t.Fatal(
			"expected successful version response",
		)
	}
	if payload.Data.Version != "1.2.3" ||
		payload.Data.Revision != "abcdef" ||
		payload.Data.BuiltAt !=
			"2026-07-24T00:00:00Z" {
		t.Fatalf(
			"unexpected version metadata: %+v",
			payload.Data,
		)
	}
}
