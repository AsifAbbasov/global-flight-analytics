package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"
const RequestIDLocalKey = "request_id"

const maximumRequestIDLength = 64

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := normalizeRequestID(
			c.Get(
				RequestIDHeader,
			),
		)

		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Locals(
			RequestIDLocalKey,
			requestID,
		)
		c.Set(
			RequestIDHeader,
			requestID,
		)

		return c.Next()
	}
}

func normalizeRequestID(
	value string,
) string {
	normalized := strings.TrimSpace(
		value,
	)

	if normalized == "" ||
		len(normalized) > maximumRequestIDLength {
		return ""
	}

	for _, character := range normalized {
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			character == '-' ||
			character == '_' ||
			character == '.' ||
			character == ':' {
			continue
		}

		return ""
	}

	return normalized
}
