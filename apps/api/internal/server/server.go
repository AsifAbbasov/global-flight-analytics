package server

import (
	"fmt"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	internalmiddleware "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg Config) (*fiber.App, error) {
	normalizedConfig, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	app := fiber.New(newFiberConfig(normalizedConfig))
	if err := registerMiddleware(app, normalizedConfig); err != nil {
		return nil, err
	}
	v1 := app.Group("/api").Group("/v1")
	registerSystemRoutes(
		v1,
		normalizedConfig.DatabasePool,
	)
	if normalizedConfig.DatabasePool != nil {
		mutationAuthorization, err :=
			internalmiddleware.NewMutationAuthorization(
				internalmiddleware.MutationAuthorizationConfig{
					ExpectedDigest: normalizedConfig.Protection.MutationKeyDigest,
					Configured:     normalizedConfig.Protection.MutationKeyConfigured,
				},
			)
		if err != nil {
			return nil, fmt.Errorf(
				"create mutation authorization middleware: %w",
				err,
			)
		}

		if err := registerDatabaseRoutes(
			v1,
			normalizedConfig.DatabasePool,
			normalizedConfig.OpenMeteoTimeout,
			mutationAuthorization,
		); err != nil {
			return nil, err
		}
	}
	return app, nil
}
func registerMiddleware(app *fiber.App, cfg Config) error {
	app.Use(recover.New())
	app.Use(internalmiddleware.RequestID())
	app.Use(internalmiddleware.RequestLogger(cfg.Logger))
	app.Use(internalmiddleware.SecurityHeaders())
	app.Use(cors.New(cors.Config{AllowOrigins: cfg.Protection.AllowedOrigins, AllowMethods: "GET,HEAD,POST,OPTIONS", AllowHeaders: "Accept,Content-Type,X-Request-ID", ExposeHeaders: "X-Request-ID,X-RateLimit-Limit,X-RateLimit-Remaining,X-RateLimit-Reset"}))
	limiter, err := internalmiddleware.NewRateLimiter(internalmiddleware.RateLimiterConfig{MaxRequests: cfg.Protection.RateLimitMax, Window: cfg.Protection.RateLimitWindow, Next: shouldSkipRateLimit, LimitReached: rateLimitReached})
	if err != nil {
		return fmt.Errorf("create api rate limiter: %w", err)
	}
	app.Use(limiter)
	return nil
}
func registerSystemRoutes(
	v1 fiber.Router,
	databasePool *pgxpool.Pool,
) {
	var readinessProbe handlers.ReadinessProbe
	if databasePool != nil {
		readinessProbe = databasePool.Ping
	}

	v1.Get("/health", handlers.Health)
	v1.Get(
		"/ready",
		handlers.Readiness(
			readinessProbe,
		),
	)
	v1.Get("/version", handlers.Version)
}

// STAGE-14-5-MUTATION-ENDPOINT-PROTECTION
