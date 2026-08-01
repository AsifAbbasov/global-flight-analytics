package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestNewRequestTimeoutRejectsNonPositiveDuration(
	t *testing.T,
) {
	for _, timeout := range []time.Duration{
		0,
		-time.Second,
	} {
		handler, err := NewRequestTimeout(
			timeout,
		)
		if handler != nil {
			t.Fatal(
				"expected nil request timeout middleware",
			)
		}
		if !errors.Is(
			err,
			ErrRequestTimeoutInvalid,
		) {
			t.Fatalf(
				"expected invalid timeout error, got %v",
				err,
			)
		}
	}
}

func TestRequestTimeoutProvidesDeadlineToUserContext(
	t *testing.T,
) {
	const timeout = 2 * time.Second

	handler, err := NewRequestTimeout(
		timeout,
	)
	if err != nil {
		t.Fatalf(
			"create request timeout middleware: %v",
			err,
		)
	}

	var observedDeadline time.Time

	app := fiber.New()
	app.Use(
		handler,
	)
	app.Get(
		"/",
		func(
			c *fiber.Ctx,
		) error {
			var ok bool
			observedDeadline, ok = c.UserContext().Deadline()
			if !ok {
				return errors.New(
					"expected request deadline",
				)
			}
			return c.SendStatus(
				fiber.StatusNoContent,
			)
		},
	)

	startedAt := time.Now()
	response, err := app.Test(
		httptest.NewRequest(
			fiber.MethodGet,
			"/",
			nil,
		),
	)
	if err != nil {
		t.Fatalf(
			"execute request: %v",
			err,
		)
	}
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf(
			"expected status 204, got %d",
			response.StatusCode,
		)
	}

	minimumDeadline := startedAt.Add(
		timeout - 250*time.Millisecond,
	)
	maximumDeadline := startedAt.Add(
		timeout + 250*time.Millisecond,
	)
	if observedDeadline.Before(minimumDeadline) ||
		observedDeadline.After(maximumDeadline) {
		t.Fatalf(
			"unexpected request deadline %s outside [%s, %s]",
			observedDeadline,
			minimumDeadline,
			maximumDeadline,
		)
	}
}

func TestRequestTimeoutPreservesEarlierParentDeadline(
	t *testing.T,
) {
	parent, cancel := context.WithTimeout(
		context.Background(),
		500*time.Millisecond,
	)
	defer cancel()

	parentDeadline, ok := parent.Deadline()
	if !ok {
		t.Fatal(
			"expected parent deadline",
		)
	}

	handler, err := NewRequestTimeout(
		5 * time.Second,
	)
	if err != nil {
		t.Fatalf(
			"create request timeout middleware: %v",
			err,
		)
	}

	var observedDeadline time.Time

	app := fiber.New()
	app.Use(
		func(
			c *fiber.Ctx,
		) error {
			c.SetUserContext(
				parent,
			)
			return c.Next()
		},
	)
	app.Use(
		handler,
	)
	app.Get(
		"/",
		func(
			c *fiber.Ctx,
		) error {
			observedDeadline, _ = c.UserContext().Deadline()
			return c.SendStatus(
				fiber.StatusNoContent,
			)
		},
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
			"execute request: %v",
			err,
		)
	}
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf(
			"expected status 204, got %d",
			response.StatusCode,
		)
	}
	if !observedDeadline.Equal(
		parentDeadline,
	) {
		t.Fatalf(
			"expected parent deadline %s, got %s",
			parentDeadline,
			observedDeadline,
		)
	}
}

func TestRequestTimeoutReturnsGatewayTimeoutAfterDeadline(
	t *testing.T,
) {
	handler, err := NewRequestTimeout(
		10 * time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"create request timeout middleware: %v",
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
			<-c.UserContext().Done()
			return nil
		},
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
			"execute request: %v",
			err,
		)
	}
	if response.StatusCode != fiber.StatusGatewayTimeout {
		t.Fatalf(
			"expected status 504, got %d",
			response.StatusCode,
		)
	}
}

func TestRequestTimeoutOverridesHandlerErrorAfterDeadline(
	t *testing.T,
) {
	handler, err := NewRequestTimeout(
		10 * time.Millisecond,
	)
	if err != nil {
		t.Fatalf(
			"create request timeout middleware: %v",
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
			<-c.UserContext().Done()
			return errors.New(
				"repository returned after deadline",
			)
		},
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
			"execute request: %v",
			err,
		)
	}
	if response.StatusCode != fiber.StatusGatewayTimeout {
		t.Fatalf(
			"expected status 504, got %d",
			response.StatusCode,
		)
	}
}
