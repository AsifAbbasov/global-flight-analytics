package server

import (
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestNewAppliesSecurityHeadersAndRequestID(
	t *testing.T,
) {
	app := newProtectedTestApp(
		t,
		ProtectionConfig{},
	)

	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
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

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"expected health status 200, got %d",
			response.StatusCode,
		)
	}

	if response.Header.Get(
		"X-Request-ID",
	) == "" {
		t.Fatal(
			"expected response request id",
		)
	}

	if actual := response.Header.Get(
		fiber.HeaderXContentTypeOptions,
	); actual != "nosniff" {
		t.Fatalf(
			"expected nosniff header, got %q",
			actual,
		)
	}
}

func TestNewUsesControlledCORSOrigins(
	t *testing.T,
) {
	app := newProtectedTestApp(
		t,
		ProtectionConfig{
			AllowedOrigins: "https://web.example.com",
		},
	)

	allowedRequest := httptest.NewRequest(
		fiber.MethodGet,
		"/api/v1/health",
		nil,
	)
	allowedRequest.Header.Set(
		fiber.HeaderOrigin,
		"https://web.example.com",
	)

	allowedResponse, err := app.Test(
		allowedRequest,
	)
	if err != nil {
		t.Fatalf(
			"execute allowed origin request: %v",
			err,
		)
	}

	if actual := allowedResponse.Header.Get(
		fiber.HeaderAccessControlAllowOrigin,
	); actual != "https://web.example.com" {
		t.Fatalf(
			"expected allowed origin header, got %q",
			actual,
		)
	}

	disallowedRequest := httptest.NewRequest(
		fiber.MethodGet,
		"/api/v1/health",
		nil,
	)
	disallowedRequest.Header.Set(
		fiber.HeaderOrigin,
		"https://attacker.example.com",
	)

	disallowedResponse, err := app.Test(
		disallowedRequest,
	)
	if err != nil {
		t.Fatalf(
			"execute disallowed origin request: %v",
			err,
		)
	}

	if actual := disallowedResponse.Header.Get(
		fiber.HeaderAccessControlAllowOrigin,
	); actual != "" {
		t.Fatalf(
			"expected no CORS grant for disallowed origin, got %q",
			actual,
		)
	}
}

func TestNewRateLimitsNonSystemRoutes(
	t *testing.T,
) {
	app := newProtectedTestApp(
		t,
		ProtectionConfig{
			RateLimitMax:    2,
			RateLimitWindow: time.Minute,
		},
	)

	app.Get(
		"/api/v1/protected-test",
		func(
			c *fiber.Ctx,
		) error {
			return c.SendStatus(
				fiber.StatusNoContent,
			)
		},
	)

	for attempt := 1; attempt <= 3; attempt++ {
		response, err := app.Test(
			httptest.NewRequest(
				fiber.MethodGet,
				"/api/v1/protected-test",
				nil,
			),
		)
		if err != nil {
			t.Fatalf(
				"execute request %d: %v",
				attempt,
				err,
			)
		}

		expectedStatus := fiber.StatusNoContent
		if attempt == 3 {
			expectedStatus = fiber.StatusTooManyRequests
		}

		if response.StatusCode != expectedStatus {
			body, _ := io.ReadAll(
				response.Body,
			)
			t.Fatalf(
				"attempt %d: expected status %d, got %d body=%s",
				attempt,
				expectedStatus,
				response.StatusCode,
				body,
			)
		}
	}
}

func TestNewDoesNotRateLimitHealthRoute(
	t *testing.T,
) {
	app := newProtectedTestApp(
		t,
		ProtectionConfig{
			RateLimitMax:    1,
			RateLimitWindow: time.Minute,
		},
	)

	for attempt := 0; attempt < 3; attempt++ {
		response, err := app.Test(
			httptest.NewRequest(
				fiber.MethodGet,
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

		if response.StatusCode != fiber.StatusOK {
			t.Fatalf(
				"expected health route to remain available, got %d",
				response.StatusCode,
			)
		}
	}
}

func TestErrorHandlerDoesNotExposeInternalFailure(
	t *testing.T,
) {
	app := newProtectedTestApp(
		t,
		ProtectionConfig{},
	)

	app.Get(
		"/api/v1/failure-test",
		func(
			*fiber.Ctx,
		) error {
			return errors.New(
				"database password=secret-value",
			)
		},
	)

	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/api/v1/failure-test",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute failure request: %v",
			err,
		)
	}

	body, err := io.ReadAll(
		response.Body,
	)
	if err != nil {
		t.Fatalf(
			"read failure response: %v",
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
		"secret-value",
	) {
		t.Fatalf(
			"internal failure leaked in response: %s",
			body,
		)
	}

	if !strings.Contains(
		string(body),
		"INTERNAL_SERVER_ERROR",
	) {
		t.Fatalf(
			"expected safe error code, got %s",
			body,
		)
	}
}

func TestNewRejectsInvalidProtectionConfiguration(
	t *testing.T,
) {
	app, err := New(
		Config{
			Logger: newDiscardLogger(),
			Protection: ProtectionConfig{
				BodyLimitBytes: -1,
			},
		},
	)

	if err == nil {
		t.Fatal(
			"expected protection configuration error",
		)
	}

	if app != nil {
		t.Fatal(
			"expected nil app for invalid protection configuration",
		)
	}
}

func newProtectedTestApp(
	t *testing.T,
	protection ProtectionConfig,
) *fiber.App {
	t.Helper()

	app, err := New(
		Config{
			Logger:     newDiscardLogger(),
			Protection: protection,
		},
	)
	if err != nil {
		t.Fatalf(
			"create protected server: %v",
			err,
		)
	}

	return app
}
