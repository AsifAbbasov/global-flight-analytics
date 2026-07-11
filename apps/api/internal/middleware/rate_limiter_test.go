package middleware

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestRateLimiterRejectsRequestsBeyondWindowLimit(
	t *testing.T,
) {
	now := time.Date(
		2026,
		time.July,
		11,
		21,
		0,
		0,
		0,
		time.UTC,
	)

	handler, err := NewRateLimiter(
		RateLimiterConfig{
			MaxRequests: 2,
			Window:      time.Minute,
			KeyGenerator: func(
				*fiber.Ctx,
			) string {
				return "client-a"
			},
			Now: func() time.Time {
				return now
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create rate limiter: %v",
			err,
		)
	}

	app := fiber.New()
	app.Use(
		handler,
	)
	app.Get(
		"/",
		func(
			c *fiber.Ctx,
		) error {
			return c.SendStatus(
				fiber.StatusOK,
			)
		},
	)

	for attempt := 1; attempt <= 3; attempt++ {
		response, requestErr := app.Test(
			httptest.NewRequest(
				fiber.MethodGet,
				"/",
				nil,
			),
		)
		if requestErr != nil {
			t.Fatalf(
				"execute request %d: %v",
				attempt,
				requestErr,
			)
		}

		expectedStatus := fiber.StatusOK
		if attempt == 3 {
			expectedStatus = fiber.StatusTooManyRequests
		}

		if response.StatusCode != expectedStatus {
			t.Fatalf(
				"attempt %d: expected status %d, got %d",
				attempt,
				expectedStatus,
				response.StatusCode,
			)
		}
	}

	now = now.Add(
		time.Minute,
	)

	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute request after reset: %v",
			err,
		)
	}

	if response.StatusCode != fiber.StatusOK {
		t.Fatalf(
			"expected reset window request to succeed, got %d",
			response.StatusCode,
		)
	}
}

func TestRateLimiterSupportsSkippedRoutes(
	t *testing.T,
) {
	handler, err := NewRateLimiter(
		RateLimiterConfig{
			MaxRequests: 1,
			Window:      time.Minute,
			KeyGenerator: func(
				*fiber.Ctx,
			) string {
				return "client-a"
			},
			Next: func(
				c *fiber.Ctx,
			) bool {
				return c.Path() == "/health"
			},
		},
	)
	if err != nil {
		t.Fatalf(
			"create rate limiter: %v",
			err,
		)
	}

	app := fiber.New()
	app.Use(
		handler,
	)
	app.Get(
		"/health",
		func(
			c *fiber.Ctx,
		) error {
			return c.SendStatus(
				fiber.StatusOK,
			)
		},
	)

	for attempt := 0; attempt < 3; attempt++ {
		response, requestErr := app.Test(
			httptest.NewRequest(
				fiber.MethodGet,
				"/health",
				nil,
			),
		)
		if requestErr != nil {
			t.Fatalf(
				"execute skipped request: %v",
				requestErr,
			)
		}

		if response.StatusCode != fiber.StatusOK {
			t.Fatalf(
				"expected skipped request to succeed, got %d",
				response.StatusCode,
			)
		}
	}
}

func TestNewRateLimiterValidatesConfiguration(
	t *testing.T,
) {
	_, err := NewRateLimiter(
		RateLimiterConfig{
			Window: time.Minute,
		},
	)
	if !errors.Is(
		err,
		ErrRateLimitMaximumInvalid,
	) {
		t.Fatalf(
			"expected maximum validation error, got %v",
			err,
		)
	}

	_, err = NewRateLimiter(
		RateLimiterConfig{
			MaxRequests: 1,
		},
	)
	if !errors.Is(
		err,
		ErrRateLimitWindowInvalid,
	) {
		t.Fatalf(
			"expected window validation error, got %v",
			err,
		)
	}
}
