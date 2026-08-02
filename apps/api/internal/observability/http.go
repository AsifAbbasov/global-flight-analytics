package observability

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/security/internalapikey"
	"github.com/gofiber/fiber/v2"
)

const prometheusContentType = "text/plain; version=0.0.4; charset=utf-8"

type AuthorizationConfig struct {
	ExpectedDigest internalapikey.Digest
	Configured     bool
}

func NewAuthorization(
	config AuthorizationConfig,
) (fiber.Handler, error) {
	if config.Configured && config.ExpectedDigest.IsZero() {
		return nil, fmt.Errorf(
			"configured metrics authorization digest must not be zero",
		)
	}

	return func(ctx *fiber.Ctx) error {
		ctx.Set(fiber.HeaderCacheControl, "no-store")
		if !config.Configured {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(
				fiber.Map{
					"success": false,
					"error": fiber.Map{
						"code":    "METRICS_AUTHENTICATION_UNAVAILABLE",
						"message": "Metrics authentication is not configured",
					},
				},
			)
		}

		candidate := ctx.Get(internalapikey.HeaderName)
		if !config.ExpectedDigest.MatchesCandidate(candidate) {
			return ctx.Status(fiber.StatusUnauthorized).JSON(
				fiber.Map{
					"success": false,
					"error": fiber.Map{
						"code":    "METRICS_AUTHENTICATION_REQUIRED",
						"message": "Valid internal metrics credentials are required",
					},
				},
			)
		}

		return ctx.Next()
	}, nil
}

func HTTPMiddleware(
	registry *Registry,
) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if registry == nil || ctx.Path() == MetricsPath {
			return ctx.Next()
		}

		registry.BeginHTTPRequest()
		startedAt := time.Now()
		err := ctx.Next()

		route := "unmatched"
		if currentRoute := ctx.Route(); currentRoute != nil {
			route = currentRoute.Path
		}
		registry.FinishHTTPRequest(
			ctx.Method(),
			route,
			ctx.Response().StatusCode(),
			time.Since(startedAt),
		)
		return err
	}
}

func FiberHandler(
	registry *Registry,
) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if registry == nil {
			return fiber.NewError(
				fiber.StatusServiceUnavailable,
				"metrics registry is unavailable",
			)
		}

		ctx.Set(fiber.HeaderContentType, prometheusContentType)
		ctx.Set(fiber.HeaderCacheControl, "no-store")
		return ctx.SendString(registry.Render(ctx.UserContext()))
	}
}

func StandaloneHandler(
	registry *Registry,
	authorization AuthorizationConfig,
) (http.Handler, error) {
	if registry == nil {
		return nil, fmt.Errorf("metrics registry is required")
	}
	if authorization.Configured && authorization.ExpectedDigest.IsZero() {
		return nil, fmt.Errorf(
			"configured metrics authorization digest must not be zero",
		)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(
		"/metrics",
		func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Cache-Control", "no-store")
			if !authorization.Configured {
				http.Error(
					writer,
					"metrics authentication is not configured",
					http.StatusServiceUnavailable,
				)
				return
			}

			candidate := strings.TrimSpace(
				request.Header.Get(internalapikey.HeaderName),
			)
			if !authorization.ExpectedDigest.MatchesCandidate(candidate) {
				http.Error(
					writer,
					"valid internal metrics credentials are required",
					http.StatusUnauthorized,
				)
				return
			}

			writer.Header().Set("Content-Type", prometheusContentType)
			_, _ = writer.Write([]byte(registry.Render(request.Context())))
		},
	)
	return mux, nil
}

func renderWithTimeout(
	registry *Registry,
	timeout time.Duration,
) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return registry.Render(ctx)
}
