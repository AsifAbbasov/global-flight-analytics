package server

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/http/handlers"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func New(
	cfg Config,
) (*fiber.App, error) {
	app := fiber.New()

	registerMiddleware(
		app,
		cfg,
	)

	v1 := app.Group(
		"/api",
	).Group(
		"/v1",
	)

	registerSystemRoutes(
		v1,
	)

	if cfg.DatabasePool != nil {
		if err := registerDatabaseRoutes(
			v1,
			cfg.DatabasePool,
			cfg.OpenMeteoTimeout,
		); err != nil {
			return nil, err
		}
	}

	return app, nil
}

func registerMiddleware(
	app *fiber.App,
	cfg Config,
) {
	app.Use(
		recover.New(),
	)

	app.Use(
		middleware.RequestID(),
	)

	app.Use(
		middleware.RequestLogger(
			cfg.Logger,
		),
	)

	app.Use(
		cors.New(
			cors.Config{
				AllowOrigins: "http://localhost:3000,http://localhost:3001",
				AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
				AllowHeaders: "Origin,Content-Type,Accept,Authorization",
			},
		),
	)
}

func registerSystemRoutes(
	v1 fiber.Router,
) {
	v1.Get(
		"/health",
		handlers.Health,
	)

	v1.Get(
		"/version",
		handlers.Version,
	)
}
