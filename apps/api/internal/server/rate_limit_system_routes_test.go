package server

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestNewDoesNotRateLimitReadinessRoute(
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
		defer response.Body.Close()

		if response.StatusCode !=
			fiber.StatusServiceUnavailable {
			t.Fatalf(
				"expected readiness status 503 without PostgreSQL, got %d",
				response.StatusCode,
			)
		}
	}
}
