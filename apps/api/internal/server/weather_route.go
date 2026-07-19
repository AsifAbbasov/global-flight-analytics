package server

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerWeatherRoute(
	v1 fiber.Router,
	dbPool *pgxpool.Pool,
	openMeteoTimeout time.Duration,
) error {
	if openMeteoTimeout <= 0 {
		return fmt.Errorf(
			"open-meteo timeout must be greater than zero",
		)
	}

	dependencies, err :=
		composeWeatherRouteDependencies(
			weatherCompositionConfig{
				databasePool:     dbPool,
				openMeteoTimeout: openMeteoTimeout,
			},
		)
	if err != nil {
		return err
	}

	return registerCurrentWeatherRoute(
		v1,
		dependencies.handler,
	)
}
