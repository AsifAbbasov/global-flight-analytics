package middleware

import (
	"fmt"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
	"github.com/gofiber/fiber/v2"
)

const (
	MutationAuthenticationRequiredCode    = "MUTATION_AUTHENTICATION_REQUIRED"
	MutationAuthenticationUnavailableCode = "MUTATION_AUTHENTICATION_UNAVAILABLE"
)

type MutationAuthorizationConfig struct {
	ExpectedDigest internalapikey.Digest
	Configured     bool
}

func NewMutationAuthorization(
	config MutationAuthorizationConfig,
) (
	fiber.Handler,
	error,
) {
	if config.Configured &&
		config.ExpectedDigest.IsZero() {
		return nil, fmt.Errorf(
			"configured mutation authorization digest must not be zero",
		)
	}

	return func(
		ctx *fiber.Ctx,
	) error {
		candidate := ctx.Get(
			internalapikey.HeaderName,
		)
		authorized :=
			config.ExpectedDigest.
				MatchesCandidate(
					candidate,
				)

		ctx.Set(
			fiber.HeaderCacheControl,
			"no-store",
		)

		if !config.Configured {
			return response.Error(
				ctx,
				fiber.StatusServiceUnavailable,
				MutationAuthenticationUnavailableCode,
				"Mutation authentication is not configured",
			)
		}
		if !authorized {
			return response.Error(
				ctx,
				fiber.StatusUnauthorized,
				MutationAuthenticationRequiredCode,
				"Valid internal mutation credentials are required",
			)
		}

		return ctx.Next()
	}, nil
}
