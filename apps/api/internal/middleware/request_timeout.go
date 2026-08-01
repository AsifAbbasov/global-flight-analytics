package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
)

var ErrRequestTimeoutInvalid = errors.New(
	"request timeout must be greater than zero",
)

func NewRequestTimeout(
	timeout time.Duration,
) (
	fiber.Handler,
	error,
) {
	if timeout <= 0 {
		return nil,
			ErrRequestTimeoutInvalid
	}

	return func(
		c *fiber.Ctx,
	) error {
		parent := c.UserContext()
		if parent == nil {
			parent = context.Background()
		}

		requestContext, cancel := context.WithTimeout(
			parent,
			timeout,
		)
		defer cancel()

		c.SetUserContext(
			requestContext,
		)

		err := c.Next()

		switch {
		case errors.Is(
			requestContext.Err(),
			context.DeadlineExceeded,
		):
			return fiber.ErrGatewayTimeout

		case errors.Is(
			requestContext.Err(),
			context.Canceled,
		):
			return fiber.ErrRequestTimeout

		default:
			return err
		}
	}, nil
}
