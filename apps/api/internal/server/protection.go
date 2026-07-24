package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/response"
	internalmiddleware "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

const (
	defaultAllowedOrigins  = "http://localhost:3000,http://localhost:3001"
	defaultBodyLimitBytes  = 1024 * 1024
	defaultReadTimeout     = 10 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultRateLimitMax    = 120
	defaultRateLimitWindow = time.Minute
)

func normalizeConfig(
	cfg Config,
) (Config, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	protection, err := normalizeProtectionConfig(
		cfg.Protection,
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"validate api protection configuration: %w",
			err,
		)
	}

	cfg.Protection = protection

	return cfg, nil
}

func normalizeProtectionConfig(
	config ProtectionConfig,
) (ProtectionConfig, error) {
	if strings.TrimSpace(
		config.AllowedOrigins,
	) == "" {
		config.AllowedOrigins = defaultAllowedOrigins
	}

	allowedOrigins, err := normalizeAllowedOrigins(
		config.AllowedOrigins,
	)
	if err != nil {
		return ProtectionConfig{}, err
	}
	config.AllowedOrigins = allowedOrigins

	if config.BodyLimitBytes == 0 {
		config.BodyLimitBytes = defaultBodyLimitBytes
	}
	if config.BodyLimitBytes < 0 {
		return ProtectionConfig{}, fmt.Errorf(
			"body limit must be greater than zero",
		)
	}

	if config.ReadTimeout == 0 {
		config.ReadTimeout = defaultReadTimeout
	}
	if config.ReadTimeout < 0 {
		return ProtectionConfig{}, fmt.Errorf(
			"read timeout must be greater than zero",
		)
	}

	if config.WriteTimeout == 0 {
		config.WriteTimeout = defaultWriteTimeout
	}
	if config.WriteTimeout < 0 {
		return ProtectionConfig{}, fmt.Errorf(
			"write timeout must be greater than zero",
		)
	}

	if config.IdleTimeout == 0 {
		config.IdleTimeout = defaultIdleTimeout
	}
	if config.IdleTimeout < 0 {
		return ProtectionConfig{}, fmt.Errorf(
			"idle timeout must be greater than zero",
		)
	}

	if config.RateLimitMax == 0 {
		config.RateLimitMax = defaultRateLimitMax
	}
	if config.RateLimitMax < 0 {
		return ProtectionConfig{}, fmt.Errorf(
			"rate limit maximum must be greater than zero",
		)
	}

	if config.RateLimitWindow == 0 {
		config.RateLimitWindow = defaultRateLimitWindow
	}
	if config.RateLimitWindow < 0 {
		return ProtectionConfig{}, fmt.Errorf(
			"rate limit window must be greater than zero",
		)
	}

	if config.MutationKeyConfigured &&
		config.MutationKeyDigest.IsZero() {
		return ProtectionConfig{}, fmt.Errorf(
			"configured mutation key digest must not be zero",
		)
	}

	return config, nil
}

func normalizeAllowedOrigins(
	value string,
) (string, error) {
	parts := strings.Split(
		value,
		",",
	)

	seen := make(
		map[string]struct{},
		len(parts),
	)

	normalized := make(
		[]string,
		0,
		len(parts),
	)

	for _, part := range parts {
		origin := strings.TrimSpace(
			part,
		)
		if origin == "" {
			continue
		}

		if origin == "*" {
			return "", fmt.Errorf(
				"wildcard origins are not allowed",
			)
		}

		parsed, err := url.Parse(
			origin,
		)
		if err != nil ||
			(parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.Host == "" ||
			parsed.User != nil ||
			parsed.Path != "" ||
			parsed.RawQuery != "" ||
			parsed.Fragment != "" {
			return "", fmt.Errorf(
				"invalid allowed origin %q",
				origin,
			)
		}

		if _, exists := seen[origin]; exists {
			continue
		}

		seen[origin] = struct{}{}
		normalized = append(
			normalized,
			origin,
		)
	}

	if len(normalized) == 0 {
		return "", fmt.Errorf(
			"at least one allowed origin is required",
		)
	}

	return strings.Join(
		normalized,
		",",
	), nil
}

func newFiberConfig(
	cfg Config,
) fiber.Config {
	return fiber.Config{
		BodyLimit:             cfg.Protection.BodyLimitBytes,
		ReadTimeout:           cfg.Protection.ReadTimeout,
		WriteTimeout:          cfg.Protection.WriteTimeout,
		IdleTimeout:           cfg.Protection.IdleTimeout,
		DisableStartupMessage: true,
		ErrorHandler:          newErrorHandler(cfg.Logger),
	}
}

func newErrorHandler(
	log *slog.Logger,
) fiber.ErrorHandler {
	return func(
		c *fiber.Ctx,
		err error,
	) error {
		status := fiber.StatusInternalServerError

		var fiberError *fiber.Error
		if errors.As(
			err,
			&fiberError,
		) {
			status = fiberError.Code
		}

		code, message := safeAPIError(
			status,
		)

		if status >= fiber.StatusInternalServerError {
			requestID, _ := c.Locals(
				internalmiddleware.RequestIDLocalKey,
			).(string)

			log.Error(
				"unhandled api request error",
				"request_id",
				requestID,
				"method",
				c.Method(),
				"path",
				c.Path(),
				"status",
				status,
				"error_type",
				fmt.Sprintf(
					"%T",
					err,
				),
			)
		}

		return response.Error(
			c,
			status,
			code,
			message,
		)
	}
}

func safeAPIError(
	status int,
) (
	string,
	string,
) {
	switch status {
	case fiber.StatusBadRequest:
		return "BAD_REQUEST", "Invalid request"
	case fiber.StatusNotFound:
		return "NOT_FOUND", "Resource not found"
	case fiber.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED", "Method not allowed"
	case fiber.StatusRequestTimeout:
		return "REQUEST_TIMEOUT", "Request timed out"
	case fiber.StatusRequestEntityTooLarge:
		return "PAYLOAD_TOO_LARGE", "Request payload is too large"
	case fiber.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED", "Too many requests"
	case fiber.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE", "Service is unavailable"
	default:
		if status >= fiber.StatusBadRequest &&
			status < fiber.StatusInternalServerError {
			return "REQUEST_REJECTED", "Request rejected"
		}

		return "INTERNAL_SERVER_ERROR", "Internal server error"
	}
}

func shouldSkipRateLimit(
	c *fiber.Ctx,
) bool {
	if c.Method() == fiber.MethodOptions {
		return true
	}

	switch c.Path() {
	case "/api/v1/health",
		"/api/v1/ready",
		"/api/v1/version":
		return true
	default:
		return false
	}
}

func rateLimitReached(
	c *fiber.Ctx,
) error {
	return response.Error(
		c,
		fiber.StatusTooManyRequests,
		"RATE_LIMIT_EXCEEDED",
		"Too many requests",
	)
}

// STAGE-14-5-MUTATION-ENDPOINT-PROTECTION
